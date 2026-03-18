// SPDX-License-Identifier: GPL-3.0-only

package test

import (
	"context"
	"testing"

	"scampi.dev/scampi/diagnostic"
	"scampi.dev/scampi/engine"
	"scampi.dev/scampi/model"
	"scampi.dev/scampi/signal"
	"scampi.dev/scampi/source"
	"scampi.dev/scampi/spec"
	"scampi.dev/scampi/step/copy"
	"scampi.dev/scampi/step/sharedops"
	"scampi.dev/scampi/target"
	"scampi.dev/scampi/target/local"
)

// fakePromiserAction wraps a fakeAction with resource declarations for the
// action dependency graph and promise system.
type fakePromiserAction struct {
	fakeAction
	inputs   []spec.Resource
	promises []spec.Resource
}

func (a *fakePromiserAction) Inputs() []spec.Resource   { return a.inputs }
func (a *fakePromiserAction) Promises() []spec.Resource { return a.promises }

func paths(s ...string) []spec.Resource {
	r := make([]spec.Resource, len(s))
	for i, p := range s {
		r[i] = spec.PathResource(p)
	}
	return r
}

func users(s ...string) []spec.Resource {
	r := make([]spec.Resource, len(s))
	for i, u := range s {
		r[i] = spec.UserResource(u)
	}
	return r
}

func groups(s ...string) []spec.Resource {
	r := make([]spec.Resource, len(s))
	for i, g := range s {
		r[i] = spec.GroupResource(g)
	}
	return r
}

func mkPromiserAction(inputs, promises []spec.Resource, ops ...*fakeOp) *fakePromiserAction {
	act := &fakePromiserAction{
		inputs:   inputs,
		promises: promises,
	}
	for _, op := range ops {
		act.ops = append(act.ops, op)
		op.action = act
	}
	return act
}

// TestCheck_DeferredPath_UpstreamPromisesDirectory verifies that check mode
// does not abort when a downstream op reports a missing directory that an
// upstream action has promised to create.
func TestCheck_DeferredPath_UpstreamPromisesDirectory(t *testing.T) {
	// dir action: check says "unsatisfied" (directory doesn't exist yet)
	dirOp := &fakeOp{
		name:    "ensure-dir",
		checkFn: okCheckFn(spec.CheckUnsatisfied),
		execFn:  panicExecFn("check mode must not execute"),
	}
	dirAction := mkPromiserAction(nil, paths("/foo"), dirOp)

	// copy action: check returns CopyDestDirMissingError for /foo
	copyOp := &fakeOp{
		name: "copy-file",
		checkFn: func(context.Context, source.Source, target.Target) (spec.CheckResult, []spec.DriftDetail, error) {
			return spec.CheckUnsatisfied, nil, copy.CopyDestDirMissingError{
				Path: "/foo",
			}
		},
		execFn: panicExecFn("check mode must not execute"),
	}
	copyAction := mkPromiserAction(paths("/foo"), paths("/foo/bar"), copyOp)

	plan := spec.Plan{
		Unit: spec.Unit{
			ID:      "test-deferred",
			Actions: []spec.Action{dirAction, copyAction},
		},
	}

	cfg := spec.ResolvedConfig{
		Target: mockTargetInstance(local.POSIXTarget{}),
	}

	e, err := engine.New(context.Background(), source.LocalPosixSource{}, cfg, noopEmitter{})
	if err != nil {
		t.Fatalf("engine.New: %v", err)
	}
	defer e.Close()

	rep, _, err := e.CheckPlan(context.Background(), plan)
	if err != nil {
		t.Fatalf("CheckPlan must not return error when path is deferred, got: %v", err)
	}

	if len(rep.Actions) != 2 {
		t.Fatalf("expected 2 action reports, got %d", len(rep.Actions))
	}

	for i, ar := range rep.Actions {
		if ar.Summary.WouldChange == 0 {
			t.Errorf("action %d: expected WouldChange > 0, got %+v", i, ar.Summary)
		}
		if ar.Summary.Failed > 0 {
			t.Errorf("action %d: expected no failures, got Failed=%d", i, ar.Summary.Failed)
		}
		if ar.Summary.Aborted > 0 {
			t.Errorf("action %d: expected no aborts, got Aborted=%d", i, ar.Summary.Aborted)
		}
	}
}

