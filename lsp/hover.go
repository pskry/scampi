// SPDX-License-Identifier: GPL-3.0-only

package lsp

import (
	"context"
	"fmt"
	"strings"

	"go.lsp.dev/protocol"

	"scampi.dev/scampi/lang/check"
)

func (s *Server) Hover(
	_ context.Context,
	params *protocol.HoverParams,
) (*protocol.Hover, error) {
	doc, ok := s.docs.Get(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}

	cur := AnalyzeCursor(doc.Content, params.Position.Line, params.Position.Character)
	s.log.Printf(
		"hover: line=%d col=%d word=%q inCall=%v func=%q",
		params.Position.Line,
		params.Position.Character,
		cur.WordUnderCursor,
		cur.InCall,
		cur.FuncName,
	)

	// Attribute reference (e.g. `@secretkey`, `@std.path`) — show the
	// attribute type's schema.
	if len(cur.WordUnderCursor) > 1 && cur.WordUnderCursor[0] == '@' {
		if md := s.hoverAttribute(cur.WordUnderCursor); md != "" {
			s.log.Printf("hover: returning attribute doc (%d bytes)", len(md))
			return &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.Markdown,
					Value: md,
				},
			}, nil
		}
	}

	// Known function name always wins — handles nested calls like
	// deploy(steps=[copy(...)]) where copy is inside deploy's parens.
	if md := s.hoverFunc(params.TextDocument.URI, cur.WordUnderCursor); md != "" {
		s.log.Printf("hover: returning func doc (%d bytes), kind=%q\n---\n%s\n---", len(md), protocol.Markdown, md)
		return &protocol.Hover{
			Contents: protocol.MarkupContent{
				Kind:  protocol.Markdown,
				Value: md,
			},
		}, nil
	}

	// Kwarg name inside a call?
	if cur.InCall {
		if md := s.hoverKwarg(params.TextDocument.URI, cur); md != "" {
			s.log.Printf("hover: returning kwarg doc (%d bytes), kind=%q", len(md), protocol.Markdown)
			return &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.Markdown,
					Value: md,
				},
			}, nil
		}
	}

	// Local symbol (let, type, enum, param) — checker-driven fallback.
	// Runs after the function/kwarg paths so a name that exists both as
	// a stdlib func and a local let still prefers the func doc.
	if md := s.hoverLocalSymbol(params.TextDocument.URI, doc.Content, cur.WordUnderCursor); md != "" {
		s.log.Printf("hover: returning symbol doc (%d bytes)", len(md))
		return &protocol.Hover{
			Contents: protocol.MarkupContent{
				Kind:  protocol.Markdown,
				Value: md,
			},
		}, nil
	}

	// Stub-defined type (builtins, std types) — last resort.
	if stubDoc, ok := s.stubDefs.LookupDoc(cur.WordUnderCursor); ok && stubDoc != "" {
		md := fmt.Sprintf("```scampi\ntype %s\n```\n\n%s", cur.WordUnderCursor, stubDoc)
		s.log.Printf("hover: returning stub type doc (%d bytes)", len(md))
		return &protocol.Hover{
			Contents: protocol.MarkupContent{
				Kind:  protocol.Markdown,
				Value: md,
			},
		}, nil
	}

	return nil, nil
}

// hoverLocalSymbol resolves a name against the file's checker scope
// and renders a hover doc for lets, params, types, and enums. Falls
// through cleanly when the name doesn't resolve. Handles dotted words
// like `x.yo` by trying the trailing segment as well.
func (s *Server) hoverLocalSymbol(docURI protocol.DocumentURI, content, word string) string {
	if word == "" {
		return ""
	}
	filePath := uriToPath(docURI)
	c := tolerantCheck(filePath, []byte(content), s.modules)
	if c == nil {
		return ""
	}
	if sym := lookupSymbol(c, word); sym != nil {
		md := formatSymbolDoc(sym)
		if sym.Kind == check.SymEnum {
			if doc := s.enumDoc(word); doc != "" {
				md += doc
			}
		}
		if md != "" {
			return md
		}
	}
	if i := strings.LastIndexByte(word, '.'); i >= 0 && i < len(word)-1 {
		prefix := word[:i]
		variant := word[i+1:]
		// Try local scope first, then imported module scopes.
		sym := c.FileScope().Lookup(prefix)
		if sym == nil {
			for _, imp := range c.FileScope().AllImports() {
				if modScope, ok := s.modules[imp.Name]; ok {
					if s := modScope.Lookup(prefix); s != nil {
						sym = s
						break
					}
				}
			}
		}
		if sym != nil && sym.Kind == check.SymEnum {
			if et, ok := sym.Type.(*check.EnumType); ok {
				for _, v := range et.Variants {
					if v == variant {
						sig := "type " + prefix + "." + variant
						doc := s.enumDoc(prefix)
						if doc != "" {
							return fencedSymbolDoc(sig) + doc
						}
						return fencedSymbolDoc(sig)
					}
				}
			}
		}
		if sym := lookupSymbol(c, variant); sym != nil {
			return formatSymbolDoc(sym)
		}
	}
	return ""
}

