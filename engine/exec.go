package engine

import (
	"cmp"
	"context"
	"maps"
	"slices"
	"sync"
	"time"

	"godoit.dev/doit/diagnostic"
	"godoit.dev/doit/diagnostic/event"
	"godoit.dev/doit/model"
	"godoit.dev/doit/source"
	"godoit.dev/doit/spec"
	"godoit.dev/doit/target"
	"godoit.dev/doit/util"
	"golang.org/x/sync/errgroup"
)

const actionTimeout = 5 * time.Second

const opOutcomeUnknown = model.OpOutcome(0xff)

func validateOpReport(r model.OpReport) {
	switch r.Outcome {
	case model.OpSucceeded:
		if r.Result == nil || r.Err != nil {
			panic(util.BUG("succeeded op must have result only, had error: %w", r.Err))
		}
	case model.OpFailed, model.OpAborted:
		if r.Err == nil || r.Result != nil {
			panic(util.BUG("failed/aborted op must have err only, had result: %+v", r.Result))
		}
	case model.OpSkipped:
		if r.Result != nil || r.Err != nil {
			panic(util.BUG("skipped op must have no result or err"))
		}
	default:
		panic(util.BUG("unknown op outcome"))
	}
}

type opNode struct {
	op         spec.Op
	deps       []*opNode
	dependents []*opNode

	indegree int // original dependency count
	pending  int // runtime counter

	satisfied bool

	outcome model.OpOutcome
	result  *spec.Result
	err     error
}

type scheduler struct {
	src source.Source
	tgt target.Target
	em  diagnostic.Emitter

	// action context for building event.Subject
	actIdx  int
	actKind string
	actDesc string

	mu      sync.Mutex
	results []spec.Result
	grp     *errgroup.Group
	ctx     context.Context
}

func (s *scheduler) opSubject(op spec.Op) event.OpSubject {
	return event.OpSubject{
		StepIndex: s.actIdx,
		StepKind:  s.actKind,
		StepDesc:  s.actDesc,
		DisplayID: diagnostic.OpDisplayID(op),
	}
}

func (s *scheduler) schedule(n *opNode) {
	if n.satisfied {
		return
	}

	s.grp.Go(func() error {
		start := time.Now()
		subj := s.opSubject(n.op)

		s.em.EmitOpLifecycle(diagnostic.OpExecuteStarted(subj))

		res, err := n.op.Execute(s.ctx, s.src, s.tgt)

		s.em.EmitOpLifecycle(diagnostic.OpExecuted(subj, res.Changed, time.Since(start), err))

		s.mu.Lock()
		defer s.mu.Unlock()

		if err != nil {
			n.outcome = model.OpFailed
			n.err = err
			n.result = nil
			return err
		}

		n.outcome = model.OpSucceeded
		n.err = nil
		n.result = &res
		s.results = append(s.results, res)

		// unblock unsatisfied dependents
		for _, d := range n.dependents {
			if d.satisfied {
				continue
			}

			d.pending--
			if d.pending == 0 {
				s.schedule(d)
			}
		}

		return nil
	})
}

func (s *scheduler) runChecks(nodes []*opNode) error {
	g, ctx := errgroup.WithContext(s.ctx)

	for _, n := range nodes {
		n := n
		g.Go(func() error {
			subj := s.opSubject(n.op)
			s.em.EmitOpLifecycle(diagnostic.OpCheckStarted(subj))

			res, err := n.op.Check(ctx, s.src, s.tgt)
			if err != nil {
				dr, consumed := emitDiagnostics(s.em, subj, err)

				s.em.EmitOpLifecycle(diagnostic.OpChecked(subj, res, err))
				if dr.ShouldAbort() {
					s.mu.Lock()
					n.outcome = model.OpAborted
					n.err = err
					s.mu.Unlock()
					return AbortError{Causes: []error{err}}
				}

				if consumed {
					return nil
				}

				return err
			}

			s.em.EmitOpLifecycle(diagnostic.OpChecked(subj, res, nil))

			s.mu.Lock()
			if res == spec.CheckSatisfied {
				n.satisfied = true
				n.outcome = model.OpSkipped
			} else {
				n.satisfied = false
			}
			s.mu.Unlock()
			return nil
		})
	}

	return g.Wait()
}

func (s *scheduler) initPending(nodes []*opNode) {
	for _, n := range nodes {
		n.pending = 0
	}

	for _, n := range nodes {
		if n.satisfied {
			continue
		}

		for _, d := range n.dependents {
			if !d.satisfied {
				d.pending++
			}
		}
	}
}

func (e *Engine) ExecutePlan(ctx context.Context, plan spec.Plan) (model.ExecutionReport, error) {
	res, err := e.executePlan(ctx, plan)
	if err != nil {
		return res, panicIfNotAbortError(err)
	}
	return res, nil
}

