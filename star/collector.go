// SPDX-License-Identifier: GPL-3.0-only

package star

import (
	"go.starlark.net/starlark"

	"godoit.dev/doit/source"
	"godoit.dev/doit/spec"
)

const collectorKey = "collector"

// Collector accumulates targets and deploy blocks during Starlark evaluation.
type Collector struct {
	path    string
	targets map[string]spec.TargetInstance
	deploy  map[string]spec.DeployBlock
	sources *spec.SourceStore
	src     source.Source
}

func newCollector(path string, sources *spec.SourceStore, src source.Source) *Collector {
	return &Collector{
		path:    path,
		targets: make(map[string]spec.TargetInstance),
		deploy:  make(map[string]spec.DeployBlock),
		sources: sources,
		src:     src,
	}
}

func threadCollector(thread *starlark.Thread) *Collector {
	return thread.Local(collectorKey).(*Collector)
}

// AddTarget registers a target instance. Returns an error if the name
// is already taken.
func (c *Collector) AddTarget(name string, inst spec.TargetInstance, span spec.SourceSpan) error {
	if _, exists := c.targets[name]; exists {
		return &DuplicateTargetError{Name: name, Source: span}
	}
	c.targets[name] = inst
	return nil
}

// AddDeploy registers a deploy block. Returns an error if the name
// is already taken.
func (c *Collector) AddDeploy(name string, block spec.DeployBlock, span spec.SourceSpan) error {
	if _, exists := c.deploy[name]; exists {
		return &DuplicateDeployError{Name: name, Source: span}
	}
	c.deploy[name] = block
	return nil
}

// Config drains the collector into a spec.Config.
func (c *Collector) Config() spec.Config {
	cfg := spec.Config{
		Path:    c.path,
		Targets: c.targets,
		Deploy:  c.deploy,
	}
	if c.sources != nil {
		cfg.Sources = *c.sources
	}
	return cfg
}