// TestCheck_DeferredPath_NoPromise_StillAborts verifies that a missing
// directory error still aborts when no upstream action promises the path.
func TestCheck_DeferredPath_NoPromise_StillAborts(t *testing.T) {
	copyOp := &fakeOp{
		name: "copy-file",
		checkFn: func(context.Context, source.Source, target.Target) (spec.CheckResult, []spec.DriftDetail, error) {
			return spec.CheckUnsatisfied, nil, copy.CopyDestDirMissingError{
				Path: "/nonexistent",
			}
		},
		execFn: panicExecFn("check mode must not execute"),
	}
	copyAction := mkPromiserAction(nil, paths("/nonexistent/bar"), copyOp)

	plan := spec.Plan{
		Unit: spec.Unit{
			ID:      "test-no-promise",
			Actions: []spec.Action{copyAction},
		},
	}

	cfg := spec.ResolvedConfig{
		Target: mockTargetInstance(local.POSIXTarget{}),
	}

	e, err := engine.New(context.Background(), source.LocalPosixSource{}, cfg, noopEmitter{})
	if err != nil {
		t.Fatalf("engine.New: %v", err)
	}
	defer e.Close()

	_, _, err = e.CheckPlan(context.Background(), plan)
	if err == nil {
		t.Fatalf("CheckPlan must abort when no upstream action promises the path")
	}
}

// TestCheck_DeferredPath_UpstreamSatisfied_NoPromise verifies that a
// satisfied upstream action (CheckSatisfied, WouldChange=0) does NOT add
// its paths to the promised set, so a downstream missing-dir error still aborts.
func TestCheck_DeferredPath_UpstreamSatisfied_NoPromise(t *testing.T) {
	// dir action: already satisfied (directory exists)
	dirOp := &fakeOp{
		name:    "ensure-dir",
		checkFn: okCheckFn(spec.CheckSatisfied),
		execFn:  panicExecFn("check mode must not execute"),
	}
	dirAction := mkPromiserAction(nil, paths("/foo"), dirOp)

	// copy action: missing dir error
	copyOp := &fakeOp{
		name: "copy-file",
		checkFn: func(context.Context, source.Source, target.Target) (spec.CheckResult, []spec.DriftDetail, error) {
			return spec.CheckUnsatisfied, nil, copy.CopyDestDirMissingError{
				Path: "/foo",
			}
		},
		execFn: panicExecFn("check mode must not execute"),
	}
	copyAction := mkPromiserAction(paths("/foo"), paths("/foo/bar"), copyOp)

	plan := spec.Plan{
		Unit: spec.Unit{
			ID:      "test-satisfied-no-promise",
			Actions: []spec.Action{dirAction, copyAction},
		},
	}

	cfg := spec.ResolvedConfig{
		Target: mockTargetInstance(local.POSIXTarget{}),
	}

	e, err := engine.New(context.Background(), source.LocalPosixSource{}, cfg, noopEmitter{})
	if err != nil {
		t.Fatalf("engine.New: %v", err)
	}
	defer e.Close()

	_, _, err = e.CheckPlan(context.Background(), plan)
	if err == nil {
		t.Fatalf("CheckPlan must abort: upstream is satisfied so path is not promised")
	}
}

// TestCheck_DeferredPath_NonDeferrableError_StillAborts verifies that abort
// errors that don't implement Deferrable are not deferred even when a
// matching promised path exists.
func TestCheck_DeferredPath_NonDeferrableError_StillAborts(t *testing.T) {
	dirOp := &fakeOp{
		name:    "ensure-dir",
		checkFn: okCheckFn(spec.CheckUnsatisfied),
		execFn:  panicExecFn("check mode must not execute"),
	}
	dirAction := mkPromiserAction(nil, paths("/foo"), dirOp)

	// This op returns a plain abort diagnostic (not Deferrable)
	abortOp := &fakeOp{
		name:    "abort-op",
		checkFn: diagCheckFn(signal.Error, diagnostic.ImpactAbort),
		execFn:  panicExecFn("check mode must not execute"),
	}
	abortAction := mkPromiserAction(paths("/foo"), paths("/foo/file"), abortOp)

	plan := spec.Plan{
		Unit: spec.Unit{
			ID:      "test-non-deferrable",
			Actions: []spec.Action{dirAction, abortAction},
		},
	}

	cfg := spec.ResolvedConfig{
		Target: mockTargetInstance(local.POSIXTarget{}),
	}

	e, err := engine.New(context.Background(), source.LocalPosixSource{}, cfg, noopEmitter{})
	if err != nil {
		t.Fatalf("engine.New: %v", err)
	}
	defer e.Close()

	_, _, err = e.CheckPlan(context.Background(), plan)
	if err == nil {
		t.Fatalf("CheckPlan must abort for non-deferrable errors")
	}
}

