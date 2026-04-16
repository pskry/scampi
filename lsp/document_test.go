// SPDX-License-Identifier: GPL-3.0-only

package lsp

import (
	"testing"

	"go.lsp.dev/protocol"
)

func TestApplyIncremental_UnknownURI(t *testing.T) {
	docs := NewDocuments()
	got := docs.ApplyIncremental("file:///unknown.scampi", nil, 2)
	if got != "" {
		t.Errorf("expected empty string for unknown URI, got %q", got)
	}
}

func TestApplyIncremental_FullReplacement(t *testing.T) {
	docs := NewDocuments()
	docURI := protocol.DocumentURI("file:///test.scampi")
	docs.Open(docURI, "old content", 1)

	got := docs.ApplyIncremental(docURI, []protocol.TextDocumentContentChangeEvent{
		{Text: "new content"},
	}, 2)
	if got != "new content" {
		t.Errorf("expected 'new content', got %q", got)
	}
}

func TestApplyIncremental_RangeSplice(t *testing.T) {
	docs := NewDocuments()
	docURI := protocol.DocumentURI("file:///test.scampi")
	docs.Open(docURI, "hello world", 1)

	// Replace "world" (offset 6..11) with "scampi"
	got := docs.ApplyIncremental(docURI, []protocol.TextDocumentContentChangeEvent{
		{
			Range: protocol.Range{
				Start: protocol.Position{Line: 0, Character: 6},
				End:   protocol.Position{Line: 0, Character: 11},
			},
			Text: "scampi",
		},
	}, 2)
	if got != "hello scampi" {
		t.Errorf("expected 'hello scampi', got %q", got)
	}
}

func TestApplyIncremental_MultipleChanges(t *testing.T) {
	docs := NewDocuments()
	docURI := protocol.DocumentURI("file:///test.scampi")
	docs.Open(docURI, "aaa bbb ccc", 1)

	// First change: replace "bbb" with "xxx"
	// Second change: full replacement
	got := docs.ApplyIncremental(docURI, []protocol.TextDocumentContentChangeEvent{
		{
			Range: protocol.Range{
				Start: protocol.Position{Line: 0, Character: 4},
				End:   protocol.Position{Line: 0, Character: 7},
			},
			Text: "xxx",
		},
		{Text: "replaced"},
	}, 2)
	if got != "replaced" {
		t.Errorf("expected 'replaced', got %q", got)
	}
}

func TestApplyIncremental_MultiLine(t *testing.T) {
	docs := NewDocuments()
	docURI := protocol.DocumentURI("file:///test.scampi")
	docs.Open(docURI, "line0\nline1\nline2\n", 1)

	// Replace "line1" on line 1 with "CHANGED"
	got := docs.ApplyIncremental(docURI, []protocol.TextDocumentContentChangeEvent{
		{
			Range: protocol.Range{
				Start: protocol.Position{Line: 1, Character: 0},
				End:   protocol.Position{Line: 1, Character: 5},
			},
			Text: "CHANGED",
		},
	}, 2)
	if got != "line0\nCHANGED\nline2\n" {
		t.Errorf("expected 'line0\\nCHANGED\\nline2\\n', got %q", got)
	}
}

func TestApplyIncremental_InsertAtEnd(t *testing.T) {
	docs := NewDocuments()
	docURI := protocol.DocumentURI("file:///test.scampi")
	docs.Open(docURI, "abc", 1)

	// Insert at end of line
	got := docs.ApplyIncremental(docURI, []protocol.TextDocumentContentChangeEvent{
		{
			Range: protocol.Range{
				Start: protocol.Position{Line: 0, Character: 3},
				End:   protocol.Position{Line: 0, Character: 3},
			},
			Text: "def",
		},
	}, 2)
	if got != "abcdef" {
		t.Errorf("expected 'abcdef', got %q", got)
	}
}

func TestApplyIncremental_DeleteRange(t *testing.T) {
	docs := NewDocuments()
	docURI := protocol.DocumentURI("file:///test.scampi")
	docs.Open(docURI, "abcdef", 1)

	// Delete "cd" (chars 2..4)
	got := docs.ApplyIncremental(docURI, []protocol.TextDocumentContentChangeEvent{
		{
			Range: protocol.Range{
				Start: protocol.Position{Line: 0, Character: 2},
				End:   protocol.Position{Line: 0, Character: 4},
			},
			Text: "",
		},
	}, 2)
	if got != "abef" {
		t.Errorf("expected 'abef', got %q", got)
	}
}

func TestPosToOffset(t *testing.T) {
	content := "line0\nline1\nline2"

	tests := []struct {
		name     string
		line     uint32
		char     uint32
		expected int
	}{
		{"start of file", 0, 0, 0},
		{"mid first line", 0, 3, 3},
		{"start of line 1", 1, 0, 6},
		{"mid line 1", 1, 3, 9},
		{"start of line 2", 2, 0, 12},
		{"end of line 2", 2, 5, 17},
		{"past end of content", 5, 0, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := posToOffset(content, protocol.Position{
				Line:      tt.line,
				Character: tt.char,
			})
			if got != tt.expected {
				t.Errorf("posToOffset(line=%d, char=%d) = %d, want %d",
					tt.line, tt.char, got, tt.expected)
			}
		})
	}
}

func TestPosToOffset_EmptyContent(t *testing.T) {
	got := posToOffset("", protocol.Position{Line: 0, Character: 0})
	if got != 0 {
		t.Errorf("posToOffset on empty content = %d, want 0", got)
	}
}
