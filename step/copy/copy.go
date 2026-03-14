// SPDX-License-Identifier: GPL-3.0-only

package copy

import (
	"io/fs"
	"path/filepath"

	"scampi.dev/scampi/errs"
	"scampi.dev/scampi/spec"
	"scampi.dev/scampi/step/sharedops"
	"scampi.dev/scampi/step/sharedops/fileops"
)

type (
	Copy       struct{}
	CopyConfig struct {
		_ struct{} `summary:"Copy files with owner and permission management"`

		Desc    string `step:"Human-readable description" optional:"true"`
		Src     string `step:"Source file path" optional:"true" exclusive:"source" example:"./config.yaml"`
		Content string `step:"Inline file content" optional:"true" exclusive:"source"`
		Dest    string `step:"Destination file path" example:"/etc/app/config.yaml"`
		Perm    string `step:"File permissions" example:"0644|u=rw,g=r,o=r|rw-r--r--"`
		Owner   string `step:"Owner user name or UID" example:"root"`
		Group   string `step:"Group name or GID" example:"root"`
	}
	copyAction struct {
		idx     int
		desc    string
		kind    string
		src     string
		content string
		dest    string
		mode    fs.FileMode
		owner   string
		group   string
		step    spec.StepInstance
	}
)

func (Copy) Kind() string   { return "copy" }
func (Copy) NewConfig() any { return &CopyConfig{} }

func (c *CopyConfig) Validate(step spec.StepInstance) error {
	hasSrc := c.Src != ""
	hasContent := c.Content != ""
	if hasSrc == hasContent {
		var got []string
		source := step.Source
		if hasSrc {
			got = []string{"src", "content"}
			source = step.Fields["content"].Value
		}
		return MutuallyExclusiveError{
			Fields: []string{"src", "content"},
			Got:    got,
			Source: source,
		}
	}
	return nil
}

func (c Copy) Plan(idx int, step spec.StepInstance) (spec.Action, error) {
	cfg, ok := step.Config.(*CopyConfig)
	if !ok {
		return nil, errs.BUG("expected %T got %T", &CopyConfig{}, step.Config)
	}

	if err := cfg.Validate(step); err != nil {
		return nil, err
	}

	if !filepath.IsAbs(cfg.Dest) {
		return nil, sharedops.RelativePathError{
			Field:  "dest",
			Path:   cfg.Dest,
			Source: step.Fields["dest"].Value,
		}
	}

	mode, err := fileops.ParsePerm(cfg.Perm, step.Fields["perm"].Value)
	if err != nil {
		return nil, err
	}

	return &copyAction{
		idx:     idx,
		desc:    cfg.Desc,
		kind:    c.Kind(),
		src:     cfg.Src,
		content: cfg.Content,
		dest:    cfg.Dest,
		mode:    mode,
		owner:   cfg.Owner,
		group:   cfg.Group,

		step: step,
	}, nil
}

func (c *copyAction) Desc() string { return c.desc }
func (c *copyAction) Kind() string { return c.kind }
func (c *copyAction) InputPaths() []string {
	if c.src != "" {
		return []string{c.src}
	}
	return nil
}
func (c *copyAction) OutputPaths() []string { return []string{c.dest} }

func (c *copyAction) Ops() []spec.Op {
	cp := &copyFileOp{
		BaseOp: sharedops.BaseOp{
			SrcSpan:  c.step.Fields["src"].Value,
			DestSpan: c.step.Fields["dest"].Value,
		},
		src:     c.src,
		content: c.content,
		dest:    c.dest,
	}
	chown := &fileops.EnsureOwnerOp{
		BaseOp: sharedops.BaseOp{
			DestSpan: c.step.Fields["dest"].Value,
		},
		Path:      c.dest,
		Owner:     c.owner,
		Group:     c.group,
		OwnerSpan: c.step.Fields["owner"].Value,
		GroupSpan: c.step.Fields["group"].Value,
	}
	chmod := &fileops.EnsureModeOp{
		BaseOp: sharedops.BaseOp{
			DestSpan: c.step.Fields["dest"].Value,
		},
		Path: c.dest,
		Mode: c.mode,
	}

	cp.SetAction(c)
	chown.SetAction(c)
	chmod.SetAction(c)

	chown.AddDependency(cp)
	chmod.AddDependency(cp)

	return []spec.Op{
		cp,
		chown,
		chmod,
	}
}
