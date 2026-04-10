// SPDX-License-Identifier: GPL-3.0-only

package linker

import (
	"errors"
	"testing"

	"scampi.dev/scampi/lang/ast"
	"scampi.dev/scampi/lang/token"
	"scampi.dev/scampi/spec"
)

// stubBackend is a tiny in-memory secret.Backend for testing.
type stubBackend struct {
	keys     map[string]string
	lookupOK bool
}

func (b *stubBackend) Name() string { return "stub" }

func (b *stubBackend) Lookup(key string) (string, bool, error) {
	if !b.lookupOK {
		return "", false, errors.New("backend offline")
	}
	v, ok := b.keys[key]
	return v, ok, nil
}

func TestSecretKeyAttribute_LiteralFound(t *testing.T) {
	backend := &stubBackend{
		keys:     map[string]string{"db.password": "p4ss"},
		lookupOK: true,
	}
	ctx := &linkContext{backend: backend}
	arg := stringLitExpr("db.password")

	SecretKeyAttribute{}.StaticCheck(
		ctx,
		[]BoundArg{{Field: "name", Value: arg, SrcSpan: spec.SourceSpan{}}},
		spec.SourceSpan{},
	)
	if len(ctx.diags) != 0 {
		t.Errorf("expected no diagnostics for known key, got %d: %v", len(ctx.diags), ctx.diags)
	}
}

func TestSecretKeyAttribute_LiteralNotFound(t *testing.T) {
	backend := &stubBackend{
		keys:     map[string]string{"db.password": "p4ss"},
		lookupOK: true,
	}
	ctx := &linkContext{backend: backend}
	arg := stringLitExpr("totally.unknown")

	SecretKeyAttribute{}.StaticCheck(
		ctx,
		[]BoundArg{{Field: "name", Value: arg, SrcSpan: spec.SourceSpan{}}},
		spec.SourceSpan{},
	)
	if len(ctx.diags) != 1 {
		t.Fatalf("expected 1 diagnostic for unknown key, got %d", len(ctx.diags))
	}
	if _, ok := ctx.diags[0].(*secretKeyNotFoundError); !ok {
		t.Errorf("expected *secretKeyNotFoundError, got %T", ctx.diags[0])
	}
}

func TestSecretKeyAttribute_ComputedArgSkipped(t *testing.T) {
	// A non-literal expression should be skipped — the runtime check
	// handles dynamic args in lang/eval.
	backend := &stubBackend{lookupOK: true}
	ctx := &linkContext{backend: backend}
	arg := &ast.Ident{Name: "some_var", SrcSpan: token.Span{Start: 1, End: 9}}

	SecretKeyAttribute{}.StaticCheck(
		ctx,
		[]BoundArg{{Field: "name", Value: arg, SrcSpan: spec.SourceSpan{}}},
		spec.SourceSpan{},
	)
	if len(ctx.diags) != 0 {
		t.Errorf("expected no diagnostics for computed arg, got %d", len(ctx.diags))
	}
}

func TestSecretKeyAttribute_NoBackendSkipped(t *testing.T) {
	// With no backend configured, the static check should be a no-op
	// and let the runtime check handle it.
	ctx := &linkContext{backend: nil}
	arg := stringLitExpr("any.key")

	SecretKeyAttribute{}.StaticCheck(
		ctx,
		[]BoundArg{{Field: "name", Value: arg, SrcSpan: spec.SourceSpan{}}},
		spec.SourceSpan{},
	)
	if len(ctx.diags) != 0 {
		t.Errorf("expected no diagnostics with nil backend, got %d", len(ctx.diags))
	}
}

func TestSecretKeyAttribute_LookupError(t *testing.T) {
	backend := &stubBackend{lookupOK: false}
	ctx := &linkContext{backend: backend}
	arg := stringLitExpr("db.password")

	SecretKeyAttribute{}.StaticCheck(
		ctx,
		[]BoundArg{{Field: "name", Value: arg, SrcSpan: spec.SourceSpan{}}},
		spec.SourceSpan{},
	)
	if len(ctx.diags) != 1 {
		t.Fatalf("expected 1 diagnostic for backend error, got %d", len(ctx.diags))
	}
	if _, ok := ctx.diags[0].(*secretKeyLookupError); !ok {
		t.Errorf("expected *secretKeyLookupError, got %T", ctx.diags[0])
	}
}

// stringLitExpr builds a single-segment string literal AST node for
// tests, mirroring what the parser produces for `"value"`.
func stringLitExpr(value string) *ast.StringLit {
	return &ast.StringLit{
		Parts: []ast.StringPart{
			&ast.StringText{Raw: value, SrcSpan: token.Span{Start: 1, End: uint32(1 + len(value))}},
		},
		SrcSpan: token.Span{Start: 0, End: uint32(2 + len(value))},
	}
}
