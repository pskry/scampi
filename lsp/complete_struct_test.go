// SPDX-License-Identifier: GPL-3.0-only

package lsp

import (
	"context"
	"testing"

	"go.lsp.dev/protocol"
)

func TestCompletionStructFieldForLoopVar(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	text := `module main

type Item {
  id:   int
  name: string
}

let items = [
  Item { id = 1, name = "a" },
]

func use(items: list[Item]) string {
  for item in items {
    let x = item.
  }
  return ""
}
`
	s.docs.Open(docURI, text, 1)

	result, err := s.Completion(context.Background(), &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
			Position:     protocol.Position{Line: 13, Character: 17},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result == nil || len(result.Items) == 0 {
		t.Fatal("expected struct field completions for 'item.' in for loop")
	}

	labels := make(map[string]bool)
	for _, item := range result.Items {
		labels[item.Label] = true
	}
	if !labels["id"] {
		t.Error("expected 'id' in completions")
	}
	if !labels["name"] {
		t.Error("expected 'name' in completions")
	}
}

func TestCompletionStructFieldKind(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	text := `module main

type Box {
  label: string
}

let b = Box { label = "x" }
let l = b.
`
	s.docs.Open(docURI, text, 1)

	result, err := s.Completion(context.Background(), &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
			Position:     protocol.Position{Line: 7, Character: 10},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result == nil || len(result.Items) == 0 {
		t.Fatal("expected completions")
	}

	for _, item := range result.Items {
		if item.Label == "label" {
			if item.Kind != protocol.CompletionItemKindField {
				t.Errorf("struct field completion should have Field kind, got %v", item.Kind)
			}
			return
		}
	}
	t.Error("'label' not found in completions")
}