func (e *Engine) executePlan(ctx context.Context, plan spec.Plan) (model.ExecutionReport, error) {
	var rep model.ExecutionReport

	for i, act := range plan.Unit.Actions {
		res, err := e.executeAction(ctx, i, act)
		rep.Actions = append(rep.Actions, res)
		if err != nil {
			rep.Err = err
			return rep, err
		}

	}

	return rep, nil
}

func (e *Engine) executeAction(ctx context.Context, idx int, act spec.Action) (model.ActionReport, error) {
	start := time.Now()
	kind := act.Kind()
	desc := act.Desc()
	e.em.EmitActionLifecycle(diagnostic.ActionStarted(idx, kind, desc))

	actCtx, cancel := context.WithTimeout(ctx, actionTimeout)
	defer cancel()

	res, err := e.runAction(actCtx, idx, act)

	e.em.EmitActionLifecycle(
		diagnostic.ActionFinished(
			idx,
			kind,
			desc,
			res.Summary,
			time.Since(start),
			err,
		),
	)

	if err != nil {
		return res, err
	}

	return res, nil
}

func (e *Engine) runAction(ctx context.Context, idx int, act spec.Action) (model.ActionReport, error) {
	nodes, planErr := buildPlan(act.Ops())
	if planErr != nil {
		return model.ActionReport{}, planErr
	}

	s := &scheduler{
		src:     e.src,
		tgt:     e.tgt,
		em:      e.em,
		actIdx:  idx,
		actKind: act.Kind(),
		actDesc: act.Desc(),
	}
	s.grp, s.ctx = errgroup.WithContext(ctx)

	checkErr := s.runChecks(nodes)

	var execErr error
	if checkErr == nil {
		s.initPending(nodes)

		// Hold lock while reading node state to avoid race with goroutines
		// decrementing pending counts
		s.mu.Lock()
		for _, n := range nodes {
			if !n.satisfied && n.pending == 0 {
				s.schedule(n)
			}
		}
		s.mu.Unlock()

		execErr = s.grp.Wait()
	}

	// First error wins
	err := cmp.Or(checkErr, execErr)
	// Mark any ops without outcome as aborted
	if err != nil {
		for _, n := range nodes {
			if n.outcome == opOutcomeUnknown {
				n.outcome = model.OpAborted
				n.err = err
			}
		}
	}
	// Enforce invariant: every op MUST have an outcome
	for _, n := range nodes {
		if n.outcome == opOutcomeUnknown {
			panic(util.BUG("op left without outcome"))
		}
	}

	if err != nil {
		dr, consumed := emitDiagnostics(
			e.em,
			event.ActionSubject{
				StepIndex: idx,
				StepKind:  act.Kind(),
				StepDesc:  act.Desc(),
			},
			err,
		)
		if dr.ShouldAbort() {
			err = AbortError{Causes: []error{err}}
		} else if consumed {
			err = nil
		}
	}

	// Build ActionReport
	var rep model.ActionReport
	rep.Action = act

	for _, n := range nodes {
		or := model.OpReport{
			Op:      n.op,
			Outcome: n.outcome,
			Result:  n.result,
			Err:     n.err,
		}
		validateOpReport(or)
		rep.Ops = append(rep.Ops, or)

		rep.Summary.Total++

		switch n.outcome {
		case model.OpSucceeded:
			rep.Summary.Succeeded++
			if n.result != nil && n.result.Changed {
				rep.Summary.Changed++
			}
		case model.OpFailed:
			rep.Summary.Failed++
		case model.OpAborted:
			rep.Summary.Aborted++
		case model.OpSkipped:
			rep.Summary.Skipped++
		}
	}

	return rep, err
}

func buildPlan(ops []spec.Op) ([]*opNode, error) {
	nodes := map[spec.Op]*opNode{}

	for _, op := range ops {
		nodes[op] = &opNode{
			op: op,
			// explicit invariants
			outcome: opOutcomeUnknown,
			result:  nil,
			err:     nil,
		}
	}

	for _, n := range nodes {
		for _, dep := range n.op.DependsOn() {
			dn, ok := nodes[dep]
			if !ok {
				panic(util.BUG(
					"op %p depends on unknown op %p (StepType implementation error)",
					n.op, dep,
				))
			}

			n.deps = append(n.deps, dn)
			dn.dependents = append(dn.dependents, n)
			n.indegree++
		}
	}

	tmp := make(map[*opNode]int)
	for _, n := range nodes {
		tmp[n] = n.indegree
	}

	var queue []*opNode
	for n, deg := range tmp {
		if deg == 0 {
			queue = append(queue, n)
		}
	}

	visited := 0
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		visited++

		for _, d := range n.dependents {
			tmp[d]--
			if tmp[d] == 0 {
				queue = append(queue, d)
			}
		}
	}

	if visited != len(nodes) {
		panic(util.BUG("cycle detected in op graph (StepType implementation error)"))
	}

	for _, n := range nodes {
		n.pending = n.indegree
	}

	return slices.Collect(maps.Values(nodes)), nil
}
