package engine

import (
	"context"
	"time"

	"godoit.dev/doit/diagnostic"
	"godoit.dev/doit/source"
	"godoit.dev/doit/spec"
)

func Apply(
	ctx context.Context,
	em diagnostic.Emitter,
	cfgPath string,
	store *spec.SourceStore,
	opts spec.ResolveOptions,
) error {
	src := source.WithRoot(cfgPath, source.LocalPosixSource{})
	cfg, err := LoadConfigWithOptions(ctx, em, cfgPath, store, src, opts)
	if err != nil {
		return err
	}

	resolved, err := ResolveMultiple(cfg, opts)
	if err != nil {
		if impact, ok := emitEngineDiagnostic(em, cfgPath, err); ok {
			if impact.ShouldAbort() {
				return AbortError{Causes: []error{err}}
			}
		}
		return err
	}

	for _, res := range resolved {
		e, err := New(ctx, src, res, em)
		if err != nil {
			return err
		}

		if err := e.Apply(ctx); err != nil {
			e.Close()
			return err
		}
		e.Close()
	}

	return nil
}

func (e *Engine) Apply(ctx context.Context) error {
	start := time.Now()
	e.em.EmitEngineLifecycle(diagnostic.EngineStarted())

	plan, _, err := plan(e.cfg, e.em, e.tgt.Capabilities())
	if err != nil {
		return err
	}

	rep, err := e.ExecutePlan(ctx, plan)
	if err != nil {
		return err
	}

	e.em.EmitEngineLifecycle(diagnostic.EngineFinished(rep, time.Since(start), err, false))

	return nil
}
