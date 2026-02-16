// SPDX-License-Identifier: GPL-3.0-only

// Package secret provides the Backend interface for resolving secret values
// and implementations of that interface.
package secret

// Backend resolves secret values by key.
type Backend interface {
	Name() string
	Lookup(key string) (string, bool, error)
}
