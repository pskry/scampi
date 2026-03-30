// SPDX-License-Identifier: GPL-3.0-only

package testkit

import (
	"go.starlark.net/starlark"

	"scampi.dev/scampi/spec"
	"scampi.dev/scampi/target"
)

var assertionAttrs = []string{
	"file",
	"dir",
	"service",
	"package",
	"symlink",
	"container",
	"command_ran",
}

// AssertionBuilder is the Starlark value returned by test.assert.that(t).
// Step packages call its attribute methods to register domain-specific assertions.
// Uses target.Target (interface) so assertions work against any target type.
type AssertionBuilder struct {
	tgt       target.Target
	collector *Collector
}

// Target returns the underlying target for use by assertion implementations.
func (b *AssertionBuilder) Target() target.Target { return b.tgt }

// NewAssertionBuilder returns an AssertionBuilder wrapping tgt and collector.
func NewAssertionBuilder(tgt target.Target, collector *Collector) *AssertionBuilder {
	return &AssertionBuilder{tgt: tgt, collector: collector}
}

func (b *AssertionBuilder) String() string        { return "assert_that" }
func (b *AssertionBuilder) Type() string          { return "assert_that" }
func (b *AssertionBuilder) Freeze()               {}
func (b *AssertionBuilder) Truth() starlark.Bool  { return starlark.True }
func (b *AssertionBuilder) Hash() (uint32, error) { return 0, nil }

func (b *AssertionBuilder) AttrNames() []string { return assertionAttrs }

func (b *AssertionBuilder) Attr(name string) (starlark.Value, error) {
	for _, attr := range assertionAttrs {
		if name == attr {
			return starlark.NewBuiltin("assert_that."+name, stubBuiltin), nil
		}
	}
	return nil, nil
}

// RegisterAssertion registers an assertion with the collector.
// Step packages call this from assertion value methods once they have resolved
// their check closure against the MemTarget.
func (b *AssertionBuilder) RegisterAssertion(desc string, source spec.SourceSpan, check func() error) {
	b.collector.Add(Assertion{Description: desc, Source: source, Check: check})
}

func stubBuiltin(
	_ *starlark.Thread,
	_ *starlark.Builtin,
	_ starlark.Tuple,
	_ []starlark.Tuple,
) (starlark.Value, error) {
	return starlark.None, nil
}