func lookupSymbol(c *check.Checker, name string) *check.Symbol {
	if sym := c.FileScope().Lookup(name); sym != nil {
		return sym
	}
	if sym, ok := c.AllBindings()[name]; ok {
		return sym
	}
	return nil
}

// formatSymbolDoc renders a hover doc for a checker symbol. Funcs are
// intentionally skipped — function hovers go through the catalog path
// in hoverFunc, which has richer params/docs from the FuncInfo
// representation.
func formatSymbolDoc(sym *check.Symbol) string {
	switch sym.Kind {
	case check.SymLet:
		return fencedSymbolDoc("let " + sym.Name + typeSuffix(sym))
	case check.SymParam:
		return fencedSymbolDoc(sym.Name + typeSuffix(sym))
	case check.SymType:
		return formatTypeDoc(sym)
	case check.SymEnum:
		return fencedSymbolDoc("enum " + sym.Name)
	}
	return ""
}

// formatTypeDoc renders hover documentation for a user-defined type.
// For struct types, shows the field listing with types and
// required/optional status.
func formatTypeDoc(sym *check.Symbol) string {
	st, ok := sym.Type.(*check.StructType)
	if !ok || len(st.Fields) == 0 {
		return fencedSymbolDoc("type " + sym.Name)
	}

	var b strings.Builder
	b.WriteString("```scampi\ntype " + sym.Name + " {\n")
	nameW := 0
	for _, f := range st.Fields {
		if l := len(f.Name); l > nameW {
			nameW = l
		}
	}
	for _, f := range st.Fields {
		req := ""
		if f.HasDef {
			req = "  // has default"
		}
		_, _ = fmt.Fprintf(&b, "  %-*s  %s%s\n", nameW, f.Name+":", f.Type.String(), req)
	}
	b.WriteString("}\n```\n\n---\n")
	return b.String()
}

func typeSuffix(sym *check.Symbol) string {
	if sym.Type == nil {
		return ""
	}
	return ": " + sym.Type.String()
}

func (s *Server) enumDoc(enumName string) string {
	if doc, ok := s.stubDefs.LookupDoc(enumName); ok && doc != "" {
		return doc
	}
	for _, mod := range s.catalog.Modules() {
		if doc, ok := s.stubDefs.LookupDoc(mod + "." + enumName); ok && doc != "" {
			return doc
		}
	}
	return ""
}

func fencedSymbolDoc(sig string) string {
	return "```scampi\n" + sig + "\n```\n\n---\n"
}

func (s *Server) hoverFunc(docURI protocol.DocumentURI, word string) string {
	f, ok := s.lookupFunc(docURI, word)
	if !ok {
		return ""
	}
	return formatFuncDoc(f)
}

// hoverAttribute renders the documentation for an attribute type
// reference. The word includes the leading `@`. Looks up the type in
// the catalog and renders its qualified name plus its schema fields
// (or "marker" for empty attribute types).
func (s *Server) hoverAttribute(word string) string {
	a, ok := s.catalog.LookupAttrType(word)
	if !ok {
		return ""
	}
	return formatAttrTypeDoc(a)
}

