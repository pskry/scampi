// SPDX-License-Identifier: GPL-3.0-only

package lsp

import (
	"context"
	"fmt"
	"strings"

	"go.lsp.dev/protocol"
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

	return nil, nil
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
