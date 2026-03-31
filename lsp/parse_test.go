// SPDX-License-Identifier: GPL-3.0-only

package lsp

import (
	"testing"

	"go.lsp.dev/protocol"
)

func TestParseValidSource(t *testing.T) {
	src := `
x = 1
y = "hello"
`
	f, diags := Parse("test.scampi", []byte(src))
	if len(diags) > 0 {
		t.Errorf("expected no diagnostics, got %d: %v", len(diags), diags)
	}
	if f == nil {
		t.Error("expected non-nil AST")
	}
}

func TestParseSyntaxError(t *testing.T) {
	src := `x = (`
	_, diags := Parse("test.scampi", []byte(src))
	if len(diags) == 0 {
		t.Fatal("expected diagnostics for syntax error")
	}
	for _, d := range diags {
		if d.Severity != protocol.DiagnosticSeverityError {
			t.Errorf("expected error severity, got %v", d.Severity)
		}
		if d.Source != "scampi" {
			t.Errorf("expected source 'scampi', got %q", d.Source)
		}
	}
}

func TestParseSetLiteral(t *testing.T) {
	src := `x = set([1, 2, 3])`
	f, diags := Parse("test.scampi", []byte(src))
	if len(diags) > 0 {
		t.Errorf("set literals should be allowed: %v", diags)
	}
	if f == nil {
		t.Error("expected non-nil AST")
	}
}

func TestParseWhileLoop(t *testing.T) {
	src := `
x = 0
while x < 10:
    x += 1
`
	f, diags := Parse("test.scampi", []byte(src))
	if len(diags) > 0 {
		t.Errorf("while loops should be allowed: %v", diags)
	}
	if f == nil {
		t.Error("expected non-nil AST")
	}
}

func TestParseDiagnosticPosition(t *testing.T) {
	// Error on line 2 (1-indexed), the parser should give us line 1 (0-indexed).
	src := "x = 1\ny = (\n"
	_, diags := Parse("test.scampi", []byte(src))
	if len(diags) == 0 {
		t.Fatal("expected diagnostics")
	}
	d := diags[0]
	if d.Range.Start.Line < 1 {
		t.Errorf("expected error on line >= 1 (0-indexed), got line %d", d.Range.Start.Line)
	}
}
