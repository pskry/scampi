// SPDX-License-Identifier: GPL-3.0-only

package engine

import (
	"context"
	"fmt"
	"strings"
	"time"

	"scampi.dev/scampi/diagnostic"
	"scampi.dev/scampi/diagnostic/event"
	"scampi.dev/scampi/model"
	"scampi.dev/scampi/spec"
)

func Check(
	ctx context.Context,
	em diagnostic.Emitter,
	cfgPath string,
	store *diagnostic.SourceStore,
	opts spec.ResolveOptions,
) error {
	return forEachResolved(ctx, em, cfgPath, store, opts, func(ctx context.Context, e *Engine) error {
		return e.converge(ctx, true)
	})
}

func Apply(
	ctx context.Context,
	em diagnostic.Emitter,
	cfgPath string,
	store *diagnostic.SourceStore,
	opts spec.ResolveOptions,
) error {
	return forEachResolved(ctx, em, cfgPath, store, opts, func(ctx context.Context, e *Engine) error {
		return e.converge(ctx, false)
	})
}

func (e *Engine) Check(ctx context.Context) error { return e.converge(ctx, true) }
func (e *Engine) Apply(ctx context.Context) error { return e.converge(ctx, false) }

func (e *Engine) converge(ctx context.Context, checkOnly bool) error {
	start := time.Now()
	p, _, hp, err := plan(e.cfg, e.em, e.tgt.Capabilities())
	if err != nil {
		return err
	}
	e.storeSourcePaths(ctx, p)

	var rep model.ExecutionReport
	var promisedPaths map[spec.Resource]bool
	if checkOnly {
		rep, promisedPaths, err = e.CheckPlan(ctx, p)
	} else {
		rep, err = e.ExecutePlan(ctx, p)
	}
	if err != nil {
		return err
	}

	hookRep, err := e.executeHooks(ctx, rep, hp, checkOnly, promisedPaths)
	if err != nil {
		return err
	}
	rep.Actions = append(rep.Actions, hookRep.Actions...)

	e.em.EmitProgress(event.Progress{
		Time: time.Now(),
		Text: summarizeRun(rep, checkOnly, time.Since(start)),
	})

	return nil
}

// summarizeRun produces the final one-line summary the CLI shows at end
// of run. Computed from the ExecutionReport per the diagnostics design.
func summarizeRun(rep model.ExecutionReport, checkOnly bool, dur time.Duration) string {
	var changed, wouldChange, failed int
	for _, ar := range rep.Actions {
		changed += ar.Summary.Changed
		wouldChange += ar.Summary.WouldChange
		failed += ar.Summary.Failed + ar.Summary.Aborted
	}
	var parts []string
	if checkOnly {
		parts = append(parts, fmt.Sprintf("%d would change", wouldChange))
	} else {
		parts = append(parts, fmt.Sprintf("%d changed", changed))
	}
	if failed > 0 {
		parts = append(parts, fmt.Sprintf("%d failed", failed))
	}
	parts = append(parts, dur.Round(time.Millisecond).String())
	return "done: " + strings.Join(parts, ", ")
}
