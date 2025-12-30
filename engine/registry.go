package engine

import (
	"maps"
	"slices"

	"godoit.dev/doit/spec"
	"godoit.dev/doit/tasks"
)

type Registry struct {
	specs map[string]spec.Spec
}

func NewRegistry() (*Registry, error) {
	// TODO: this probably needs to be automatic at some point
	// also: this would be where we need to put extensions
	// for now (probably a while) this is just a manual list
	specs := []spec.Spec{
		tasks.CopySpec{},
	}

	r := &Registry{}
	r.specs = make(map[string]spec.Spec)
	for _, spec := range specs {
		r.specs[spec.Kind()] = spec
	}

	return r, nil
}

func (r *Registry) Specs() []spec.Spec {
	return slices.Collect(maps.Values(r.specs))
}

func (r *Registry) SpecForKind(kind string) (spec.Spec, bool) {
	spec, ok := r.specs[kind]
	return spec, ok
}
