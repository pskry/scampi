// SPDX-License-Identifier: GPL-3.0-only

package linker

import (
	"scampi.dev/scampi/diagnostic"
	"scampi.dev/scampi/lang/ast"
	"scampi.dev/scampi/secret"
	"scampi.dev/scampi/spec"
)

// AttributeBehaviour is the scampi-specific semantics attached to an
// attribute type. Lang owns the schema (declared via `type @name { ... }`
// in stubs); the linker owns what an attribute *means* at link time.
// The LSP separately owns UX semantics (completion, hover) keyed by
// the same qualified name.
//
// Implementations are looked up in an [AttributeRegistry] keyed by the
// fully qualified attribute type name (e.g. `std.@secretkey`,
// `posix.@path`).
//
// Eval-time checks for dynamic argument values are not yet wired
// (#159 follow-up). For now the existing per-builtin runtime checks
// in lang/eval continue to handle dynamic args; StaticCheck handles
// the literal-args case.
type AttributeBehaviour interface {
	// StaticCheck is invoked at link time, once per call site of a
	// function whose parameter carries this attribute. args carries
	// the bound arguments in declaration order — typically a single
	// entry holding the user's call-site argument expression for the
	// annotated parameter.
	//
	// Implementations should emit diagnostics through the supplied
	// LinkContext rather than returning them so the standard pipeline
	// (source spans, --color, --json) handles formatting.
	StaticCheck(ctx LinkContext, args []BoundArg, useSpan spec.SourceSpan)
}

// LinkContext is the linker-side context passed to an attribute's
// StaticCheck hook. It exposes the secrets backend and a diagnostic
// emitter without requiring the attribute to import the entire
// linker package.
type LinkContext interface {
	// Emit records a diagnostic with the standard pipeline.
	Emit(d diagnostic.Diagnostic)

	// Secrets returns the configured secrets backend, or nil if no
	// backend has been configured. Attribute hooks like `@secretkey`
	// use this to validate literal keys against known entries.
	Secrets() secret.Backend
}

// BoundArg is a single argument bound to a declared field of an
// attribute type, in its raw AST form. Behaviours inspect the
// expression to detect literal values they can validate eagerly.
type BoundArg struct {
	Field   string   // declared field name
	Value   ast.Expr // raw expression (may be a literal or a variable)
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
// scampi attribute behaviours.
func DefaultAttributes() *AttributeRegistry {
	r := NewAttributeRegistry()
	r.Register("std.@secretkey", SecretKeyAttribute{})
	return r
}