// TestCheck_DeferredPath_AncestorPromise verifies that a promised path like
// /foo/bar also defers errors for /foo (MkdirAll creates ancestors).
func TestCheck_DeferredPath_AncestorPromise(t *testing.T) {
	// dir action promises /foo/bar (MkdirAll would create /foo too)
	dirOp := &fakeOp{
		name:    "ensure-dir",
		checkFn: okCheckFn(spec.CheckUnsatisfied),
		execFn:  panicExecFn("check mode must not execute"),
	}
	dirAction := mkPromiserAction(nil, paths("/foo/bar"), dirOp)

	// copy needs /foo to exist (parent of /foo/file)
	// Input depends on /foo/bar so the graph orders dir before copy.
	copyOp := &fakeOp{
		name: "copy-file",
		checkFn: func(context.Context, source.Source, target.Target) (spec.CheckResult, []spec.DriftDetail, error) {
			return spec.CheckUnsatisfied, nil, copy.CopyDestDirMissingError{
				Path: "/foo",
			}
		},
		execFn: panicExecFn("check mode must not execute"),
	}
	copyAction := mkPromiserAction(paths("/foo/bar"), paths("/foo/file"), copyOp)

	plan := spec.Plan{
		Unit: spec.Unit{
			ID:      "test-ancestor-promise",
			Actions: []spec.Action{dirAction, copyAction},
		},
	}

	cfg := spec.ResolvedConfig{
		Target: mockTargetInstance(local.POSIXTarget{}),
	}

	e, err := engine.New(context.Background(), source.LocalPosixSource{}, cfg, noopEmitter{})
	if err != nil {
		t.Fatalf("engine.New: %v", err)
	}
	defer e.Close()

	rep, _, err := e.CheckPlan(context.Background(), plan)
	if err != nil {
		t.Fatalf("CheckPlan must not abort when ancestor path is promised, got: %v", err)
	}

	for i, ar := range rep.Actions {
		if ar.Summary.WouldChange == 0 {
			t.Errorf("action %d: expected WouldChange > 0", i)
		}
	}
}

// TestCheck_DeferredPath_OpOutcomeIsWouldChange verifies that deferred ops
// get OpWouldChange outcome (not OpAborted).
func TestCheck_DeferredPath_OpOutcomeIsWouldChange(t *testing.T) {
	dirOp := &fakeOp{
		name:    "ensure-dir",
		checkFn: okCheckFn(spec.CheckUnsatisfied),
		execFn:  panicExecFn("check mode must not execute"),
	}
	dirAction := mkPromiserAction(nil, paths("/foo"), dirOp)

	copyOp := &fakeOp{
		name: "copy-file",
		checkFn: func(context.Context, source.Source, target.Target) (spec.CheckResult, []spec.DriftDetail, error) {
			return spec.CheckUnsatisfied, nil, copy.CopyDestDirMissingError{
				Path: "/foo",
			}
		},
		execFn: panicExecFn("check mode must not execute"),
	}
	copyAction := mkPromiserAction(paths("/foo"), paths("/foo/bar"), copyOp)

	plan := spec.Plan{
		Unit: spec.Unit{
			ID:      "test-outcome",
			Actions: []spec.Action{dirAction, copyAction},
		},
	}

	cfg := spec.ResolvedConfig{
		Target: mockTargetInstance(local.POSIXTarget{}),
	}

	e, err := engine.New(context.Background(), source.LocalPosixSource{}, cfg, noopEmitter{})
	if err != nil {
		t.Fatalf("engine.New: %v", err)
	}
	defer e.Close()

	rep, _, err := e.CheckPlan(context.Background(), plan)
	if err != nil {
		t.Fatalf("CheckPlan: %v", err)
	}

	// The copy action's single op should be WouldChange
	copyReport := rep.Actions[1]
	if len(copyReport.Ops) != 1 {
		t.Fatalf("expected 1 op in copy action, got %d", len(copyReport.Ops))
	}

	if copyReport.Ops[0].Outcome != model.OpWouldChange {
		t.Errorf("deferred op outcome = %v, want OpWouldChange", copyReport.Ops[0].Outcome)
	}
	if copyReport.Ops[0].Err != nil {
		t.Errorf("deferred op should have nil error, got %v", copyReport.Ops[0].Err)
	}
}