func formatAttrTypeDoc(a AttrTypeInfo) string {
	var b strings.Builder
	qname := "@" + a.Module + "." + a.Name
	b.WriteString("```scampi\ntype " + qname + " { ... }\n```\n\n---\n\n")
	if a.Summary != "" {
		b.WriteString(a.Summary + "\n\n")
	}
	if len(a.Fields) == 0 {
		b.WriteString("_marker attribute — no fields_\n")
		b.WriteString("\n---\n")
		return b.String()
	}
	nameW, typeW := 0, 0
	for _, p := range a.Fields {
		if l := len(p.Name); l > nameW {
			nameW = l
		}
		if l := len(p.Type); l > typeW {
			typeW = l
		}
	}
	b.WriteString("```\n")
	for _, p := range a.Fields {
		req := ""
		if p.Required {
			req = "required"
		}
		line := strings.TrimRight(
			padRight(p.Name, nameW)+"  "+padRight(p.Type, typeW)+"  "+padRight(req, 8),
			" ",
		)
		if p.Default != "" {
			line += "  (default: " + p.Default + ")"
		}
		b.WriteString(line + "\n")
	}
	b.WriteString("```\n\n---\n")
	return b.String()
}

// padRight returns s padded with spaces on the right to width w.
func padRight(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-len(s))
}

func (s *Server) hoverKwarg(docURI protocol.DocumentURI, cur CursorContext) string {
	f, ok := s.lookupFunc(docURI, cur.FuncName)
	if !ok {
		return ""
	}

	// Check if the word under cursor matches a param name.
	for _, p := range f.Params {
		if p.Name == cur.WordUnderCursor {
			return formatParamDoc(cur.FuncName, p)
		}
	}
	return ""
}

func formatFuncDoc(f FuncInfo) string {
	var b strings.Builder

	// Signature in a fenced code block, like gopls. Followed by a horizontal
	// rule that visually separates the signature from the description and
	// parameter list.
	b.WriteString("```scampi\n" + formatSignature(f) + "\n```\n\n---\n\n")

	if f.Summary != "" {
		b.WriteString(f.Summary + "\n\n")
	}

	if len(f.Params) > 0 {
		b.WriteString(formatParamTable(f.Params))
	}

	// Trailing horizontal rule + blank line so the floating window has visual
	// closure at the bottom rather than touching content to the border.
	b.WriteString("\n---\n")
	return b.String()
}

// formatParamTable renders the parameter list as an aligned monospace block
// inside a fenced code block. Columns are name, type, required-marker, then
// description (with default and examples appended inline). Code-block content
// is rendered in a monospace font in editors that support markdown hovers, so
// the spaces here line up properly — unlike a markdown table, which renders
// poorly in many editors (notably Neovim's floating windows).
func formatParamTable(params []ParamInfo) string {
	nameW, typeW := 0, 0
	for _, p := range params {
		if l := len(p.Name); l > nameW {
			nameW = l
		}
		if l := len(p.Type); l > typeW {
			typeW = l
		}
	}

	var b strings.Builder
	b.WriteString("```\n")
	for _, p := range params {
		req := ""
		if p.Required {
			req = "required"
		}
		line := fmt.Sprintf("%-*s  %-*s  %-8s", nameW, p.Name, typeW, p.Type, req)
		// Append description and metadata. If neither column 3 nor any trailing
		// info exists, strip the trailing padding so the line ends cleanly.
		extra := p.Desc
		if p.Default != "" {
			if extra != "" {
				extra += " "
			}
			extra += "(default: " + p.Default + ")"
		}
		if len(p.Examples) > 0 {
			if extra != "" {
				extra += " "
			}
			extra += "(e.g. " + strings.Join(p.Examples, ", ") + ")"
		}
		if extra != "" {
			line += "  " + extra
		}
		line = strings.TrimRight(line, " ")
		b.WriteString(line + "\n")
	}
	b.WriteString("```\n")
	return b.String()
}

func formatSignature(f FuncInfo) string {
	var params []string
	for _, p := range f.Params {
		s := p.Name
		if !p.Required {
			s += "?"
		}
		params = append(params, s)
	}
	return f.Name + "(" + strings.Join(params, ", ") + ")"
}

func formatParamDoc(funcName string, p ParamInfo) string {
	var b strings.Builder

	// Signature block: shows the param in context of its function.
	req := "optional"
	if p.Required {
		req = "required"
	}
	b.WriteString("```scampi\n")
	b.WriteString(funcName + "(" + p.Name + ": " + p.Type + ")  // " + req + "\n")
	b.WriteString("```\n\n")

	if p.Desc != "" {
		b.WriteString(p.Desc + "\n\n")
	}

	if p.Default != "" {
		b.WriteString("**Default:** `" + p.Default + "`\n\n")
	}
	if len(p.Examples) > 0 {
		b.WriteString("**Examples:** `" + strings.Join(p.Examples, "`, `") + "`\n")
	}

	return strings.TrimRight(b.String(), "\n") + "\n"
}
