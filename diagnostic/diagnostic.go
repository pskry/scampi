//go:generate stringer -type=Impact
package diagnostic

import (
	"reflect"
	"time"

	"godoit.dev/doit/diagnostic/event"
	"godoit.dev/doit/model"
	"godoit.dev/doit/signal"
	"godoit.dev/doit/spec"
)

// OpDisplayID derives a display identifier for an op.
// Uses OpDescriber template ID if available, otherwise falls back to type name.
func OpDisplayID(op spec.Op) string {
	if d, ok := op.(spec.OpDescriber); ok {
		if desc := d.OpDescription(); desc != nil {
			if id := desc.PlanTemplate().ID; id != "" {
				return id
			}
		}
	}
	// Fallback: use the struct type name
	t := reflect.TypeOf(op)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t.Name()
}

type Impact uint8

const (
	ImpactAbort Impact = 1 << iota
	ImpactNone  Impact = 0
)

type (
	Emitter interface {
		EmitEngineLifecycle(e event.EngineEvent)
		EmitPlanLifecycle(e event.PlanEvent)
		EmitActionLifecycle(e event.ActionEvent)
		EmitOpLifecycle(e event.OpEvent)

		EmitDiagnostic(e event.Event)
	}
	Diagnostic interface {
		EventTemplate() event.Template
		Severity() signal.Severity
		Impact() Impact
	}
	DiagnosticProvider interface {
		Diagnostics(subject event.Subject) []event.Event
	}
)

// Engine lifecycle
// ===============================================

func EngineStarted() event.EngineEvent {
	return event.EngineEvent{
		Time:       time.Now(),
		Kind:       event.EngineStarted,
		Severity:   signal.Info,
		Chattiness: event.Subtle,
	}
}

func EngineFinished(rep model.ExecutionReport, dur time.Duration, err error) event.EngineEvent {
	e := event.EngineEvent{
		Time: time.Now(),
		Kind: event.EngineFinished,
		Detail: &event.EngineFinishedDetail{
			Duration: dur,
			Err:      err,
		},
	}

	for _, ar := range rep.Actions {
		e.Detail.TotalCount += ar.Summary.Total
		e.Detail.ChangedCount += ar.Summary.Changed
		e.Detail.FailedCount += ar.Summary.Failed
	}

	switch {
	case err != nil:
		e.Severity = signal.Fatal
		e.Chattiness = event.Normal

	case e.Detail.FailedCount > 0:
		e.Severity = signal.Error
		e.Chattiness = event.Normal

	case e.Detail.FailedCount > 0:
		e.Severity = signal.Notice
		e.Chattiness = event.Subtle

	default:
		e.Severity = signal.Info
		e.Chattiness = event.Subtle
	}

	return e
}

// Plan lifecycle
// ===============================================

func PlanStarted(unitID spec.UnitID) event.PlanEvent {
	return event.PlanEvent{
		Time: time.Now(),
		Kind: event.PlanStarted,
		StartedDetail: &event.PlanStartedDetail{
			UnitID: string(unitID),
		},
		Severity:   signal.Info,
		Chattiness: event.Subtle,
	}
}

func PlanFinished(unitID spec.UnitID, successfulSteps, failedSteps int, dur time.Duration) event.PlanEvent {
	e := event.PlanEvent{
		Time: time.Now(),
		Kind: event.PlanFinished,
		FinishedDetail: &event.PlanFinishedDetail{
			UnitID:          string(unitID),
			SuccessfulSteps: successfulSteps,
			FailedSteps:     failedSteps,
			Duration:        dur,
		},
	}

	switch {
	case failedSteps > 0:
		e.Severity = signal.Error
		e.Chattiness = event.Reserved

	case successfulSteps == 0:
		e.Severity = signal.Warning
		e.Chattiness = event.Normal

	default:
		e.Severity = signal.Info
		e.Chattiness = event.Subtle
	}

	return e
}

func StepPlanned(
	index int,
	desc string,
	kind string,
) event.PlanEvent {
	return event.PlanEvent{
		Time: time.Now(),
		Kind: event.StepPlanned,
		Subject: event.PlanSubject{
			StepIndex: index,
			StepDesc:  desc,
			StepKind:  kind,
		},
		Severity:   signal.Debug,
		Chattiness: event.Chatty,
	}
}

