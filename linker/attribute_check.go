// SPDX-License-Identifier: GPL-3.0-only

package linker

import (
	"strings"

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
	switch n := n.(type) {
	case *ast.CallExpr:
		if ft := v.resolveCallTarget(n.Fn); ft != nil {
			v.checkCall(n, ft)
		}
	case *ast.StructLit:
		// Decl invocations like `posix.copy { ... }` parse as struct
		// literals with a typed dotted name. Look up the dotted name
		// in the modules map and dispatch attribute behaviours
		// against the field initializers.
		if dt := v.resolveStructLitDecl(n); dt != nil {
			v.checkStructLit(n, dt)
		}
	}
	return true
}

// resolveStructLitDecl looks up the DeclType for a struct literal's
// dotted type name. Returns nil if the struct literal is anonymous,
// not declared via `decl ...`, or otherwise not relevant to attribute
// dispatch.
func (v *attributeCheckVisitor) resolveStructLitDecl(sl *ast.StructLit) *check.DeclType {
	if sl.Type == nil {
		return nil
	}
	nt, ok := sl.Type.(*ast.NamedType)
	if !ok {
		return nil
	}
	parts := nt.Name.Parts
	switch len(parts) {
	case 1:
		if v.fileScope == nil {
			return nil
		}
		sym := v.fileScope.Lookup(parts[0].Name)
		if sym == nil {
			return nil
		}
		dt, _ := sym.Type.(*check.DeclType)
		return dt
	case 2:
		mod := v.modules[parts[0].Name]
		if mod == nil {
			return nil
		}
		sym := mod.Lookup(parts[1].Name)
		if sym == nil {
			return nil
		}
		dt, _ := sym.Type.(*check.DeclType)
		return dt
	}
	return nil
}

// checkStructLit dispatches attribute behaviours for the parameters
// of a decl invocation written as a struct literal. Field
// initializers bind by name (struct literals don't have positional
// args).
func (v *attributeCheckVisitor) checkStructLit(sl *ast.StructLit, dt *check.DeclType) {
	if len(dt.Params) == 0 {
		return
	}
	byName := make(map[string]ast.Expr, len(sl.Fields))
	for _, fi := range sl.Fields {
		byName[fi.Name.Name] = fi.Value
	}
	for _, p := range dt.Params {
		if len(p.Attributes) == 0 {
			continue
		}
		argExpr, ok := byName[p.Name]
		if !ok {
			continue
		}
		for _, attr := range p.Attributes {
			behaviour := v.registry.Lookup(attr.QualifiedName)
			if behaviour == nil {
				continue
			}
			useSpan := nodeSourceSpan(argExpr, v.source, v.cfgPath)
			behaviour.StaticCheck(StaticCheckContext{
				Linker:    v.ctx,
				AttrName:  attr.QualifiedName,
				AttrArgs:  attr.Args,
				AttrDoc:   v.attrDocFor(attr.QualifiedName),
				ParamName: p.Name,
				ParamArg:  argExpr,
				UseSpan:   useSpan,
			})
		}
	}
}

// attrDocFor returns the doc-comment block of the attribute type
// declared at the given qualified name (e.g. `std.@secretkey`).
// Looks up the AttrType in the modules map; returns "" when the
// declaring module isn't loaded or when the type carries no doc.
func (v *attributeCheckVisitor) attrDocFor(qualifiedName string) string {
	// Qualified name shape: `<module>.@<name>`. Split into module
	// and bare name. Anything that doesn't match the shape returns "".
	atIdx := strings.IndexByte(qualifiedName, '@')
	if atIdx <= 0 || qualifiedName[atIdx-1] != '.' {
		return ""
	}
	modName := qualifiedName[:atIdx-1]
	bare := qualifiedName[atIdx+1:]
	mod, ok := v.modules[modName]
	if !ok {
		return ""
	}
	sym := mod.Lookup("@" + bare)
	if sym == nil || sym.Kind != check.SymAttrType {
		return ""
	}
	at, ok := sym.Type.(*check.AttrType)
	if !ok {
		return ""
	}
	return at.Doc
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
// behaviour with the corresponding argument expression. The
// ResolvedAttribute on the FieldDef carries the attribute's own
// resolved literal arguments — behaviours like `@pattern(regex)`
// read them via ctx.AttrArgs without needing AST access.
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
		for _, attr := range p.Attributes {
			behaviour := v.registry.Lookup(attr.QualifiedName)
			if behaviour == nil {
				continue
			}
			useSpan := nodeSourceSpan(argExpr, v.source, v.cfgPath)
			behaviour.StaticCheck(StaticCheckContext{
				Linker:    v.ctx,
				AttrName:  attr.QualifiedName,
				AttrArgs:  attr.Args,
				AttrDoc:   v.attrDocFor(attr.QualifiedName),
				ParamName: p.Name,
				ParamArg:  argExpr,
				UseSpan:   useSpan,
			})
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
