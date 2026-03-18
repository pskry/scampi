// SPDX-License-Identifier: GPL-3.0-only

// Package group implements the group step type for managing system groups.
package group

import (
	"scampi.dev/scampi/errs"
	"scampi.dev/scampi/spec"
)

const (
	StatePresent = "present"
	StateAbsent  = "absent"
)

type (
	Group       struct{}
	GroupConfig struct {
		_ struct{} `summary:"Ensure a group exists or is absent on the target"`

		Desc   string `step:"Human-readable description" optional:"true"`
		Name   string `step:"Group name to manage" example:"appusers"`
		State  string `step:"Desired state" default:"present" example:"absent"`
		GID    int    `step:"Group ID" optional:"true" example:"1100"`
		System bool   `step:"Create as system group" optional:"true"`
	}
	groupAction struct {
		idx    int
		desc   string
		name   string
		state  string
		gid    int
		system bool
		step   spec.StepInstance
	}
)

func (Group) Kind() string   { return "group" }
func (Group) NewConfig() any { return &GroupConfig{} }

func (g Group) Plan(idx int, step spec.StepInstance) (spec.Action, error) {
	cfg, ok := step.Config.(*GroupConfig)
	if !ok {
		return nil, errs.BUG("expected %T got %T", &GroupConfig{}, step.Config)
	}

	if err := cfg.Validate(step); err != nil {
		return nil, err
	}

	return &groupAction{
		idx:    idx,
		desc:   cfg.Desc,
		name:   cfg.Name,
		state:  cfg.State,
		gid:    cfg.GID,
		system: cfg.System,
		step:   step,
	}, nil
}

func (c *GroupConfig) Validate(step spec.StepInstance) error {
	switch c.State {
	case StatePresent, StateAbsent:
	default:
		return InvalidStateError{
			Got:     c.State,
			Allowed: []string{StatePresent, StateAbsent},
			Source:  step.Fields["state"].Value,
		}
	}
	return nil
}

func (a *groupAction) Desc() string            { return a.desc }
func (a *groupAction) Kind() string            { return "group" }
func (a *groupAction) Inputs() []spec.Resource { return nil }

func (a *groupAction) Promises() []spec.Resource {
	if a.state == StatePresent {
		return []spec.Resource{spec.GroupResource(a.name)}
	}
	return nil
}

func (a *groupAction) Ops() []spec.Op {
	nameSource := a.step.Fields["name"].Value

	switch a.state {
	case StateAbsent:
		op := &removeGroupOp{
			name:       a.name,
			nameSource: nameSource,
		}
		op.SetAction(a)
		return []spec.Op{op}

	default:
		op := &ensureGroupOp{
			name:       a.name,
			gid:        a.gid,
			system:     a.system,
			nameSource: nameSource,
		}
		op.SetAction(a)
		return []spec.Op{op}
	}
}
