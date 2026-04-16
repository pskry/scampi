// SPDX-License-Identifier: GPL-3.0-only

package lsp

import (
	"context"
	"strings"
	"testing"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"

	"scampi.dev/scampi/lang/check"
)

func TestCodeAction_AddImportUnknownModule(t *testing.T) {
	s := testServer()
	src := "module main\nposix.pkg_system { name = \"vim\" }\n"
	docURI := protocol.DocumentURI(uri.File("/test/codeaction.scampi"))
	s.docs.Open(docURI, src, 1)

	diag := protocol.Diagnostic{
		Range: protocol.Range{
			Start: protocol.Position{Line: 1, Character: 0},
			End:   protocol.Position{Line: 1, Character: 5},
		},
		Message: "unknown module: std/posix",
		Code:    check.CodeUnknownModule,
	}

	actions, err := s.codeAction(context.Background(), &protocol.CodeActionParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
		Context:      protocol.CodeActionContext{Diagnostics: []protocol.Diagnostic{diag}},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	if !strings.Contains(actions[0].Title, `import "std/posix"`) {
		t.Errorf("expected add-import action, got %q", actions[0].Title)
	}
	if actions[0].Kind != protocol.QuickFix {
		t.Errorf("expected QuickFix kind, got %v", actions[0].Kind)
	}
}

func TestCodeAction_AddImportUndefinedName(t *testing.T) {
	s := testServer()
	src := "module main\nlocal.target { name = \"h\" }\n"
	docURI := protocol.DocumentURI(uri.File("/test/codeaction_undef.scampi"))
	s.docs.Open(docURI, src, 1)

	diag := protocol.Diagnostic{
		Range: protocol.Range{
			Start: protocol.Position{Line: 1, Character: 0},
			End:   protocol.Position{Line: 1, Character: 5},
		},
		Message: "undefined: local",
		Code:    check.CodeUndefined,
	}

	actions, err := s.codeAction(context.Background(), &protocol.CodeActionParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
		Context:      protocol.CodeActionContext{Diagnostics: []protocol.Diagnostic{diag}},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	if !strings.Contains(actions[0].Title, `import "std/local"`) {
		t.Errorf("expected add-import for std/local, got %q", actions[0].Title)
	}
}

func TestCodeAction_NoActionForAlreadyImported(t *testing.T) {
	s := testServer()
	src := "module main\nimport \"std/local\"\nlocal.target { name = \"h\" }\n"
	docURI := protocol.DocumentURI(uri.File("/test/codeaction_dup.scampi"))
	s.docs.Open(docURI, src, 1)

	diag := protocol.Diagnostic{
		Range: protocol.Range{
			Start: protocol.Position{Line: 2, Character: 0},
			End:   protocol.Position{Line: 2, Character: 5},
		},
		Message: "undefined: local",
		Code:    check.CodeUndefined,
	}

	actions, err := s.codeAction(context.Background(), &protocol.CodeActionParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
		Context:      protocol.CodeActionContext{Diagnostics: []protocol.Diagnostic{diag}},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(actions) != 0 {
		t.Errorf("expected 0 actions when already imported, got %d", len(actions))
	}
}

func TestCodeAction_RemoveDuplicateImport(t *testing.T) {
	s := testServer()
	src := "module main\nimport \"std\"\nimport \"std\"\n"
	docURI := protocol.DocumentURI(uri.File("/test/codeaction_dupimport.scampi"))
	s.docs.Open(docURI, src, 1)

	diag := protocol.Diagnostic{
		Range: protocol.Range{
			Start: protocol.Position{Line: 2, Character: 0},
			End:   protocol.Position{Line: 2, Character: 12},
		},
		Message: "duplicate import",
		Code:    check.CodeDuplicateImport,
	}

	actions, err := s.codeAction(context.Background(), &protocol.CodeActionParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
		Context:      protocol.CodeActionContext{Diagnostics: []protocol.Diagnostic{diag}},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	if actions[0].Title != "Remove duplicate import" {
		t.Errorf("expected 'Remove duplicate import', got %q", actions[0].Title)
	}
	// Verify the edit deletes line 2
	edits := actions[0].Edit.Changes[docURI]
	if len(edits) != 1 {
		t.Fatalf("expected 1 edit, got %d", len(edits))
	}
	if edits[0].NewText != "" {
		t.Errorf("expected empty new text (deletion), got %q", edits[0].NewText)
	}
}

func TestCodeAction_RemoveDuplicateField(t *testing.T) {
	s := testServer()
	src := "module main\nimport \"std/local\"\nlet h = local.target { name = \"a\", name = \"b\" }\n"
	docURI := protocol.DocumentURI(uri.File("/test/codeaction_dupfield.scampi"))
	s.docs.Open(docURI, src, 1)

	diag := protocol.Diagnostic{
		Range: protocol.Range{
			Start: protocol.Position{Line: 2, Character: 36},
			End:   protocol.Position{Line: 2, Character: 40},
		},
		Message: "duplicate field",
		Code:    check.CodeDuplicateField,
	}

	actions, err := s.codeAction(context.Background(), &protocol.CodeActionParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
		Context:      protocol.CodeActionContext{Diagnostics: []protocol.Diagnostic{diag}},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	if actions[0].Title != "Remove duplicate field" {
		t.Errorf("expected 'Remove duplicate field', got %q", actions[0].Title)
	}
}

func TestCodeAction_UnknownCode(t *testing.T) {
	s := testServer()
	src := "module main\n"
	docURI := protocol.DocumentURI(uri.File("/test/codeaction_unknown.scampi"))
	s.docs.Open(docURI, src, 1)

	diag := protocol.Diagnostic{
		Range:   protocol.Range{},
		Message: "some random error",
		Code:    "totally.unknown.code",
	}

	actions, err := s.codeAction(context.Background(), &protocol.CodeActionParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
		Context:      protocol.CodeActionContext{Diagnostics: []protocol.Diagnostic{diag}},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(actions) != 0 {
		t.Errorf("expected 0 actions for unknown code, got %d", len(actions))
	}
}

func TestCodeAction_UnknownModuleNotStd(t *testing.T) {
	s := testServer()
	src := "module main\n"
	docURI := protocol.DocumentURI(uri.File("/test/codeaction_custom.scampi"))
	s.docs.Open(docURI, src, 1)

	diag := protocol.Diagnostic{
		Range:   protocol.Range{},
		Message: "unknown module: custom/thing",
		Code:    check.CodeUnknownModule,
	}

	actions, err := s.codeAction(context.Background(), &protocol.CodeActionParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
		Context:      protocol.CodeActionContext{Diagnostics: []protocol.Diagnostic{diag}},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(actions) != 0 {
		t.Errorf("expected 0 actions for non-std module, got %d", len(actions))
	}
}

func TestCodeAction_NoDocument(t *testing.T) {
	s := testServer()
	docURI := protocol.DocumentURI(uri.File("/test/missing.scampi"))

	actions, err := s.codeAction(context.Background(), &protocol.CodeActionParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
		Context:      protocol.CodeActionContext{Diagnostics: []protocol.Diagnostic{{}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if actions != nil {
		t.Errorf("expected nil actions for missing doc, got %v", actions)
	}
}

func TestFindImportInsertLine(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    int
	}{
		{
			name:    "after last import",
			content: "module main\nimport \"std\"\nimport \"std/local\"\nlet x = 1\n",
			want:    3,
		},
		{
			name:    "after module decl when no imports",
			content: "module main\nlet x = 1\n",
			want:    2, // moduleLine(0) + 2
		},
		{
			name:    "empty content",
			content: "",
			want:    0,
		},
		{
			name:    "single import",
			content: "module main\nimport \"std\"\n",
			want:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findImportInsertLine(tt.content)
			if got != tt.want {
				t.Errorf("findImportInsertLine() = %d, want %d", got, tt.want)
			}
		})
	}
}
