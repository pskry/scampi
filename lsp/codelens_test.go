// SPDX-License-Identifier: GPL-3.0-only

package lsp

import (
	"context"
	"strings"
	"testing"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestCodeLens_ReferenceCounts(t *testing.T) {
	s := testServer()
	src := `module main
import "std"
import "std/local"
type Server {
  name: string
}
let srv = Server { name = "web" }
let other = Server { name = "api" }
`
	docURI := protocol.DocumentURI(uri.File("/test/codelens.scampi"))
	s.docs.Open(docURI, src, 1)

	lenses, err := s.CodeLens(context.Background(), &protocol.CodeLensParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Should have lenses for: Server, srv, other
	if len(lenses) < 3 {
		t.Fatalf("expected at least 3 lenses, got %d", len(lenses))
	}

	// Find the Server type lens — used twice (in srv and other lets)
	var serverLens *protocol.CodeLens
	for i, l := range lenses {
		if l.Command != nil && strings.Contains(l.Command.Title, "reference") {
			// Check if this is on line 3 (the type Server line)
			if l.Range.Start.Line == 3 {
				serverLens = &lenses[i]
				break
			}
		}
	}
	if serverLens == nil {
		t.Fatal("expected a code lens for Server type")
	}
	if !strings.Contains(serverLens.Command.Title, "2 references") {
		t.Errorf("expected '2 references' for Server, got %q", serverLens.Command.Title)
	}
}

func TestCodeLens_SingularReference(t *testing.T) {
	s := testServer()
	src := `module main
import "std"
import "std/local"
type Box {
  label: string
}
let b = Box { label = "x" }
`
	docURI := protocol.DocumentURI(uri.File("/test/codelens_singular.scampi"))
	s.docs.Open(docURI, src, 1)

	lenses, err := s.CodeLens(context.Background(), &protocol.CodeLensParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Box is used once (in the let), so "1 reference"
	for _, l := range lenses {
		if l.Range.Start.Line == 3 && l.Command != nil {
			if l.Command.Title != "1 reference" {
				t.Errorf("expected '1 reference' for Box, got %q", l.Command.Title)
			}
			return
		}
	}
	t.Fatal("expected a code lens for Box type")
}

func TestCodeLens_ZeroReferences(t *testing.T) {
	s := testServer()
	src := `module main
type Unused {
  x: string
}
`
	docURI := protocol.DocumentURI(uri.File("/test/codelens_zero.scampi"))
	s.docs.Open(docURI, src, 1)

	lenses, err := s.CodeLens(context.Background(), &protocol.CodeLensParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(lenses) == 0 {
		t.Fatal("expected at least 1 lens for the Unused type")
	}
	if lenses[0].Command.Title != "0 references" {
		t.Errorf("expected '0 references', got %q", lenses[0].Command.Title)
	}
}

func TestCodeLens_EmptyFile(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI(uri.File("/test/codelens_empty.scampi"))
	s.docs.Open(docURI, "module main\n", 1)

	lenses, err := s.CodeLens(context.Background(), &protocol.CodeLensParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(lenses) != 0 {
		t.Errorf("expected 0 lenses for file with no decls, got %d", len(lenses))
	}
}

func TestCodeLens_NoDocument(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI(uri.File("/test/missing.scampi"))

	lenses, err := s.CodeLens(context.Background(), &protocol.CodeLensParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
	})
	if err != nil {
		t.Fatal(err)
	}
	if lenses != nil {
		t.Errorf("expected nil lenses for missing doc, got %v", lenses)
	}
}
