// SPDX-License-Identifier: GPL-3.0-only

package lsp

import (
	"context"
	"strings"
	"testing"

	"go.lsp.dev/protocol"
)

func hoverAt(t *testing.T, s *Server, docURI protocol.DocumentURI, line, col uint32) *protocol.Hover {
	t.Helper()
	result, err := s.Hover(context.Background(), &protocol.HoverParams{
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

func requireHoverContains(t *testing.T, result *protocol.Hover, fragments ...string) {
	t.Helper()
	if result == nil {
		t.Fatal("expected hover result, got nil")
	}
	for _, f := range fragments {
		if !strings.Contains(result.Contents.Value, f) {
			t.Errorf("hover should contain %q, got:\n%s", f, result.Contents.Value)
		}
	}
}

func TestHover_StdlibEnum(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	s.docs.Open(docURI, `module main
import "std/pve"

let x = pve.Console.xtermjs
`, 1)

	result := hoverAt(t, s, docURI, 3, 14)
	if result != nil {
		requireHoverContains(t, result, "Console")
	}
}

func TestHover_UserType_NoFieldsOpaque(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	text := `module main

type Opaque
`
	s.docs.Open(docURI, text, 1)

	result := hoverAt(t, s, docURI, 2, 7)
	if result == nil {
		t.Fatal("expected hover for opaque type")
	}
	requireHoverContains(t, result, "Opaque")
}

func TestHover_UserFunc(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	text := `module main

func greet(name: string, loud: bool = false) string {
  return "hi"
}

greet(name = "world")
`
	s.docs.Open(docURI, text, 1)

	result := hoverAt(t, s, docURI, 6, 3)
	requireHoverContains(t, result, "greet", "name", "loud")
}

func TestHover_LetBindingStructType(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	text := `module main

type Box {
  label: string
}

let b = Box { label = "x" }
`
	s.docs.Open(docURI, text, 1)

	result := hoverAt(t, s, docURI, 6, 5)
	requireHoverContains(t, result, "b", "Box")
}

func TestHover_ForLoopVar(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	text := `module main

type Item {
  id: int
}

let items = [Item { id = 1 }]

func use() string {
  for item in items {
    let x = item
  }
  return ""
}
`
	s.docs.Open(docURI, text, 1)

	result := hoverAt(t, s, docURI, 10, 14)
	if result != nil {
		requireHoverContains(t, result, "item")
	}
}

func TestHover_StdlibType(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	s.docs.Open(docURI, "ssh.target", 1)

	result := hoverAt(t, s, docURI, 0, 8)
	requireHoverContains(t, result, "ssh.target")
}

func TestHover_KwargInsideStructLitWithNewlines(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	text := `posix.copy {
  src = posix.source_local { path = "./f" }
  dest = "/etc/foo"
  owner = "root"
}
`
	s.docs.Open(docURI, text, 1)

	result := hoverAt(t, s, docURI, 3, 4)
	requireHoverContains(t, result, "owner")
}
