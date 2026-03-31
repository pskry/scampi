// SPDX-License-Identifier: GPL-3.0-only

package lsp

import (
	"context"
	"fmt"
	"strings"

	"go.lsp.dev/protocol"
)

func (s *Server) Completion(
	_ context.Context,
	params *protocol.CompletionParams,
) (*protocol.CompletionList, error) {
	doc, ok := s.docs.Get(params.TextDocument.URI)
	if !ok {
		s.log.Printf("completion: no doc for %s", params.TextDocument.URI)
		return nil, nil
	}

	cur := AnalyzeCursor(doc.Content, params.Position.Line, params.Position.Character)
	s.log.Printf(
		"completion: line=%d col=%d word=%q inCall=%v inList=%v func=%q",
		params.Position.Line,
		params.Position.Character,
		cur.WordUnderCursor,
		cur.InCall,
		cur.InList,
		cur.FuncName,
	)

	var items []protocol.CompletionItem

	switch {
	case cur.InList:
		items = s.completeTopLevel(cur.WordUnderCursor)
	case cur.InCall:
		items = s.completeKwargs(cur)
	case isDotPrefix(cur.WordUnderCursor):
		items = s.completeModule(cur.WordUnderCursor)
	default:
		items = s.completeTopLevel(cur.WordUnderCursor)
	}

	s.log.Printf("completion: returning %d items", len(items))
	return &protocol.CompletionList{
		IsIncomplete: false,
		Items:        items,
	}, nil
}

// completeTopLevel offers all top-level builtins (non-dotted and module names).
func (s *Server) completeTopLevel(prefix string) []protocol.CompletionItem {
	var items []protocol.CompletionItem

	for _, name := range s.catalog.Names() {
		// Skip dotted names at top level — offer the module name instead.
		if strings.Contains(name, ".") {
			continue
		}
		if prefix != "" && !strings.HasPrefix(name, prefix) {
			continue
		}

		f, _ := s.catalog.Lookup(name)
		kind := protocol.CompletionItemKindFunction
		items = append(items, protocol.CompletionItem{
			Label:         name,
			Kind:          kind,
			Detail:        f.Summary,
			InsertText:    name + "(",
			Documentation: f.Summary,
		})
	}

	// Offer module names.
	for _, mod := range s.catalog.Modules() {
		if prefix != "" && !strings.HasPrefix(mod, prefix) {
			continue
		}
		items = append(items, protocol.CompletionItem{
			Label:      mod,
			Kind:       protocol.CompletionItemKindModule,
			Detail:     mod + " namespace",
			InsertText: mod + ".",
		})
	}

	return items
}

// completeModule offers members of a dotted module prefix (e.g. "target." → "ssh", "local", "rest").
func (s *Server) completeModule(word string) []protocol.CompletionItem {
	dot := strings.LastIndexByte(word, '.')
	if dot < 0 {
		return nil
	}
	mod := word[:dot]
	prefix := word[dot+1:]

	members := s.catalog.ModuleMembers(mod)
	var items []protocol.CompletionItem
	for _, member := range members {
		if prefix != "" && !strings.HasPrefix(member, prefix) {
			continue
		}
		fullName := mod + "." + member
		f, ok := s.catalog.Lookup(fullName)
		if !ok {
			continue
		}
		items = append(items, protocol.CompletionItem{
			Label:         member,
			Kind:          protocol.CompletionItemKindFunction,
			Detail:        f.Summary,
			InsertText:    member + "(",
			Documentation: f.Summary,
		})
	}
	return items
}

// completeKwargs offers keyword arguments for the function being called.
func (s *Server) completeKwargs(cur CursorContext) []protocol.CompletionItem {
	f, ok := s.catalog.Lookup(cur.FuncName)
	if !ok {
		return nil
	}

	present := make(map[string]bool, len(cur.PresentKwargs))
	for _, k := range cur.PresentKwargs {
		present[k] = true
	}

	var items []protocol.CompletionItem
	for _, p := range f.Params {
		if present[p.Name] {
			continue
		}

		detail := p.Type
		if p.Required {
			detail += " (required)"
		}

		doc := p.Desc
		if p.Default != "" {
			doc += fmt.Sprintf("\n\nDefault: %s", p.Default)
		}
		if len(p.Examples) > 0 {
			doc += fmt.Sprintf("\n\nExamples: %s", strings.Join(p.Examples, ", "))
		}

		items = append(items, protocol.CompletionItem{
			Label:         p.Name,
			Kind:          protocol.CompletionItemKindProperty,
			Detail:        detail,
			InsertText:    p.Name + "=",
			Documentation: doc,
		})
	}
	return items
}

func isDotPrefix(word string) bool {
	return strings.Contains(word, ".")
}
