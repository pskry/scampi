// SPDX-License-Identifier: GPL-3.0-only

// Package mount implements the mount step for managing filesystem mounts.
package mount

import (
	"path/filepath"

	"scampi.dev/scampi/errs"
	"scampi.dev/scampi/spec"
	"scampi.dev/scampi/step/sharedops"
)

// FsType represents a supported filesystem type.
type FsType uint8

const (
	FsNFS FsType = iota + 1
	FsNFS4
	FsCIFS
	FsExt4
	FsXFS
	FsBtrfs
	FsTmpfs
	FsGlusterfs
	FsCeph
)

func (f FsType) String() string {
	switch f {
	case FsNFS:
		return "nfs"
	case FsNFS4:
		return "nfs4"
	case FsCIFS:
		return "cifs"
	case FsExt4:
		return "ext4"
	case FsXFS:
		return "xfs"
	case FsBtrfs:
		return "btrfs"
	case FsTmpfs:
		return "tmpfs"
	case FsGlusterfs:
		return "glusterfs"
	case FsCeph:
		return "ceph"
	default:
		return "unknown"
	}
}

func parseFsType(s string) (FsType, bool) {
	switch s {
	case "nfs":
		return FsNFS, true
	case "nfs4":
		return FsNFS4, true
	case "cifs":
		return FsCIFS, true
	case "ext4":
		return FsExt4, true
	case "xfs":
		return FsXFS, true
	case "btrfs":
		return FsBtrfs, true
	case "tmpfs":
		return FsTmpfs, true
	case "glusterfs":
		return FsGlusterfs, true
	case "ceph":
		return FsCeph, true
	default:
		return 0, false
	}
}

// NeedsHelper reports whether this filesystem type requires a userspace
// mount helper (e.g. mount.nfs, mount.cifs).
func (f FsType) NeedsHelper() bool {
	return f.HelperBinary() != ""
}

// HelperBinary returns the mount helper binary name, or empty if none needed.
func (f FsType) HelperBinary() string {
	switch f {
	case FsNFS, FsNFS4:
		return "mount.nfs"
	case FsCIFS:
		return "mount.cifs"
	case FsGlusterfs:
		return "mount.glusterfs"
	case FsCeph:
		return "mount.ceph"
	default:
		return ""
	}
}

// State represents the desired mount state.
type State uint8

const (
	StateMounted State = iota + 1
	StateUnmounted
	StateAbsent
)

func (s State) String() string {
	switch s {
	case StateMounted:
		return "mounted"
	case StateUnmounted:
		return "unmounted"
	case StateAbsent:
		return "absent"
	default:
		return "unknown"
	}
}

type (
	Mount       struct{}
	MountConfig struct {
		_ struct{} `summary:"Manage filesystem mounts and fstab entries"`

		Desc  string `step:"Human-readable description" optional:"true"`
		Src   string `step:"Mount source (device or remote path)" example:"10.10.2.2:/volume2/data"`
		Dest  string `step:"Mount point path" example:"/mnt/data"`
		Type  string `step:"Filesystem type" example:"nfs"`
		Opts  string `step:"Mount options" optional:"true" default:"defaults" example:"defaults,noatime"`
		State string `step:"Desired state" optional:"true" default:"mounted" example:"mounted|unmounted|absent"`
	}
	mountAction struct {
		desc  string
		src   string
		dest  string
		fstyp FsType
		opts  string
		state State
		step  spec.StepInstance
	}
)

func (Mount) Kind() string   { return "mount" }
func (Mount) NewConfig() any { return &MountConfig{} }

func (Mount) Plan(step spec.StepInstance) (spec.Action, error) {
	cfg, ok := step.Config.(*MountConfig)
	if !ok {
		return nil, errs.BUG("expected %T got %T", &MountConfig{}, step.Config)
	}

	if cfg.Dest == "" {
		return nil, MissingFieldError{
			Field:  "dest",
			Source: step.Source,
		}
	}

	if cfg.Type == "" {
		return nil, MissingFieldError{
			Field:  "type",
			Source: step.Source,
		}
	}

	fstyp, ok := parseFsType(cfg.Type)
	if !ok {
		return nil, InvalidTypeError{
			Got:    cfg.Type,
			Source: step.Fields["type"].Value,
		}
	}

	if cfg.Src == "" {
		return nil, MissingFieldError{
			Field:  "src",
			Source: step.Source,
		}
	}

	if !filepath.IsAbs(cfg.Dest) {
		return nil, sharedops.RelativePathError{
			Field:  "dest",
			Path:   cfg.Dest,
			Source: step.Fields["dest"].Value,
		}
	}

	opts := cfg.Opts
	if opts == "" {
		opts = "defaults"
	}

	var state State
	switch cfg.State {
	case "", "mounted":
		state = StateMounted
	case "unmounted":
		state = StateUnmounted
	case "absent":
		state = StateAbsent
	default:
		return nil, InvalidStateError{
			Got:    cfg.State,
			Source: step.Fields["state"].Value,
		}
	}

	if state == StateAbsent && cfg.Src == "" {
		cfg.Src = "*"
	}

	return &mountAction{
		desc:  cfg.Desc,
		src:   cfg.Src,
		dest:  cfg.Dest,
		fstyp: fstyp,
		opts:  opts,
		state: state,
		step:  step,
	}, nil
}

func (a *mountAction) Desc() string { return a.desc }
func (a *mountAction) Kind() string { return "mount" }
func (a *mountAction) Ops() []spec.Op {
	op := &ensureMountOp{
		src:   a.src,
		dest:  a.dest,
		fstyp: a.fstyp,
		opts:  a.opts,
		state: a.state,
	}
	op.SetAction(a)
	return []spec.Op{op}
}
