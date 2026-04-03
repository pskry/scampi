// SPDX-License-Identifier: GPL-3.0-only

// Package secret provides the Backend interface for resolving secret values
// and implementations of that interface.
package secret

// Backend resolves secret values by key.
type Backend interface {
	Name() string
	Lookup(key string) (string, bool, error)
}

// PlaceholderBackend returns placeholder values for all known keys.
// Used by the LSP to continue evaluation without real secret access.
// Cause records why the real backend failed, so the LSP can surface it
// as a non-fatal hint.
type PlaceholderBackend struct {
	Keys  map[string]string
	Cause error
}

func (p *PlaceholderBackend) Name() string { return "placeholder" }

func (p *PlaceholderBackend) Lookup(key string) (string, bool, error) {
	if p.Keys == nil {
		return "<secret>", true, nil
	}
	v, ok := p.Keys[key]
	return v, ok, nil
}
