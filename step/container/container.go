// SPDX-License-Identifier: GPL-3.0-only

package container

import (
	"scampi.dev/scampi/errs"
	"scampi.dev/scampi/spec"
)

// State represents the desired container state.
type State uint8

const (
	StateRunning State = iota + 1
	StateStopped
	StateAbsent
)

func (s State) String() string {
	switch s {
	case StateRunning:
		return "running"
	case StateStopped:
		return "stopped"
	case StateAbsent:
		return "absent"
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
	case "absent":
		return StateAbsent
	default:
		panic(errs.BUG("invalid container state %q — should have been caught by validate", s))
	}
}

type (
	Instance       struct{}
	InstanceConfig struct {
		_ struct{} `summary:"Manage container lifecycle: running, stopped, or absent"`

		Desc    string   `step:"Human-readable description" optional:"true"`
		Name    string   `step:"Container name" example:"prometheus"`
		Image   string   `step:"Container image" example:"prom/prometheus:v3.2.0"`
		State   string   `step:"Desired container state" default:"running" example:"stopped|absent"`
		Restart string   `step:"Restart policy" default:"unless-stopped" example:"always|on-failure|no"`
		Ports   []string `step:"Port mappings (host:container)" optional:"true" example:"[\"9090:9090\"]"`
	}
	instanceAction struct {
		desc    string
		name    string
		image   string
		state   State
		restart string
		ports   []string
		step    spec.StepInstance
	}
)

func (Instance) Kind() string   { return "container.instance" }
func (Instance) NewConfig() any { return &InstanceConfig{} }

func (Instance) Plan(step spec.StepInstance) (spec.Action, error) {
	cfg, ok := step.Config.(*InstanceConfig)
	if !ok {
		return nil, errs.BUG("expected %T got %T", &InstanceConfig{}, step.Config)
	}

	if err := cfg.validate(step); err != nil {
		return nil, err
	}

	return &instanceAction{
		desc:    cfg.Desc,
		name:    cfg.Name,
		image:   cfg.Image,
		state:   parseState(cfg.State),
		restart: cfg.Restart,
		ports:   cfg.Ports,
		step:    step,
	}, nil
}

func (c *InstanceConfig) validate(step spec.StepInstance) error {
	switch c.State {
	case "running", "stopped", "absent":
	default:
		return InvalidStateError{
			Got:     c.State,
			Allowed: []string{"running", "stopped", "absent"},
			Source:  step.Fields["state"].Value,
		}
	}

	switch c.Restart {
	case "always", "on-failure", "unless-stopped", "no":
	default:
		return InvalidRestartError{
			Got:     c.Restart,
			Allowed: []string{"always", "on-failure", "unless-stopped", "no"},
			Source:  step.Fields["restart"].Value,
		}
	}

	if c.State != "absent" && c.Image == "" {
		return EmptyImageError{
			Source: step.Source,
		}
	}

	return nil
}

func (a *instanceAction) Desc() string { return a.desc }
func (a *instanceAction) Kind() string { return "container.instance" }

func (a *instanceAction) Ops() []spec.Op {
	op := &ensureContainerOp{
		name:    a.name,
		image:   a.image,
		state:   a.state,
		restart: a.restart,
		ports:   a.ports,
		step:    a.step,
	}
	op.SetAction(a)
	return []spec.Op{op}
}
