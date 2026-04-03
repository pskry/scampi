// SPDX-License-Identifier: GPL-3.0-only

package lsp

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"go.starlark.net/syntax"
)

func (s *Server) DocumentSymbol(
	_ context.Context,
	params *protocol.DocumentSymbolParams,
) ([]any, error) {
	doc, ok := s.docs.Get(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}

	filePath := uriToPath(params.TextDocument.URI)
	f, _ := Parse(filePath, []byte(doc.Content))
	if f == nil {
		return nil, nil
	}

	var symbols []any
	for _, stmt := range f.Stmts {
		switch st := stmt.(type) {
		case *syntax.DefStmt:
			start, end := st.Span()
			symbols = append(symbols, protocol.DocumentSymbol{
				Name:           st.Name.Name,
				Kind:           protocol.SymbolKindFunction,
				Range:          spanToSymbolRange(start, end),
				SelectionRange: posToLSPRange(st.Name.NamePos),
			})
		case *syntax.AssignStmt:
			if ident, ok := st.LHS.(*syntax.Ident); ok {
				start, end := st.Span()
				symbols = append(symbols, protocol.DocumentSymbol{
					Name:           ident.Name,
					Kind:           protocol.SymbolKindVariable,
					Range:          spanToSymbolRange(start, end),
					SelectionRange: posToLSPRange(ident.NamePos),
				})
			}
		case *syntax.LoadStmt:
			start, end := st.Span()
			symbols = append(symbols, protocol.DocumentSymbol{
				Name:           st.ModuleName(),
				Kind:           protocol.SymbolKindModule,
				Range:          spanToSymbolRange(start, end),
				SelectionRange: spanToSymbolRange(start, end),
			})
		}
	}

	return symbols, nil
}

func (s *Server) Symbols(
	_ context.Context,
	params *protocol.WorkspaceSymbolParams,
) ([]protocol.SymbolInformation, error) {
	if s.rootDir == "" {
		return nil, nil
	}

	query := strings.ToLower(params.Query)
	var symbols []protocol.SymbolInformation

	_ = filepath.WalkDir(s.rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".scampi" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		f, _ := Parse(path, data)
		if f == nil {
			return nil
		}

		for _, stmt := range f.Stmts {
			var name string
			var kind protocol.SymbolKind
			var pos syntax.Position

			switch st := stmt.(type) {
			case *syntax.DefStmt:
				name = st.Name.Name
				kind = protocol.SymbolKindFunction
				pos = st.Name.NamePos
			case *syntax.AssignStmt:
				if ident, ok := st.LHS.(*syntax.Ident); ok {
					name = ident.Name
					kind = protocol.SymbolKindVariable
					pos = ident.NamePos
				}
			}

			if name == "" {
				continue
			}
			if query != "" && !strings.Contains(strings.ToLower(name), query) {
				continue
			}

			symbols = append(symbols, protocol.SymbolInformation{
				Name: name,
				Kind: kind,
				Location: protocol.Location{
					URI:   uri.File(path),
					Range: posToLSPRange(pos),
				},
			})
		}
		return nil
	})

	return symbols, nil
}

func spanToSymbolRange(start, end syntax.Position) protocol.Range {
	startLine := uint32(0)
	if start.Line > 0 {
		startLine = uint32(start.Line - 1)
	}
	startCol := uint32(0)
	if start.Col > 0 {
		startCol = uint32(start.Col - 1)
	}
	endLine := startLine
	if end.Line > 0 {
		endLine = uint32(end.Line - 1)
	}
	endCol := startCol
	if end.Col > 0 {
		endCol = uint32(end.Col - 1)
	}
	return protocol.Range{
		Start: protocol.Position{Line: startLine, Character: startCol},
		End:   protocol.Position{Line: endLine, Character: endCol},
	}
}
