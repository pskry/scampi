package tasks

import (
	"context"
	"fmt"

	"godoit.dev/doit/spec"
)

type (
	CopySpec   struct{}
	CopyConfig struct {
		Src   string
		Dest  string
		Mode  string
		Owner string
		Group string
	}
	copyRtTask struct {
		idx int
		cfg CopyConfig
	}
)

func (CopySpec) Kind() string { return "copy" }
func (CopySpec) Schema() string {
	return `
package doit

#Task: {
  kind: string

  if kind == "copy" {
    copy: {
      src: string
      dest: string
      mode: string
      owner: string
      group: string
    }
  }
}`
}

func (CopySpec) NewConfig() any {
	return &CopyConfig{}
}

func (CopySpec) Plan(idx int, config any) (spec.RtTask, error) {
	cfg, ok := config.(*CopyConfig)
	if !ok {
		return nil, fmt.Errorf("expected %T got %T", &CopyConfig{}, config)
	}

	return &copyRtTask{
		idx: idx,
		cfg: *cfg,
	}, nil
}

func (c *copyRtTask) Name() string {
	return fmt.Sprintf("copy[%d]", c.idx)
}

func (c *copyRtTask) Ops() []spec.Op {
	copyFile := &copyFileOp{
		src:  c.cfg.Src,
		dest: c.cfg.Dest,
	}
	ensureOwner := &ensureOwnerOp{
		path:  c.cfg.Dest,
		owner: c.cfg.Owner,
		group: c.cfg.Group,
	}
	ensureMode := &ensureModeOp{
		path: c.cfg.Dest,
		mode: c.cfg.Group,
	}

	return []spec.Op{
		copyFile,
		ensureOwner,
		ensureMode,
	}
}

type (
	copyFileOp struct {
		src  string
		dest string
	}
	ensureOwnerOp struct {
		path  string
		owner string
		group string
	}
	ensureModeOp struct {
		path string
		mode string
	}
)

func (*copyFileOp) Name() string                                    { return "copyFileOp" }
func (*copyFileOp) Execute(context.Context) (spec.Result, error)    { return spec.Result{}, nil }
func (*ensureOwnerOp) Name() string                                 { return "ensureOwnerOp" }
func (*ensureOwnerOp) Execute(context.Context) (spec.Result, error) { return spec.Result{}, nil }
func (*ensureModeOp) Name() string                                  { return "ensureModeOp" }
func (*ensureModeOp) Execute(context.Context) (spec.Result, error)  { return spec.Result{}, nil }
