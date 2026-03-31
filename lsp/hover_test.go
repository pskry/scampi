// SPDX-License-Identifier: GPL-3.0-only

package lsp

import (
	"context"
	"strings"
	"testing"

	"go.lsp.dev/protocol"
)

func TestHoverOnFuncName(t *testing.T) {
	s := testServer()
	uri := protocol.DocumentURI("file:///test.scampi")
	s.docs.Open(uri, "copy", 1)

	result, err := s.Hover(context.Background(), &protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: uri},
			Position:     protocol.Position{Line: 0, Character: 2},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected hover result")
	}
	if !strings.Contains(result.Contents.Value, "copy") {
		t.Errorf("hover should mention 'copy', got %q", result.Contents.Value)
	}
	if !strings.Contains(result.Contents.Value, "`src`") {
		t.Error("hover should include parameter docs")
	}
}

func TestHoverOnKwarg(t *testing.T) {
	s := testServer()
	uri := protocol.DocumentURI("file:///test.scampi")
	text := `copy(dest="/etc/foo")`
	s.docs.Open(uri, text, 1)

	// Hover on "dest" (chars 5-8)
	result, err := s.Hover(context.Background(), &protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: uri},
			Position:     protocol.Position{Line: 0, Character: 6},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected hover result for kwarg")
	}
	if !strings.Contains(result.Contents.Value, "dest") {
		t.Errorf("hover should mention 'dest', got %q", result.Contents.Value)
	}
}

func TestHoverOnUnknown(t *testing.T) {
	s := testServer()
	uri := protocol.DocumentURI("file:///test.scampi")
	s.docs.Open(uri, "foobar = 42", 1)

	result, err := s.Hover(context.Background(), &protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: uri},
			Position:     protocol.Position{Line: 0, Character: 3},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result != nil {
		t.Error("expected nil hover for unknown word")
	}
}

func TestHoverOnDottedFunc(t *testing.T) {
	s := testServer()
	uri := protocol.DocumentURI("file:///test.scampi")
	s.docs.Open(uri, "target.ssh", 1)

	result, err := s.Hover(context.Background(), &protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: uri},
			Position:     protocol.Position{Line: 0, Character: 8},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected hover for target.ssh")
	}
	if !strings.Contains(result.Contents.Value, "target.ssh") {
		t.Errorf("hover should mention 'target.ssh', got %q", result.Contents.Value)
	}
}
