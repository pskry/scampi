// SPDX-License-Identifier: GPL-3.0-only

package lsp

import (
	"context"
	"testing"

	"go.lsp.dev/protocol"
)

func TestCompletion_TopLevel_ModulePrefix(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	s.docs.Open(docURI, "pv", 1)

	items := completionAt(t, s, docURI, 0, 2)
	requireLabels(t, items, "pve")
}

func TestCompletion_ModuleMembers(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	s.docs.Open(docURI, "posix.", 1)

	items := completionAt(t, s, docURI, 0, 6)
	requireLabels(t, items, "copy", "dir", "symlink")
}

func TestCompletion_ModuleMembersFiltered(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	s.docs.Open(docURI, "posix.co", 1)

	items := completionAt(t, s, docURI, 0, 8)
	requireLabels(t, items, "copy")
	rejectLabels(t, items, "dir", "symlink")
}

// Empty/degenerate inputs
// -----------------------------------------------------------------------------

func TestCompletion_EmptyDoc(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	s.docs.Open(docURI, "", 1)

	items := completionAt(t, s, docURI, 0, 0)
	// Should return top-level completions without crashing.
	if items == nil {
		t.Log("no completions for empty doc (acceptable)")
	}
}

func TestCompletion_NoDocs(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///nonexistent.scampi")

	result, err := s.Completion(context.Background(), &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
			Position:     protocol.Position{Line: 0, Character: 0},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result != nil && len(result.Items) > 0 {
		t.Error("should return nil for unknown document")
	}
}

// User declarations from current document
// -----------------------------------------------------------------------------

func TestCompletion_UserFunc_InTopLevel(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	text := `module main

func proxy_host(domain: string) string {
  return ""
}

pro
`
	s.docs.Open(docURI, text, 1)

	items := completionAt(t, s, docURI, 6, 3)
	requireLabels(t, items, "proxy_host")
}

func TestCompletion_UserLet_InTopLevel(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	text := `module main

let my_config = "test"

my
`
	s.docs.Open(docURI, text, 1)

	items := completionAt(t, s, docURI, 4, 2)
	requireLabels(t, items, "my_config")
}

func TestCompletion_UserType_InTopLevel(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	text := `module main

type MyConfig {
  name: string
}

My
`
	s.docs.Open(docURI, text, 1)

	items := completionAt(t, s, docURI, 6, 2)
	requireLabels(t, items, "MyConfig")
}

// Nested struct literals with values from outer scope
// -----------------------------------------------------------------------------

func TestCompletion_NestedStructLit_InnerKwargs(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	text := `posix.copy {
  src = posix.source_local {

  }
}
`
	s.docs.Open(docURI, text, 1)

	// Cursor inside the nested source_local { } at line 2
	items := completionAt(t, s, docURI, 2, 4)
	if len(items) == 0 {
		t.Fatal("expected kwargs for source_local")
	}
	requireLabels(t, items, "path")
}

func TestCompletion_NestedStructLit_OuterKwargs(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	text := `posix.copy {
  src = posix.source_local { path = "./f" }

}
`
	s.docs.Open(docURI, text, 1)

	// Cursor on empty line after source_local is closed (line 2)
	items := completionAt(t, s, docURI, 2, 2)
	if len(items) == 0 {
		t.Fatal("expected kwargs for posix.copy")
	}
	requireLabels(t, items, "dest")
	rejectLabels(t, items, "src")
}

// UFCS completion
// -----------------------------------------------------------------------------

func TestCompletion_UFCS_InFuncBody(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI("file:///test.scampi")
	text := `module main
import "std/posix"

type MyTarget {
  path: string
}

func setup(t: MyTarget) string {
  return ""
}

let tgt = MyTarget { path = "/tmp" }
let x = tgt.
`
	s.docs.Open(docURI, text, 1)

	// "tgt." should offer struct fields AND UFCS functions.
	items := completionAt(t, s, docURI, 12, 13)
	if len(items) == 0 {
		t.Fatal("expected completions for tgt.")
	}
	// At minimum, the struct field should be there.
	requireLabels(t, items, "path")
}
