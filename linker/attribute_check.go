// SPDX-License-Identifier: GPL-3.0-only

package linker

import (
	"scampi.dev/scampi/diagnostic"
	"scampi.dev/scampi/lang/ast"
	"scampi.dev/scampi/lang/check"
	"scampi.dev/scampi/secret"
	"scampi.dev/scampi/spec"
)

// runAttributeStaticChecks walks the parsed file looking for call
// sites of functions whose parameters carry `@`-attributes. For each
// such call site it dispatches the corresponding registered
// AttributeBehaviour with the literal argument expressions, allowing
// behaviours to validate literal values before plan/apply runs.
//
// Function calls whose target cannot be resolved (e.g. user-defined
// helpers without registered behaviour) are skipped silently — the
// type checker has already validated them at the lang layer.
func runAttributeStaticChecks(
	f *ast.File,
	source []byte,
	cfgPath string,
	fileScope *check.Scope,
	modules map[string]*check.Scope,
	registry *AttributeRegistry,
	backend secret.Backend,
) error {
	if registry == nil {
		return nil
	}
	ctx := &linkContext{
		backend: backend,
	}
	visitor := &attributeCheckVisitor{
		ctx:       ctx,
		registry:  registry,
		fileScope: fileScope,
		modules:   modules,
		source:    source,
		cfgPath:   cfgPath,
	}
	ast.Walk(f, visitor.enter, nil)
	if len(ctx.diags) == 0 {
		return nil
	}
	return ctx.diags
}

// attributeCheckVisitor walks the AST and dispatches attribute
// behaviours for each annotated call site.
type attributeCheckVisitor struct {
	ctx       *linkContext
	registry  *AttributeRegistry
	fileScope *check.Scope
	modules   map[string]*check.Scope
	source    []byte
	cfgPath   string
}

func (v *attributeCheckVisitor) enter(n ast.Node) bool {
	call, ok := n.(*ast.CallExpr)
	if !ok {
		return true
	}
	ft := v.resolveCallTarget(call.Fn)
	if ft == nil {
		return true
	}
	v.checkCall(call, ft)
	return true
}

// resolveCallTarget walks the call's function expression and tries to
// look up the resulting symbol's FuncType. Currently handles two
// shapes: a bare identifier (resolves in the file scope) and a
// dotted name like `std.secret` (resolves in the imported module's
// scope). Returns nil if the function isn't a known annotated target.
func (v *attributeCheckVisitor) resolveCallTarget(e ast.Expr) *check.FuncType {
	switch fn := e.(type) {
	case *ast.Ident:
		if v.fileScope == nil {
			return nil
		}
		sym := v.fileScope.Lookup(fn.Name)
		if sym == nil {
			return nil
		}
		ft, _ := sym.Type.(*check.FuncType)
		return ft
	case *ast.SelectorExpr:
		modIdent, ok := fn.X.(*ast.Ident)
		if !ok {
			return nil
		}
		mod := v.modules[modIdent.Name]
		if mod == nil {
			return nil
		}
		sym := mod.Lookup(fn.Sel.Name)
		if sym == nil {
			return nil
		}
		ft, _ := sym.Type.(*check.FuncType)
		return ft
	}
	return nil
}

// checkCall iterates the parameters of the resolved function and, for
// each parameter that carries an attribute, dispatches the registered
// behaviour with the corresponding argument expression.
func (v *attributeCheckVisitor) checkCall(call *ast.CallExpr, ft *check.FuncType) {
	if len(ft.Params) == 0 {
		return
	}
	// Build a map: param index → arg expression. The lang's
	// CallExpr.Args entries are either positional (no Name) or
	// keyword (Name=value). We need to bind them back to params.
	argFor := bindCallArgs(call, ft)
	for i, p := range ft.Params {
		if len(p.Attributes) == 0 {
			continue
		}
		argExpr, ok := argFor[i]
		if !ok {
			continue
		}
		for _, attrName := range p.Attributes {
			behaviour := v.registry.Lookup(attrName)
			if behaviour == nil {
				continue
			}
			useSpan := nodeSourceSpan(argExpr, v.source, v.cfgPath)
			behaviour.StaticCheck(v.ctx, []BoundArg{
				{
					Field:   p.Name,
					Value:   argExpr,
					SrcSpan: useSpan,
				},
			}, useSpan)
		}
	}
}

// bindCallArgs maps parameter indices to argument expressions for a
// CallExpr. Positional arguments bind to the leading parameters in
// order; keyword arguments bind by name. Parameters with no
// corresponding argument (e.g. optionals using their default) are
// absent from the result.
func bindCallArgs(call *ast.CallExpr, ft *check.FuncType) map[int]ast.Expr {
	out := make(map[int]ast.Expr, len(call.Args))
	posIdx := 0
	for _, a := range call.Args {
		if a.Name == nil {
			if posIdx < len(ft.Params) {
				out[posIdx] = a.Value
			}
			posIdx++
			continue
		}
		for i, p := range ft.Params {
			if p.Name == a.Name.Name {
				out[i] = a.Value
				break
			}
		}
	}
	return out
}

// linkContext is the linker-side LinkContext implementation passed
// to AttributeBehaviour.StaticCheck. It collects diagnostics emitted
// during the static check pass; the caller wraps them into a single
// diagnostic.Diagnostics for return through the standard pipeline.
type linkContext struct {
	backend secret.Backend
	diags   diagnostic.Diagnostics
}

func (lc *linkContext) Emit(d diagnostic.Diagnostic) {
	lc.diags = append(lc.diags, d)
}

func (lc *linkContext) Secrets() secret.Backend {
	return lc.backend
}

// nodeSourceSpan converts an AST node's token.Span into the
// spec.SourceSpan shape used by the diagnostic pipeline, resolving
// byte offsets to line/column via the source bytes.
func nodeSourceSpan(node ast.Node, source []byte, cfgPath string) spec.SourceSpan {
	span := node.Span()
	startLine, startCol := offsetToLineCol(source, int(span.Start))
	endLine, endCol := offsetToLineCol(source, int(span.End))
	return spec.SourceSpan{
		Filename:  cfgPath,
		StartLine: startLine,
		StartCol:  startCol,
		EndLine:   endLine,
		EndCol:    endCol,
	}
}