func PlanProduced(plan spec.Plan) event.PlanEvent {
	// ------------------------------------------------------------
	// 1. Flatten all ops and assign GLOBAL indices
	// ------------------------------------------------------------
	var allOps []spec.Op
	opIndex := make(map[spec.Op]int)
	actionOpBase := make(map[int]int) // action index → first op index
	for i, act := range plan.Unit.Actions {
		actionOpBase[i] = len(allOps)
		for _, op := range act.Ops() {
			opIndex[op] = len(allOps)
			allOps = append(allOps, op)
		}
	}

	// ------------------------------------------------------------
	// 2. Build PlannedOps with dependency indices
	// ------------------------------------------------------------
	plannedOps := make([]event.PlannedOp, len(allOps))
	for i, op := range allOps {
		var tmpl *spec.PlanTemplate

		if d, ok := op.(spec.OpDescriber); ok {
			if desc := d.OpDescription(); desc != nil {
				t := desc.PlanTemplate()
				tmpl = &t
			}
		}

		var deps []int
		for _, dep := range op.DependsOn() {
			deps = append(deps, opIndex[dep])
		}

		plannedOps[i] = event.PlannedOp{
			Index:     i,
			DisplayID: OpDisplayID(op),
			DependsOn: deps,
			Template:  tmpl, // nil = fallback
		}
	}

	// ------------------------------------------------------------
	// 3. Re-slice ops back into PlannedActions
	// ------------------------------------------------------------
	detail := event.PlanDetail{
		UnitID:   string(plan.Unit.ID),
		UnitDesc: plan.Unit.Desc,
	}
	for i, act := range plan.Unit.Actions {
		start := actionOpBase[i]
		end := start + len(act.Ops())

		detail.Actions = append(detail.Actions, event.PlannedAction{
			Index: i,
			Desc:  act.Desc(),
			Kind:  act.Kind(),
			Ops:   plannedOps[start:end],
		})
	}

	return event.PlanEvent{
		Time:       time.Now(),
		Kind:       event.PlanProduced,
		Detail:     &detail,
		Severity:   signal.Notice,
		Chattiness: event.Subtle,
	}
}

// Action lifecycle
// ===============================================

func ActionStarted(idx int, kind, desc string) event.ActionEvent {
	return event.ActionEvent{
		Time: time.Now(),
		Kind: event.ActionStarted,
		Subject: event.ActionSubject{
			StepIndex: idx,
			StepKind:  kind,
			StepDesc:  desc,
		},
		Severity:   signal.Notice,
		Chattiness: event.Normal,
	}
}

func ActionFinished(
	idx int,
	kind,
	desc string,
	summary model.ActionSummary,
	dur time.Duration,
	err error,
) event.ActionEvent {
	e := event.ActionEvent{
		Time: time.Now(),
		Kind: event.ActionFinished,
		Subject: event.ActionSubject{
			StepIndex: idx,
			StepKind:  kind,
			StepDesc:  desc,
		},
		Detail: &event.ActionDetail{
			Summary:  summary,
			Duration: dur,
			Err:      err,
		},
	}

	s := summary
	switch {

	case s.Failed > 0 || s.Aborted > 0 || err != nil:
		e.Severity = signal.Error
		e.Chattiness = event.Normal

	case s.Changed > 0:
		e.Severity = signal.Notice
		e.Chattiness = event.Normal

	default:
		e.Severity = signal.Info
		e.Chattiness = event.Reserved
	}

	return e
}

// Op lifecycle
// ===============================================

func OpCheckStarted(subject event.Subject) event.OpEvent {
	return event.OpEvent{
		Time:       time.Now(),
		Kind:       event.OpCheckStarted,
		Subject:    subject,
		Severity:   signal.Debug,
		Chattiness: event.Chatty,
	}
}

func OpChecked(subject event.Subject, res spec.CheckResult, err error) event.OpEvent {
	e := event.OpEvent{
		Time:    time.Now(),
		Kind:    event.OpChecked,
		Subject: subject,
		CheckDetail: &event.OpCheckDetail{
			Result: res,
			Err:    err,
		},
	}

	switch res {
	case spec.CheckSatisfied:
		e.Severity = signal.Info
		e.Chattiness = event.Subtle

	case spec.CheckUnsatisfied:
		e.Severity = signal.Notice
		e.Chattiness = event.Normal

	case spec.CheckUnknown:
		e.Severity = signal.Warning
		e.Chattiness = event.Reserved
	}

	return e
}

func OpExecuteStarted(subject event.Subject) event.OpEvent {
	return event.OpEvent{
		Time:       time.Now(),
		Kind:       event.OpExecuteStarted,
		Subject:    subject,
		Severity:   signal.Info,
		Chattiness: event.Chatty,
	}
}

func OpExecuted(subject event.Subject, changed bool, dur time.Duration, err error) event.OpEvent {
	e := event.OpEvent{
		Time:    time.Now(),
		Kind:    event.OpExecuted,
		Subject: subject,
		ExecuteDetail: &event.OpExecuteDetail{
			Changed:  changed,
			Duration: dur,
			Err:      err,
		},
	}

	switch {
	case err != nil:
		e.Severity = signal.Error
		e.Chattiness = event.Normal

	case changed:
		e.Severity = signal.Notice
		e.Chattiness = event.Normal

	default:
		e.Severity = signal.Info
		e.Chattiness = event.Reserved
	}

	return e
}

// Diagnostics
// ===============================================

func DiagnosticRaised(subject event.Subject, d Diagnostic) event.Event {
	return event.Event{
		Time:    time.Now(),
		Kind:    event.DiagnosticRaised,
		Scope:   subject.Scope(),
		Subject: subject,
		Detail: event.DiagnosticDetail{
			Template: d.EventTemplate(),
		},
		Severity:   d.Severity(),
		Chattiness: event.Subtle,
	}
}
