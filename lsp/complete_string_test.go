// SPDX-License-Identifier: GPL-3.0-only

package lsp

import (
	"context"
	"testing"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestCompletion_SuppressedInsideString(t *testing.T) {
	s := testServer()
	// Cursor inside the string literal "he|llo" — should get zero completions
	src := `module main
let x = "hello"
`
	docURI := protocol.DocumentURI(uri.File("/test/string_suppress.scampi"))
	s.docs.Open(docURI, src, 1)

	result, err := s.Completion(context.Background(), &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
			Position:     protocol.Position{Line: 1, Character: 11}, // inside "hello"
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result != nil && len(result.Items) > 0 {
		t.Errorf("expected 0 completions inside bare string, got %d: %v",
			len(result.Items), result.Items)
	}
}

func TestCompletion_NotSuppressedOutsideString(t *testing.T) {
	s := testServer()
	// Cursor at top level typing "pos" — should get completions
	src := "module main\npos"
	docURI := protocol.DocumentURI(uri.File("/test/string_nosuppress.scampi"))
	s.docs.Open(docURI, src, 1)

	result, err := s.Completion(context.Background(), &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
			Position:     protocol.Position{Line: 1, Character: 3},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result == nil || len(result.Items) == 0 {
		t.Fatal("expected completions outside string")
	}

	found := false
	for _, item := range result.Items {
		if item.Label == "posix" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'posix' in completions outside string")
	}
}

func TestCompletion_KeywordsOffered(t *testing.T) {
	s := testServer()
	src := "module main\nle"
	docURI := protocol.DocumentURI(uri.File("/test/keyword_complete.scampi"))
	s.docs.Open(docURI, src, 1)

	result, err := s.Completion(context.Background(), &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
			Position:     protocol.Position{Line: 1, Character: 2},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected completion result")
	}

	found := false
	for _, item := range result.Items {
		if item.Label == "let" {
			found = true
			if item.Kind != protocol.CompletionItemKindKeyword {
				t.Errorf("expected Keyword kind for 'let', got %v", item.Kind)
			}
			break
		}
	}
	if !found {
		t.Error("expected 'let' keyword in completions")
	}
}
