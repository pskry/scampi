// SPDX-License-Identifier: GPL-3.0-only

package lxc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"scampi.dev/scampi/capability"
	"scampi.dev/scampi/source"
	"scampi.dev/scampi/spec"
	"scampi.dev/scampi/step/sharedops"
	"scampi.dev/scampi/target"
)

const rebootLxcID = "step.pve.lxc.reboot"

// rebootCheck probes the running container for a field that needs a reboot
// to take effect. Returns nil if the field is already converged.
type rebootCheck struct {
	field   string
	desired string
	probe   func(ctx context.Context, cmdr target.Command, id int) string
}

// rebootChecks lists all fields that are written to config immediately
// but only take effect after a container reboot.
func buildRebootChecks(op *rebootLxcOp) []rebootCheck {
	checks := []rebootCheck{
		{
			field:   "hostname",
			desired: op.hostname,
			probe: func(ctx context.Context, cmdr target.Command, id int) string {
				r, err := cmdr.RunPrivileged(ctx, fmt.Sprintf("pct exec %d -- hostname", id))
				if err != nil || r.ExitCode != 0 {
					return ""
				}
				return strings.TrimSpace(r.Stdout)
			},
		},
	}

	// Features: compare running config's features against desired.
	// pct config shows the written value; the running container may
	// still have the old one. We compare desired vs current config
	// and if they differ, the container needs a reboot.
	if op.features != nil {
		checks = append(checks, rebootCheck{
			field:   "features",
			desired: formatFeatures(op.features),
			probe: func(ctx context.Context, cmdr target.Command, _ int) string {
				cfg, err := op.inspectConfig(ctx, cmdr)
				if err != nil {
					return ""
				}
				return formatFeatures(&cfg.Features)
			},
		})
	}

	return checks
}

type rebootLxcOp struct {
	sharedops.BaseOp
	pveCmd
	hostname string
	features *LxcFeatures
}

func (op *rebootLxcOp) Check(
	ctx context.Context,
	_ source.Source,
	tgt target.Target,
) (spec.CheckResult, []spec.DriftDetail, error) {
	cmdr := target.Must[target.Command](rebootLxcID, tgt)
	return op.checkWith(ctx, cmdr)
}

func (op *rebootLxcOp) checkWith(
	ctx context.Context,
	cmdr target.Command,
) (spec.CheckResult, []spec.DriftDetail, error) {
	result, err := cmdr.RunPrivileged(ctx, fmt.Sprintf("pct status %d", op.id))
	if err != nil || result.ExitCode != 0 {
		return spec.CheckSatisfied, nil, nil
	}
	if parsePctStatus(result.Stdout) != stateRunning {
		return spec.CheckSatisfied, nil, nil
	}

	var drift []spec.DriftDetail
	for _, check := range buildRebootChecks(op) {
		current := check.probe(ctx, cmdr, op.id)
		if current == "" || current == check.desired {
			continue
		}
		drift = append(drift, spec.DriftDetail{
			Field:   check.field + " (reboot)",
			Current: current,
			Desired: check.desired,
		})
	}

	if len(drift) == 0 {
		return spec.CheckSatisfied, nil, nil
	}
	return spec.CheckUnsatisfied, drift, nil
}

func (op *rebootLxcOp) Execute(
	ctx context.Context,
	_ source.Source,
	tgt target.Target,
) (spec.Result, error) {
	cmdr := target.Must[target.Command](rebootLxcID, tgt)

	// Re-check: container may have been stopped since Check ran.
	result, err := cmdr.RunPrivileged(ctx, fmt.Sprintf("pct status %d", op.id))
	if err == nil && parsePctStatus(result.Stdout) != stateRunning {
		return spec.Result{}, nil
	}

	cmd := fmt.Sprintf("pct reboot %d --timeout 30", op.id)
	if err := op.runCmd(ctx, cmdr, "reboot", cmd); err != nil {
		return spec.Result{}, err
	}

	if err := op.waitRunning(ctx, cmdr); err != nil {
		return spec.Result{}, err
	}

	return spec.Result{Changed: true}, nil
}

func (op *rebootLxcOp) waitRunning(
	ctx context.Context,
	cmdr target.Command,
) error {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return op.cmdErr("reboot", "container did not come back within 60s")
		case <-ticker.C:
			r, err := cmdr.RunPrivileged(ctx, fmt.Sprintf("pct status %d", op.id))
			if err != nil || r.ExitCode != 0 {
				continue
			}
			if parsePctStatus(r.Stdout) == stateRunning {
				return nil
			}
		}
	}
}

func (rebootLxcOp) RequiredCapabilities() capability.Capability {
	return capability.PVE | capability.Command
}

// OpDescription
// -----------------------------------------------------------------------------

type rebootLxcDesc struct {
	VMID int
}

func (d rebootLxcDesc) PlanTemplate() spec.PlanTemplate {
	return spec.PlanTemplate{
		ID:   rebootLxcID,
		Text: `reboot LXC {{.VMID}} if needed`,
		Data: d,
	}
}

func (op *rebootLxcOp) OpDescription() spec.OpDescription {
	return rebootLxcDesc{VMID: op.id}
}

func (op *rebootLxcOp) Inspect() []spec.InspectField {
	return []spec.InspectField{
		{Label: "vmid", Value: fmt.Sprintf("%d", op.id)},
	}
}
