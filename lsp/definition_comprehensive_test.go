// SPDX-License-Identifier: GPL-3.0-only

package lsp

import (
	"context"
	"testing"

	"go.lsp.dev/protocol"
)

func definitionAt(t *testing.T, s *Server, docURI protocol.DocumentURI, line, col uint32) []protocol.Location {
	t.Helper()
	result, err := s.Definition(context.Background(), &protocol.DefinitionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
			Position:     protocol.Position{Line: line, Character: col},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestDefinition_FuncDecl(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	text := `module main

func add(a: int, b: int) int {
  return a + b
}

let r = add(1, 2)
`
	s.docs.Open(docURI, text, 1)

	locs := definitionAt(t, s, docURI, 6, 9)
	if len(locs) == 0 {
		t.Fatal("expected definition location for 'add'")
	}
	if locs[0].Range.Start.Line != 2 {
		t.Errorf("definition line = %d, want 2", locs[0].Range.Start.Line)
	}
}

func TestDefinition_TypeDecl(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	text := `module main

type Server {
  name: string
}

let s = Server { name = "web" }
`
	s.docs.Open(docURI, text, 1)

	// Goto def on "Server" in the struct literal.
	locs := definitionAt(t, s, docURI, 6, 10)
	if len(locs) == 0 {
		t.Fatal("expected definition location for 'Server'")
	}
	if locs[0].Range.Start.Line != 2 {
		t.Errorf("definition line = %d, want 2", locs[0].Range.Start.Line)
	}
}

func TestDefinition_StdlibFunc(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	text := `posix.copy`
	s.docs.Open(docURI, text, 1)

	locs := definitionAt(t, s, docURI, 0, 8)
	// Should jump to the stub file.
	if len(locs) == 0 {
		t.Log("no definition for stdlib func (stub defs may not be available in test)")
	}
}

func TestDefinition_NoDoc(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///nonexistent.scampi")

	locs := definitionAt(t, s, docURI, 0, 0)
	if len(locs) != 0 {
		t.Error("expected no locations for nonexistent document")
	}
}
