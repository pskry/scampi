package engine

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"godoit.dev/doit/diagnostic"
	"godoit.dev/doit/diagnostic/event"
	"godoit.dev/doit/source"
	"godoit.dev/doit/spec"
	"godoit.dev/doit/target"
)

func Apply(ctx context.Context, em diagnostic.Emitter, cfgPath string, store *spec.SourceStore) error {
	e := New(
		source.LocalPosixSource{},
		target.LocalPosixTarget{},
		em,
	)

	return e.Apply(ctx, cfgPath, store)
}

func (e *Engine) Apply(ctx context.Context, cfgPath string, store *spec.SourceStore) error {
	start := time.Now()
	e.em.Emit(diagnostic.EngineStarted())

	cfgPath, err := filepath.Abs(cfgPath)
	if err != nil {
		panic(fmt.Errorf("BUG: filepath.Abs() failed: %w", err))
	}

	cfg, err := LoadConfigWithSource(e.em, cfgPath, store, e.src)
	if err != nil {
		dr := emitDiagnostics(
			e.em,
			event.Subject{
				CfgPath: cfgPath,
			},
			err,
		)
		if dr.ShouldAbort() {
			return AbortError{Causes: []error{err}}
		}
		return err
	}

	plan, err := Plan(cfg, e.em)
	if err != nil {
		return err
	}

	// em.EngineFinish(changed bool, duration time.Duration)
	results, err := e.ExecutePlan(ctx, plan)
	if err != nil {
		// FIXME: diagnostic
		return err
	}

	rs := diagnostic.RunSummary{
		ChangedCount: 0,
		FailedCount:  0,
		TotalCount:   len(results),
	}
	for _, res := range results {
		if res.res.Changed {
			rs.ChangedCount++
		}
		if res.err != nil {
			rs.FailedCount++
		}
	}

	e.em.Emit(diagnostic.EngineFinished(rs, time.Since(start), err))

	return err
}
