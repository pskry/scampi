// SPDX-License-Identifier: GPL-3.0-only

package linker

import (
	"scampi.dev/scampi/diagnostic"
	"scampi.dev/scampi/lang/ast"
	"scampi.dev/scampi/spec"
)

// AttributeBehaviour is the scampi-specific semantics attached to an
// attribute type. Lang owns the schema (declared via `type @name { ... }`
// in stubs); the linker owns what an attribute *means* at link time and
// at eval time. The LSP separately owns UX semantics (completion,
// hover) keyed by the same qualified name.
//
// Implementations are looked up in an [AttributeRegistry] keyed by the
// fully qualified attribute type name (e.g. `std.@secretkey`,
// `posix.@path`).
//
// Both check methods are optional. A nil hook means "the linker does
// not validate this aspect" — useful for purely informational
// attributes like `@since` or `@deprecated` which only affect docs
// and emit no diagnostics.
type AttributeBehaviour interface {
	// StaticCheck is invoked at link time, once per attribute use,
	// against the parsed (still-AST) form of the attribute argument
	// expressions. Used to validate literal arguments before eval
	// runs — e.g. `std.secret("vps.host")` looks up the literal in
	// the secrets backend at link time.
	//
	// args carries the resolved argument expressions in declaration
	// order; positional and named arguments are merged according to
	// the binding rules in lang/check. Names not bound to a positional
	// or default are absent from args.
	//
	// Implementations should emit diagnostics through the supplied
	// LinkContext rather than returning them so the standard pipeline
	// (source spans, --color, --json) handles formatting.
	StaticCheck(ctx LinkContext, args []BoundArg, useSpan spec.SourceSpan)

	// EvalCheck is invoked at eval time, once per attribute use, with
	// the *evaluated* values of the bound arguments. Used to validate
	// dynamic arguments that StaticCheck couldn't reach — e.g.
	// `std.secret("vps." + section)` where the literal isn't known
	// until eval.
	//
	// Same diagnostic-emission contract as StaticCheck.
	EvalCheck(ctx EvalContext, args []EvalArg, useSpan spec.SourceSpan)
}

// LinkContext is the linker-side context passed to an attribute's
// StaticCheck hook. It exposes the secrets backend, source store,
// and diagnostic emitter without requiring the attribute to import
// the entire linker package.
type LinkContext interface {
	// Emit records a diagnostic with the standard pipeline.
	Emit(d diagnostic.Diagnostic)

	// Secrets returns the configured secrets backend, or nil if no
	// backend has been configured. Attribute hooks like `@secretkey`
	// use this to validate literal keys.
	Secrets() SecretBackend
}

// EvalContext is the eval-side context passed to an attribute's
// EvalCheck hook. Mirrors LinkContext but is invoked from inside the
// evaluator at the moment a value gets computed.
type EvalContext interface {
	Emit(d diagnostic.Diagnostic)
	Secrets() SecretBackend
}

// SecretBackend is the minimum surface a secrets backend must expose
// to attribute hooks. Defined here (rather than reusing
// secret.Backend) so the linker package doesn't pull in the full
// secret module just for attribute behaviour.
type SecretBackend interface {
	HasKey(name string) bool
	Keys() []string
}

// BoundArg is a single argument bound to a declared field of an
// attribute type, in its raw AST form. Stage 4 onward will use this
// to inspect literal values before eval runs.
type BoundArg struct {
	Field   string   // declared field name
	Value   ast.Expr // raw expression (may be a literal or a variable)
	SrcSpan spec.SourceSpan
}

// EvalArg is the evaluated counterpart of BoundArg, supplied to
// EvalCheck. The Value is whatever the evaluator produced for the
// argument expression — typically a string, int, bool, or list.
type EvalArg struct {
	Field   string
	Value   any // evaluator-produced value
	SrcSpan spec.SourceSpan
}

// AttributeRegistry holds the AttributeBehaviour for every attribute
// type the linker recognises, keyed by the fully qualified attribute
// name (with the leading `@`, e.g. `std.@secretkey`).
//
// User-defined attribute types declared in third-party scampi modules
// are intentionally absent from this registry — they type-check at
// the lang level but have no runtime behaviour. Future tooling (or a
// future lang-level hook mechanism, see #159 future work) can attach
// behaviour to them.
type AttributeRegistry struct {
	behaviours map[string]AttributeBehaviour
}

// NewAttributeRegistry returns an empty registry. Use Register to
// add behaviours, or DefaultAttributes for the standard set.
func NewAttributeRegistry() *AttributeRegistry {
	return &AttributeRegistry{
		behaviours: make(map[string]AttributeBehaviour),
	}
}

// Register adds a behaviour for the named attribute type. The name
// must be the fully qualified form including the leading `@` (e.g.
// `std.@secretkey`). Subsequent registrations with the same name
// overwrite the previous one.
func (r *AttributeRegistry) Register(qualifiedName string, b AttributeBehaviour) {
	r.behaviours[qualifiedName] = b
}

// Lookup returns the behaviour for the named attribute, or nil if
// none is registered. A nil result means the attribute is inert at
// the linker layer (lang still type-checks it, LSP still consumes it).
func (r *AttributeRegistry) Lookup(qualifiedName string) AttributeBehaviour {
	return r.behaviours[qualifiedName]
}

// Names returns the qualified names of every registered attribute,
// in unspecified order. Useful for diagnostics ("did you mean…?")
// and for tests.
func (r *AttributeRegistry) Names() []string {
	out := make([]string, 0, len(r.behaviours))
	for name := range r.behaviours {
		out = append(out, name)
	}
	return out
}

// DefaultAttributes returns a registry populated with the standard
// scampi attribute behaviours. Stage 4 of #159 wires `@secretkey`;
// Stage 5 adds the rest of the initial set. Until then this returns
// an empty registry — the plumbing is in place, the contents are not.
func DefaultAttributes() *AttributeRegistry {
	return NewAttributeRegistry()
}
