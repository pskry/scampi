// SPDX-License-Identifier: GPL-3.0-only

package lsp

import (
	"context"
	"testing"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestDocumentHighlight_DefinitionAndUses(t *testing.T) {
	s := testServer()
	src := `module main
import "std"
import "std/local"
let host = local.target { name = "h" }
std.deploy("test", [host]) {
}
`
	docURI := protocol.DocumentURI(uri.File("/test/highlight.scampi"))
	s.docs.Open(docURI, src, 1)

	// Cursor on "host" in the let binding (line 3, col 4)
	highlights, err := s.DocumentHighlight(context.Background(), &protocol.DocumentHighlightParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
			Position:     protocol.Position{Line: 3, Character: 4},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(highlights) < 2 {
		t.Fatalf("expected at least 2 highlights (def + use), got %d", len(highlights))
	}

	var writes, reads int
	for _, h := range highlights {
		switch h.Kind {
		case protocol.DocumentHighlightKindWrite:
			writes++
		case protocol.DocumentHighlightKindRead:
			reads++
		}
	}
	if writes != 1 {
		t.Errorf("expected 1 write highlight (definition), got %d", writes)
	}
	if reads < 1 {
		t.Errorf("expected at least 1 read highlight (usage), got %d", reads)
	}
}

func TestDocumentHighlight_EmptyWord(t *testing.T) {
	s := testServer()
	src := "module main\n\n"
	docURI := protocol.DocumentURI(uri.File("/test/highlight_empty.scampi"))
	s.docs.Open(docURI, src, 1)

	// Cursor on the blank line — no word
	highlights, err := s.DocumentHighlight(context.Background(), &protocol.DocumentHighlightParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
			Position:     protocol.Position{Line: 1, Character: 0},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(highlights) != 0 {
		t.Errorf("expected 0 highlights on blank line, got %d", len(highlights))
	}
}

func TestDocumentHighlight_NoDocument(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI(uri.File("/test/nonexistent.scampi"))

	highlights, err := s.DocumentHighlight(context.Background(), &protocol.DocumentHighlightParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
			Position:     protocol.Position{Line: 0, Character: 0},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if highlights != nil {
		t.Errorf("expected nil highlights for missing doc, got %v", highlights)
	}
}

func TestDocumentHighlight_SingleOccurrence(t *testing.T) {
	s := testServer()
	src := `module main
import "std"
import "std/local"
let host = local.target { name = "h" }
`
	docURI := protocol.DocumentURI(uri.File("/test/highlight_single.scampi"))
	s.docs.Open(docURI, src, 1)

	// "host" only appears once (the let binding), no usages
	highlights, err := s.DocumentHighlight(context.Background(), &protocol.DocumentHighlightParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
			Position:     protocol.Position{Line: 3, Character: 4},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(highlights) != 1 {
		t.Fatalf("expected 1 highlight for single occurrence, got %d", len(highlights))
	}
	if highlights[0].Kind != protocol.DocumentHighlightKindWrite {
		t.Errorf("expected Write kind for definition, got %v", highlights[0].Kind)
	}
}
