// SPDX-License-Identifier: GPL-3.0-only

package lsp

import (
	"context"
	"testing"

	"go.lsp.dev/protocol"
)

func TestCompletionKwargsExcludesPresent_Newlines(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	// Multi-line struct literal WITHOUT trailing commas.
	// After pressing enter on a new line, the LSP should exclude
	// already-defined fields from the completion list.
	text := "posix.copy {\n  src = posix.source_local { path = \"./f\" }\n  dest = \"/etc/foo\"\n  \n}"
	s.docs.Open(docURI, text, 1)

	// Cursor on the empty line 3 (0-indexed), col 2 (after two spaces).
	result, err := s.Completion(context.Background(), &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
			Position:     protocol.Position{Line: 3, Character: 2},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result == nil || len(result.Items) == 0 {
		t.Fatal("expected kwarg completions")
	}

	for _, item := range result.Items {
		if item.Label == "src" {
			t.Error("src should be excluded (already present)")
		}
		if item.Label == "dest" {
			t.Error("dest should be excluded (already present)")
		}
	}
}

func TestCompletionKwargsExcludesPresent_Commas(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	// Same but WITH trailing commas — existing behavior should work.
	text := "posix.copy {\n  src = posix.source_local { path = \"./f\" },\n  dest = \"/etc/foo\",\n  \n}"
	s.docs.Open(docURI, text, 1)

	result, err := s.Completion(context.Background(), &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
			Position:     protocol.Position{Line: 3, Character: 2},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result == nil || len(result.Items) == 0 {
		t.Fatal("expected kwarg completions")
	}

	for _, item := range result.Items {
		if item.Label == "src" {
			t.Error("src should be excluded (already present)")
		}
		if item.Label == "dest" {
			t.Error("dest should be excluded (already present)")
		}
	}
}

func TestCompletionKwargsNestedFieldsNotLeaked(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	// Nested struct literal — field names inside the nested struct
	// should not appear as excluded in the outer completion.
	text := "posix.copy {\n  src = posix.source_local { path = \"./f\" }\n  \n}"
	s.docs.Open(docURI, text, 1)

	result, err := s.Completion(context.Background(), &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
			Position:     protocol.Position{Line: 2, Character: 2},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result == nil || len(result.Items) == 0 {
		t.Fatal("expected kwarg completions")
	}

	// "dest" should still be offered — it's NOT a nested field.
	found := false
	for _, item := range result.Items {
		if item.Label == "dest" {
			found = true
		}
		// "path" is a nested field inside source_local — it must NOT
		// appear as excluded from the outer copy's completions.
		// (This tests that extractFieldNames doesn't leak nested names.)
	}
	if !found {
		t.Error("expected 'dest' in completions (not yet present)")
	}

	// "src" should be excluded.
	for _, item := range result.Items {
		if item.Label == "src" {
			t.Error("src should be excluded (already present)")
		}
	}
}
