// SPDX-License-Identifier: GPL-3.0-only

// Package testkit implements the scampi test framework for Starlark modules.
package testkit

import (
	"sync"

	"scampi.dev/scampi/spec"
)

// Assertion is a single test check that runs after apply.
type Assertion struct {
	Description string
	Source      spec.SourceSpan
	Check       func() error
}

// Collector accumulates assertions during Starlark eval.
type Collector struct {
	mu         sync.Mutex
	assertions []Assertion
}

// NewCollector returns an empty Collector.
func NewCollector() *Collector {
	return &Collector{}
}

// Add appends an assertion to the collector.
func (c *Collector) Add(a Assertion) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.assertions = append(c.assertions, a)
}

// Assertions returns a snapshot of all registered assertions.
func (c *Collector) Assertions() []Assertion {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]Assertion, len(c.assertions))
	copy(out, c.assertions)
	return out
}
