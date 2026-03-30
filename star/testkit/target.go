// SPDX-License-Identifier: GPL-3.0-only

package testkit

import (
	"context"

	"scampi.dev/scampi/source"
	"scampi.dev/scampi/spec"
	"scampi.dev/scampi/target"
)

// InMemoryTargetType wraps a pre-built MemTarget as a spec.TargetType.
// Its Create method returns the existing MemTarget without modification.
type InMemoryTargetType struct {
	Tgt *target.MemTarget
}

func (t InMemoryTargetType) Kind() string   { return "mem" }
func (t InMemoryTargetType) NewConfig() any { return nil }

func (t InMemoryTargetType) Create(
	_ context.Context,
	_ source.Source,
	_ spec.TargetInstance,
) (target.Target, error) {
	return t.Tgt, nil
}

// BuildMemTarget creates a MemTarget from pre-populated state.
func BuildMemTarget(
	files map[string]string,
	packages []string,
	services map[string]string,
	dirs []string,
) *target.MemTarget {
	tgt := target.NewMemTarget()

	for path, content := range files {
		tgt.Files[path] = []byte(content)
	}
	for _, pkg := range packages {
		tgt.Pkgs[pkg] = true
	}
	for name, state := range services {
		tgt.Services[name] = (state == "running")
		tgt.EnabledServices[name] = (state == "running")
	}
	for _, dir := range dirs {
		tgt.Dirs[dir] = 0o755
	}

	return tgt
}
