package test

import (
	"context"
	"errors"
	"testing"

	"godoit.dev/doit/engine"
	"godoit.dev/doit/source"
	"godoit.dev/doit/spec"
	"godoit.dev/doit/target"
)

func TestCheck_RawErrorInOpCheck_PropagatesAndPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic for raw check error")
		}
	}()

	e := engine.New(source.LocalPosixSource{}, target.LocalPosixTarget{}, noopEmitter{})

	op := &fakeOp{
		name: "raw-error-op",
		checkFn: func(context.Context, source.Source, target.Target) (spec.CheckResult, error) {
			return spec.CheckUnsatisfied, errors.New("random check error")
		},
		execFn: panicExecFn("exec must not run on raw check error"),
	}

	plan := spec.Plan{
		Unit: spec.Unit{
			ID:   "fakeUnit",
			Desc: "fakeUnit description",
			Actions: []spec.Action{
				mkAction(op),
			},
		},
	}

	_, err := e.ExecutePlan(context.Background(), plan)

	// panicIfNotAbortError should trigger
	_ = err
}

func TestCheck_RawErrorInOpExec_PropagatesAndPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic for raw exec error")
		}
	}()

	e := engine.New(source.LocalPosixSource{}, target.LocalPosixTarget{}, noopEmitter{})

	op := &fakeOp{
		name: "raw-error-op",
		checkFn: func(context.Context, source.Source, target.Target) (spec.CheckResult, error) {
			return spec.CheckUnsatisfied, nil
		},
		execFn: func(context.Context, source.Source, target.Target) (spec.Result, error) {
			return spec.Result{}, errors.New("random exec error")
		},
	}

	plan := spec.Plan{
		Unit: spec.Unit{
			ID:   "fakeUnit",
			Desc: "fakeUnit description",
			Actions: []spec.Action{
				mkAction(op),
			},
		},
	}

	_, err := e.ExecutePlan(context.Background(), plan)

	// panicIfNotAbortError should trigger
	_ = err
}
