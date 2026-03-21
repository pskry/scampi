// SPDX-License-Identifier: GPL-3.0-only

package ctrmgr

import (
	"fmt"
	"strings"

	"scampi.dev/scampi/target"
)

// Backend builds shell commands for a container runtime.
type Backend struct {
	name      string
	NeedsRoot bool
}

func (b *Backend) Name() string { return b.name }

func (b *Backend) CmdInspect(name string) string {
	return fmt.Sprintf("%s inspect --format '{{json .}}' %s", b.name, target.ShellQuote(name))
}

func (b *Backend) CmdCreate(opts CreateOpts) string {
	parts := []string{
		b.name, "create",
		"--name", target.ShellQuote(opts.Name),
		"--restart", target.ShellQuote(opts.Restart),
	}
	for _, p := range opts.Ports {
		parts = append(parts, "-p", target.ShellQuote(p))
	}
	parts = append(parts, target.ShellQuote(opts.Image))
	return strings.Join(parts, " ")
}

func (b *Backend) CmdStart(name string) string {
	return fmt.Sprintf("%s start %s", b.name, target.ShellQuote(name))
}

func (b *Backend) CmdStop(name string) string {
	return fmt.Sprintf("%s stop %s", b.name, target.ShellQuote(name))
}

func (b *Backend) CmdRm(name string) string {
	return fmt.Sprintf("%s rm %s", b.name, target.ShellQuote(name))
}

func (b *Backend) CmdPull(image string) string {
	return fmt.Sprintf("%s pull %s", b.name, target.ShellQuote(image))
}

// CreateOpts holds parameters for creating a container.
type CreateOpts struct {
	Name    string
	Image   string
	Restart string
	Ports   []string
}
