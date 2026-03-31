// SPDX-License-Identifier: GPL-3.0-only

package lsp

import (
	"context"
	"io"
	"log"
	"testing"

	"go.lsp.dev/protocol"
)

func testServer() *Server {
	return &Server{
		catalog: NewCatalog(),
		docs:    NewDocuments(),
		log:     log.New(io.Discard, "", 0),
	}
}

func TestCompletionTopLevel(t *testing.T) {
	s := testServer()
	uri := protocol.DocumentURI("file:///test.scampi")
	s.docs.Open(uri, "cop", 1)

	result, err := s.Completion(context.Background(), &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: uri},
			Position:     protocol.Position{Line: 0, Character: 3},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result == nil || len(result.Items) == 0 {
		t.Fatal("expected completion items for 'cop'")
	}

	found := false
	for _, item := range result.Items {
		if item.Label == "copy" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'copy' in completion items")
	}
}

func TestCompletionKwargs(t *testing.T) {
	s := testServer()
	uri := protocol.DocumentURI("file:///test.scampi")
	text := `copy(src=local("./f"), `
	s.docs.Open(uri, text, 1)

	result, err := s.Completion(context.Background(), &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: uri},
			Position:     protocol.Position{Line: 0, Character: uint32(len(text))},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result == nil || len(result.Items) == 0 {
		t.Fatal("expected kwarg completions")
	}

	// "src" should be excluded since it's already present.
	for _, item := range result.Items {
		if item.Label == "src" {
			t.Error("src should be excluded from completions (already present)")
		}
	}

	// "dest" should be offered.
	found := false
	for _, item := range result.Items {
		if item.Label == "dest" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'dest' in kwarg completions")
	}
}

func TestCompletionModule(t *testing.T) {
	s := testServer()
	uri := protocol.DocumentURI("file:///test.scampi")
	s.docs.Open(uri, "target.", 1)

	result, err := s.Completion(context.Background(), &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: uri},
			Position:     protocol.Position{Line: 0, Character: 7},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result == nil || len(result.Items) == 0 {
		t.Fatal("expected module member completions")
	}

	labels := make(map[string]bool)
	for _, item := range result.Items {
		labels[item.Label] = true
	}
	for _, want := range []string{"ssh", "local", "rest"} {
		if !labels[want] {
			t.Errorf("missing target.%s in completions", want)
		}
	}
}
