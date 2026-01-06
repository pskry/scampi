package engine

import (
	"maps"
	"slices"

	"godoit.dev/doit/kinds"
	"godoit.dev/doit/spec"
)

type Registry struct {
	impls map[string]spec.KindImpl
}

func NewRegistry() (*Registry, error) {
	// TODO: this probably needs to be automatic at some point
	// also: this would be where we need to put extensions
	// for now (probably a while) this is just a manual list
	impls := []spec.KindImpl{
		kinds.CopySpec{},
	}

	r := &Registry{}
	r.impls = make(map[string]spec.KindImpl)
	for _, spec := range impls {
		r.impls[spec.Kind()] = spec
	}

	return r, nil
}

func (r *Registry) Impls() []spec.KindImpl {
	return slices.Collect(maps.Values(r.impls))
}

func (r *Registry) ImplForKind(kind string) (spec.KindImpl, bool) {
	spec, ok := r.impls[kind]
	return spec, ok
}
