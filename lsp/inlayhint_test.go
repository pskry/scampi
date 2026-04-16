// SPDX-License-Identifier: GPL-3.0-only

package lsp

import (
	"testing"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestInlayHints_LetTypeHint(t *testing.T) {
	s := testServer()
	src := `module main
import "std"
import "std/local"
let host = local.target { name = "h" }
`
	docURI := protocol.DocumentURI(uri.File("/test/hints.scampi"))
	s.docs.Open(docURI, src, 1)

	hints := s.computeInlayHints(InlayHintParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
	})

	// Should have at least one type hint on "host"
	found := false
	for _, h := range hints {
		if h.Kind == InlayHintKindType && h.Label == ": Target" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected type hint ': Target' on let host, got %v", hints)
	}
}

func TestInlayHints_ParamNameHint(t *testing.T) {
	s := testServer()
	src := `module main
import "std"
import "std/local"
let host = local.target { name = "h" }
std.deploy("test", [host]) {
}
`
	docURI := protocol.DocumentURI(uri.File("/test/param_hints.scampi"))
	s.docs.Open(docURI, src, 1)

	hints := s.computeInlayHints(InlayHintParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
	})

	// Should have param name hints on positional args "test" and [host]
	nameHints := 0
	for _, h := range hints {
		if h.Kind == InlayHintKindParameter {
			nameHints++
		}
	}
	if nameHints < 2 {
		t.Errorf("expected at least 2 parameter hints, got %d: %v", nameHints, hints)
	}
}

func TestInlayHints_KwargTypeHint(t *testing.T) {
	s := testServer()
	src := `module main
import "std"
import "std/local"
let host = local.target { name = "h" }
`
	docURI := protocol.DocumentURI(uri.File("/test/kwarg_hints.scampi"))
	s.docs.Open(docURI, src, 1)

	hints := s.computeInlayHints(InlayHintParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
	})

	// Should have a type hint on the kwarg "name"
	found := false
	for _, h := range hints {
		if h.Kind == InlayHintKindType && h.Label == ": string" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected type hint ': string' on kwarg name, got %v", hints)
	}
}

func TestInlayHints_EmptyFile(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI(uri.File("/test/empty.scampi"))
	s.docs.Open(docURI, "module main\n", 1)

	hints := s.computeInlayHints(InlayHintParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
	})

	if len(hints) != 0 {
		t.Errorf("expected 0 hints for empty file, got %d", len(hints))
	}
}
