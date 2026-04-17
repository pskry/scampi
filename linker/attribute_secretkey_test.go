// SPDX-License-Identifier: GPL-3.0-only

package linker

import (
	"errors"
	"testing"

	"scampi.dev/scampi/lang/ast"
	"scampi.dev/scampi/lang/token"
	"scampi.dev/scampi/secret"
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

func (b *stubBackend) Keys() []string { return secret.SortedKeys(b.keys) }

func newSecretCtx(backend *stubBackend, arg ast.Expr) StaticCheckContext {
	// Wrap nil typed pointer as a true nil interface so the
	// "no backend configured" path actually fires.
	var b secret.Backend
	if backend != nil {
		b = backend
	}
	return StaticCheckContext{
		Linker:          &linkContext{},
		ResolverBackend: b,
		AttrName:        "secrets.@secretkey",
		ParamName:       "key",
		ParamArg:        arg,
		UseSpan:         spec.SourceSpan{},
	}
}

func TestSecretKeyAttribute_LiteralFound(t *testing.T) {
	backend := &stubBackend{
		keys:     map[string]string{"db.password": "p4ss"},
		lookupOK: true,
	}
	ctx := newSecretCtx(backend, stringLitExpr("db.password"))

	SecretKeyAttribute{}.StaticCheck(ctx)
	lc := ctx.Linker.(*linkContext)
	if len(lc.diags) != 0 {
		t.Errorf("expected no diagnostics for known key, got %d: %v", len(lc.diags), lc.diags)
	}
}

func TestSecretKeyAttribute_LiteralNotFound(t *testing.T) {
	backend := &stubBackend{
		keys:     map[string]string{"db.password": "p4ss"},
		lookupOK: true,
	}
	ctx := newSecretCtx(backend, stringLitExpr("totally.unknown"))

	SecretKeyAttribute{}.StaticCheck(ctx)
	lc := ctx.Linker.(*linkContext)
	if len(lc.diags) != 1 {
		t.Fatalf("expected 1 diagnostic for unknown key, got %d", len(lc.diags))
	}
	if _, ok := lc.diags[0].(*secretKeyNotFoundError); !ok {
		t.Errorf("expected *secretKeyNotFoundError, got %T", lc.diags[0])
	}
}

func TestSecretKeyAttribute_ComputedArgSkipped(t *testing.T) {
	// A non-literal expression should be skipped — the runtime check
	// handles dynamic args in lang/eval.
	backend := &stubBackend{lookupOK: true}
	arg := &ast.Ident{Name: "some_var", SrcSpan: token.Span{Start: 1, End: 9}}
	ctx := newSecretCtx(backend, arg)

	SecretKeyAttribute{}.StaticCheck(ctx)
	lc := ctx.Linker.(*linkContext)
	if len(lc.diags) != 0 {
		t.Errorf("expected no diagnostics for computed arg, got %d", len(lc.diags))
	}
}

func TestSecretKeyAttribute_NoBackendSkipped(t *testing.T) {
	// With no backend configured, the static check should be a no-op
	// and let the runtime check handle it.
	ctx := newSecretCtx(nil, stringLitExpr("any.key"))

	SecretKeyAttribute{}.StaticCheck(ctx)
	lc := ctx.Linker.(*linkContext)
	if len(lc.diags) != 0 {
		t.Errorf("expected no diagnostics with nil backend, got %d", len(lc.diags))
	}
}

func TestSecretKeyAttribute_LookupError(t *testing.T) {
	backend := &stubBackend{lookupOK: false}
	ctx := newSecretCtx(backend, stringLitExpr("db.password"))

	SecretKeyAttribute{}.StaticCheck(ctx)
	lc := ctx.Linker.(*linkContext)
	if len(lc.diags) != 1 {
		t.Fatalf("expected 1 diagnostic for backend error, got %d", len(lc.diags))
	}
	if _, ok := lc.diags[0].(*secretKeyLookupError); !ok {
		t.Errorf("expected *secretKeyLookupError, got %T", lc.diags[0])
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
