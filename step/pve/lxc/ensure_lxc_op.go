// SPDX-License-Identifier: GPL-3.0-only

package lxc

import (
	"context"
	"fmt"

	"scampi.dev/scampi/capability"
	"scampi.dev/scampi/source"
	"scampi.dev/scampi/spec"
	"scampi.dev/scampi/step/sharedops"
	"scampi.dev/scampi/target"
)

const ensureLxcID = "step.pve.lxc"

type ensureLxcOp struct {
	sharedops.BaseOp
	id       int
	node     string
	template LxcTemplate
	hostname string
	state    State
	cores    int
	memory   int
	storage  string
	size     string
	network  LxcNet
	step     spec.StepInstance
}

func (op *ensureLxcOp) Check(
	ctx context.Context,
	_ source.Source,
	tgt target.Target,
) (spec.CheckResult, []spec.DriftDetail, error) {
	if op.state == StateAbsent {
		return spec.CheckUnsatisfied, nil, UnsupportedStateError{
			State:  stateAbsent,
			Source: op.step.Fields["state"].Value,
		}
	}

	cmdr := target.Must[target.Command](ensureLxcID, tgt)

	exists, status, err := op.inspect(ctx, cmdr)
	if err != nil {
		return spec.CheckUnsatisfied, nil, err
	}

	if !exists {
		return spec.CheckUnsatisfied, []spec.DriftDetail{{
			Field:   "state",
			Current: "(absent)",
			Desired: op.state.String(),
		}}, nil
	}

	switch op.state {
	case StateRunning:
		if status == stateRunning {
			return spec.CheckSatisfied, nil, nil
		}
		return spec.CheckUnsatisfied, []spec.DriftDetail{{
			Field:   "state",
			Current: status,
			Desired: stateRunning,
		}}, nil

	case StateStopped:
		if status == stateStopped {
			return spec.CheckSatisfied, nil, nil
		}
		return spec.CheckUnsatisfied, []spec.DriftDetail{{
			Field:   "state",
			Current: status,
			Desired: stateStopped,
		}}, nil
	}

	return spec.CheckSatisfied, nil, nil
}

func (op *ensureLxcOp) Execute(
	ctx context.Context,
	_ source.Source,
	tgt target.Target,
) (spec.Result, error) {
	cmdr := target.Must[target.Command](ensureLxcID, tgt)

	exists, status, err := op.inspect(ctx, cmdr)
	if err != nil {
		return spec.Result{}, err
	}

	if !exists {
		return op.executeCreate(ctx, cmdr)
	}

	switch op.state {
	case StateRunning:
		if status != stateRunning {
			return op.executeStart(ctx, cmdr)
		}
	case StateStopped:
		if status != stateStopped {
			return op.executeStop(ctx, cmdr)
		}
	}

	return spec.Result{}, nil
}

// inspect checks whether the container exists and its running status.
func (op *ensureLxcOp) inspect(ctx context.Context, cmdr target.Command) (exists bool, status string, err error) {
	result, err := cmdr.RunCommand(ctx, "pct list")
	if err != nil {
		return false, "", op.cmdErrWrap("list", err)
	}
	if result.ExitCode != 0 {
		return false, "", op.cmdErrStr("list", result.Stderr)
	}

	entries := parsePctList(result.Stdout)
	if _, ok := entries[op.id]; !ok {
		return false, "", nil
	}

	result, err = cmdr.RunCommand(ctx, fmt.Sprintf("pct status %d", op.id))
	if err != nil {
		return false, "", op.cmdErrWrap("status", err)
	}
	if result.ExitCode != 0 {
		return false, "", op.cmdErrStr("status", result.Stderr)
	}

	return true, parsePctStatus(result.Stdout), nil
}

func (op *ensureLxcOp) executeCreate(ctx context.Context, cmdr target.Command) (spec.Result, error) {
	// Create container (template already ensured by downloadTemplateOp)
	if err := op.runCmd(ctx, cmdr, "create", buildCreateCmd(*op.Action().(*lxcAction))); err != nil {
		return spec.Result{}, err
	}

	// Start if desired
	if op.state == StateRunning {
		if err := op.runCmd(ctx, cmdr, "start", fmt.Sprintf("pct start %d", op.id)); err != nil {
			return spec.Result{}, err
		}
	}

	return spec.Result{Changed: true}, nil
}

func (op *ensureLxcOp) executeStart(ctx context.Context, cmdr target.Command) (spec.Result, error) {
	if err := op.runCmd(ctx, cmdr, "start", fmt.Sprintf("pct start %d", op.id)); err != nil {
		return spec.Result{}, err
	}
	return spec.Result{Changed: true}, nil
}

func (op *ensureLxcOp) executeStop(ctx context.Context, cmdr target.Command) (spec.Result, error) {
	if err := op.runCmd(ctx, cmdr, "shutdown", fmt.Sprintf("pct shutdown %d --timeout 30", op.id)); err != nil {
		return spec.Result{}, err
	}
	return spec.Result{Changed: true}, nil
}

// runCmd executes a command and wraps any failure as a CommandFailedError.
func (op *ensureLxcOp) runCmd(ctx context.Context, cmdr target.Command, opName, cmd string) error {
	result, err := cmdr.RunCommand(ctx, cmd)
	if err != nil {
		return op.cmdErrWrap(opName, err)
	}
	if result.ExitCode != 0 {
		return op.cmdErrStr(opName, result.Stderr)
	}
	return nil
}

func (op *ensureLxcOp) cmdErrWrap(operation string, err error) CommandFailedError {
	return CommandFailedError{
		Op:     operation,
		VMID:   op.id,
		Stderr: err.Error(),
		Source: op.step.Source,
	}
}

func (op *ensureLxcOp) cmdErrStr(operation, stderr string) CommandFailedError {
	return CommandFailedError{
		Op:     operation,
		VMID:   op.id,
		Stderr: stderr,
		Source: op.step.Source,
	}
}

func (ensureLxcOp) RequiredCapabilities() capability.Capability {
	return capability.PVE | capability.Command
}

// OpDescription
// -----------------------------------------------------------------------------

type ensureLxcDesc struct {
	VMID     int
	Hostname string
	State    string
}

func (d ensureLxcDesc) PlanTemplate() spec.PlanTemplate {
	return spec.PlanTemplate{
		ID:   ensureLxcID,
		Text: `ensure LXC {{.VMID}} "{{.Hostname}}" is {{.State}}`,
		Data: d,
	}
}

func (op *ensureLxcOp) OpDescription() spec.OpDescription {
	return ensureLxcDesc{
		VMID:     op.id,
		Hostname: op.hostname,
		State:    op.state.String(),
	}
}

func (op *ensureLxcOp) Inspect() []spec.InspectField {
	fields := []spec.InspectField{
		{Label: "vmid", Value: fmt.Sprintf("%d", op.id)},
		{Label: "node", Value: op.node},
		{Label: "hostname", Value: op.hostname},
		{Label: "state", Value: op.state.String()},
		{Label: "template", Value: op.template.templatePath()},
		{Label: "cores", Value: fmt.Sprintf("%d", op.cores)},
		{Label: "memory", Value: fmt.Sprintf("%d MiB", op.memory)},
		{Label: "storage", Value: op.storage},
		{Label: "size", Value: op.size},
		{Label: "network", Value: formatNet0(op.network)},
	}
	return fields
}
