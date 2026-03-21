// SPDX-License-Identifier: GPL-3.0-only

package service

import (
	"scampi.dev/scampi/errs"
	"scampi.dev/scampi/spec"
)

// State represents the desired service state.
type State uint8

const (
	StateRunning State = iota + 1
	StateStopped
	StateRestarted
	StateReloaded
)

func (s State) String() string {
	switch s {
	case StateRunning:
		return "running"
	case StateStopped:
		return "stopped"
	case StateRestarted:
		return "restarted"
	case StateReloaded:
		return "reloaded"
	default:
		return "unknown"
	}
}

func parseState(s string) State {
	switch s {
	case "running":
		return StateRunning
	case "stopped":
		return StateStopped
	case "restarted":
		return StateRestarted
	case "reloaded":
		return StateReloaded
	default:
		panic(errs.BUG("invalid service state %q — should have been caught by Validate", s))
	}
}

type (
	Service       struct{}
	ServiceConfig struct {
		_ struct{} `summary:"Manage service state: running, stopped, restarted, or reloaded"`

		Desc    string `step:"Human-readable description" optional:"true"`
		Name    string `step:"Service name" example:"nginx"`
		State   string `step:"Desired service state" default:"running" example:"stopped"`
		Enabled bool   `step:"Whether the service should start at boot" default:"true"`
	}
	serviceAction struct {
		desc    string
		name    string
		state   State
		enabled bool
		step    spec.StepInstance
	}
)

func (Service) Kind() string   { return "service" }
func (Service) NewConfig() any { return &ServiceConfig{} }

func (s Service) Plan(step spec.StepInstance) (spec.Action, error) {
	cfg, ok := step.Config.(*ServiceConfig)
	if !ok {
		return nil, errs.BUG("expected %T got %T", &ServiceConfig{}, step.Config)
	}

	if err := cfg.Validate(step); err != nil {
		return nil, err
	}

	return &serviceAction{
		desc:    cfg.Desc,
		name:    cfg.Name,
		state:   parseState(cfg.State),
		enabled: cfg.Enabled,
		step:    step,
	}, nil
}

func (c *ServiceConfig) Validate(step spec.StepInstance) error {
	switch c.State {
	case "running", "stopped", "restarted", "reloaded":
	default:
		return InvalidStateError{
			Got:     c.State,
			Allowed: []string{"running", "stopped", "restarted", "reloaded"},
			Source:  step.Fields["state"].Value,
		}
	}
	return nil
}

func (a *serviceAction) Desc() string { return a.desc }
func (a *serviceAction) Kind() string { return "service" }

func (a *serviceAction) Ops() []spec.Op {
	switch a.state {
	case StateRestarted:
		op := &restartOp{
			name:       a.name,
			nameSource: a.step.Fields["name"].Value,
		}
		op.SetAction(a)
		return []spec.Op{op}

	case StateReloaded:
		op := &reloadOp{
			name:       a.name,
			nameSource: a.step.Fields["name"].Value,
		}
		op.SetAction(a)
		return []spec.Op{op}

	default:
		activeOp := &ensureActiveOp{
			name:       a.name,
			state:      a.state,
			nameSource: a.step.Fields["name"].Value,
		}
		activeOp.SetAction(a)

		enabledOp := &ensureEnabledOp{
			name:       a.name,
			enabled:    a.enabled,
			nameSource: a.step.Fields["name"].Value,
		}
		enabledOp.SetAction(a)

		return []spec.Op{activeOp, enabledOp}
	}
}