// TestCheck_DeferredUser_UpstreamPromisesUser verifies that check mode does
// not abort when a downstream op reports an unknown user that an upstream
// action has promised to create.
func TestCheck_DeferredUser_UpstreamPromisesUser(t *testing.T) {
	// user action: check says "unsatisfied" (user doesn't exist yet)
	userOp := &fakeOp{
		name:    "ensure-user",
		checkFn: okCheckFn(spec.CheckUnsatisfied),
		execFn:  panicExecFn("check mode must not execute"),
	}
	userAction := mkPromiserAction(nil, users("appd"), userOp)

	// dir action: check returns UnknownUserError for appd
	dirOp := &fakeOp{
		name: "ensure-owner",
		checkFn: func(context.Context, source.Source, target.Target) (spec.CheckResult, []spec.DriftDetail, error) {
			return spec.CheckUnsatisfied, nil, sharedops.UnknownUserError{
				User: "appd",
			}
		},
		execFn: panicExecFn("check mode must not execute"),
	}
	dirAction := mkPromiserAction(users("appd"), paths("/opt/app"), dirOp)

	plan := spec.Plan{
		Unit: spec.Unit{
			ID:      "test-deferred-user",
			Actions: []spec.Action{userAction, dirAction},
		},
	}

	cfg := spec.ResolvedConfig{
		Target: mockTargetInstance(local.POSIXTarget{}),
	}

	e, err := engine.New(context.Background(), source.LocalPosixSource{}, cfg, noopEmitter{})
	if err != nil {
		t.Fatalf("engine.New: %v", err)
	}
	defer e.Close()

	rep, _, err := e.CheckPlan(context.Background(), plan)
	if err != nil {
		t.Fatalf("CheckPlan must not abort when user is promised, got: %v", err)
	}

	for i, ar := range rep.Actions {
		if ar.Summary.WouldChange == 0 {
			t.Errorf("action %d: expected WouldChange > 0, got %+v", i, ar.Summary)
		}
	}
}

// TestCheck_DeferredGroup_UpstreamPromisesGroup verifies the same for groups.
func TestCheck_DeferredGroup_UpstreamPromisesGroup(t *testing.T) {
	groupOp := &fakeOp{
		name:    "ensure-group",
		checkFn: okCheckFn(spec.CheckUnsatisfied),
		execFn:  panicExecFn("check mode must not execute"),
	}
	groupAction := mkPromiserAction(nil, groups("appusers"), groupOp)

	dirOp := &fakeOp{
		name: "ensure-owner",
		checkFn: func(context.Context, source.Source, target.Target) (spec.CheckResult, []spec.DriftDetail, error) {
			return spec.CheckUnsatisfied, nil, sharedops.UnknownGroupError{
				Group: "appusers",
			}
		},
		execFn: panicExecFn("check mode must not execute"),
	}
	dirAction := mkPromiserAction(groups("appusers"), paths("/opt/app"), dirOp)

	plan := spec.Plan{
		Unit: spec.Unit{
			ID:      "test-deferred-group",
			Actions: []spec.Action{groupAction, dirAction},
		},
	}

	cfg := spec.ResolvedConfig{
		Target: mockTargetInstance(local.POSIXTarget{}),
	}

	e, err := engine.New(context.Background(), source.LocalPosixSource{}, cfg, noopEmitter{})
	if err != nil {
		t.Fatalf("engine.New: %v", err)
	}
	defer e.Close()

	rep, _, err := e.CheckPlan(context.Background(), plan)
	if err != nil {
		t.Fatalf("CheckPlan must not abort when group is promised, got: %v", err)
	}

	for i, ar := range rep.Actions {
		if ar.Summary.WouldChange == 0 {
			t.Errorf("action %d: expected WouldChange > 0, got %+v", i, ar.Summary)
		}
	}
}

// TestCheck_DeferredUser_NoPromise_StillAborts verifies that an unknown user
// error still aborts when no upstream action promises the user.
func TestCheck_DeferredUser_NoPromise_StillAborts(t *testing.T) {
	dirOp := &fakeOp{
		name: "ensure-owner",
		checkFn: func(context.Context, source.Source, target.Target) (spec.CheckResult, []spec.DriftDetail, error) {
			return spec.CheckUnsatisfied, nil, sharedops.UnknownUserError{
				User: "nobody-promised",
			}
		},
		execFn: panicExecFn("check mode must not execute"),
	}
	dirAction := mkPromiserAction(nil, paths("/opt/app"), dirOp)

	plan := spec.Plan{
		Unit: spec.Unit{
			ID:      "test-user-no-promise",
			Actions: []spec.Action{dirAction},
		},
	}

	cfg := spec.ResolvedConfig{
		Target: mockTargetInstance(local.POSIXTarget{}),
	}

	e, err := engine.New(context.Background(), source.LocalPosixSource{}, cfg, noopEmitter{})
	if err != nil {
		t.Fatalf("engine.New: %v", err)
	}
	defer e.Close()

	_, _, err = e.CheckPlan(context.Background(), plan)
	if err == nil {
		t.Fatalf("CheckPlan must abort when no upstream action promises the user")
	}
}
