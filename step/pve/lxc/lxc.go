// SPDX-License-Identifier: GPL-3.0-only

package lxc

import (
	"scampi.dev/scampi/errs"
	"scampi.dev/scampi/spec"
)

// State
// -----------------------------------------------------------------------------

type State uint8

const (
	StateRunning State = iota + 1
	StateStopped
	StateAbsent
)

const (
	stateRunning = "running"
	stateStopped = "stopped"
	stateAbsent  = "absent"
)

var StateValues = []string{stateRunning, stateStopped, stateAbsent}

func (s State) String() string {
	switch s {
	case StateRunning:
		return stateRunning
	case StateStopped:
		return stateStopped
	case StateAbsent:
		return stateAbsent
	default:
		return "unknown"
	}
}

func parseState(s string) State {
	switch s {
	case stateRunning:
		return StateRunning
	case stateStopped:
		return StateStopped
	case stateAbsent:
		return StateAbsent
	default:
		panic(errs.BUG("invalid lxc state %q — should have been caught by validate", s))
	}
}

// Config
// -----------------------------------------------------------------------------

type (
	LXC       struct{}
	LxcConfig struct {
		_ struct{} `summary:"Manage LXC container lifecycle on Proxmox VE via pct"`

		ID       int          `step:"Container VMID (unique per cluster)" example:"100"`
		Node     string       `step:"PVE node name" example:"pve1"`
		Template LxcTemplate `step:"OS template"`
		Hostname string       `step:"Container hostname" example:"pihole"`
		State    string       `step:"Desired state" default:"running"`
		Cores    int          `step:"CPU cores" default:"1"`
		Memory   int          `step:"Memory in MiB" default:"512"`
		Storage  string       `step:"Storage pool for rootfs" default:"local-zfs"`
		Size     string       `step:"Root disk size" default:"8G"`
		Network  LxcNet       `step:"Network configuration"`
		Desc     string       `step:"Human-readable description" optional:"true"`
	}
	LxcTemplate struct {
		Storage string `step:"Storage pool holding the template" default:"local"`
		Name    string `step:"Template filename" example:"debian-12-standard_12.7-1_amd64.tar.zst"`
	}
	LxcNet struct {
		Bridge string `step:"Bridge interface" default:"vmbr0"`
		IP     string `step:"IP address in CIDR or dhcp" example:"10.10.10.10/24"`
		Gw     string `step:"Gateway" optional:"true" example:"10.10.10.1"`
	}
)

func (*LxcConfig) FieldEnumValues() map[string][]string {
	return map[string][]string{
		"state": StateValues,
	}
}

func (LXC) Kind() string   { return "pve.lxc" }
func (LXC) NewConfig() any { return &LxcConfig{} }

func (LXC) Plan(step spec.StepInstance) (spec.Action, error) {
	cfg, ok := step.Config.(*LxcConfig)
	if !ok {
		return nil, errs.BUG("expected %T got %T", &LxcConfig{}, step.Config)
	}

	if err := cfg.validate(step); err != nil {
		return nil, err
	}

	return &lxcAction{
		desc:     cfg.Desc,
		id:       cfg.ID,
		node:     cfg.Node,
		template: cfg.Template,
		hostname: cfg.Hostname,
		state:    parseState(cfg.State),
		cores:    cfg.Cores,
		memory:   cfg.Memory,
		storage:  cfg.Storage,
		size:     cfg.Size,
		network:  cfg.Network,
		step:     step,
	}, nil
}

// templatePath returns the full PVE template path for pct create.
func (t LxcTemplate) templatePath() string {
	return t.Storage + ":vztmpl/" + t.Name
}

func (c *LxcConfig) validate(step spec.StepInstance) error {
	if c.ID < 100 {
		return InvalidConfigError{
			Field:  "id",
			Reason: "VMID must be >= 100 (PVE reserves 0-99)",
			Source: step.Fields["id"].Value,
		}
	}
	if c.Node == "" {
		return InvalidConfigError{
			Field:  "node",
			Reason: "node name is required",
			Source: step.Fields["node"].Value,
		}
	}
	if c.Hostname == "" {
		return InvalidConfigError{
			Field:  "hostname",
			Reason: "hostname is required",
			Source: step.Fields["hostname"].Value,
		}
	}
	if c.Template.Name == "" {
		return InvalidConfigError{
			Field:  "template.name",
			Reason: "template name is required",
			Source: step.Fields["template"].Value,
		}
	}
	if c.Network.IP == "" {
		return InvalidConfigError{
			Field:  "network.ip",
			Reason: "network IP address is required",
			Source: step.Fields["network"].Value,
		}
	}
	return nil
}

// Action
// -----------------------------------------------------------------------------

type lxcAction struct {
	desc     string
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

func (a *lxcAction) Desc() string { return a.desc }
func (a *lxcAction) Kind() string { return "pve.lxc" }

func (a *lxcAction) Ops() []spec.Op {
	dlOp := &downloadTemplateOp{
		template: a.template,
		step:     a.step,
	}
	dlOp.SetAction(a)

	lxcOp := &ensureLxcOp{
		id:       a.id,
		node:     a.node,
		template: a.template,
		hostname: a.hostname,
		state:    a.state,
		cores:    a.cores,
		memory:   a.memory,
		storage:  a.storage,
		size:     a.size,
		network:  a.network,
		step:     a.step,
	}
	lxcOp.SetAction(a)
	lxcOp.AddDependency(dlOp)

	return []spec.Op{dlOp, lxcOp}
}
