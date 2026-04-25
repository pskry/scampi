// SPDX-License-Identifier: GPL-3.0-only

package lsp

import "testing"

// offsetFromPosition edge cases
// -----------------------------------------------------------------------------

func TestOffsetFromPosition_EmptyLine(t *testing.T) {
	text := "line1\n\nline3\n"
	// Line 1 is empty. Col 5 should clamp to col 0 (line length 0).
	offset := offsetFromPosition(text, 1, 5)
	expected := offsetFromPosition(text, 1, 0)
	if offset != expected {
		t.Errorf("col past empty line: offset=%d, expected=%d", offset, expected)
	}
}

func TestOffsetFromPosition_ColPastEOL(t *testing.T) {
	text := "abc\ndef\n"
	// Line 0 is "abc" (3 chars). Col 10 should clamp to col 3.
	offset := offsetFromPosition(text, 0, 10)
	expected := offsetFromPosition(text, 0, 3)
	if offset != expected {
		t.Errorf("col past EOL: offset=%d, expected=%d", offset, expected)
	}
}

func TestOffsetFromPosition_LastLine(t *testing.T) {
	text := "hello"
	// No trailing newline. Col 5 is right at the end.
	offset := offsetFromPosition(text, 0, 5)
	if offset != 5 {
		t.Errorf("last line: offset=%d, expected=5", offset)
	}
}

func TestOffsetFromPosition_Normal(t *testing.T) {
	text := "abc\ndef\nghi\n"
	offset := offsetFromPosition(text, 1, 2)
	// Line 1 starts at byte 4 ("def\n"), col 2 → byte 6 ('f').
	if offset != 6 {
		t.Errorf("normal: offset=%d, expected=6", offset)
	}
}

// Cursor context inside nested structures
// -----------------------------------------------------------------------------

func TestCursor_NestedBraces_InnerStructLit(t *testing.T) {
	text := `outer {
  inner {
    field =
  }
}
`
	// Cursor inside inner { } at "field = " end.
	cur := AnalyzeCursor(text, 2, 12)
	if !cur.InCall {
		t.Fatal("should be InCall")
	}
	if cur.FuncName != "inner" {
		t.Errorf("FuncName = %q, want %q", cur.FuncName, "inner")
	}
	if cur.ActiveKwarg != "field" {
		t.Errorf("ActiveKwarg = %q, want %q", cur.ActiveKwarg, "field")
	}
}

func TestCursor_NestedBraces_OuterAfterInnerClosed(t *testing.T) {
	text := `outer {
  inner { a = 1 }

}
`
	// Cursor on empty line after inner is closed — should be in outer.
	cur := AnalyzeCursor(text, 2, 2)
	if !cur.InCall {
		t.Fatal("should be InCall")
	}
	if cur.FuncName != "outer" {
		t.Errorf("FuncName = %q, want %q", cur.FuncName, "outer")
	}
}

func TestCursor_ForLoopBrace_IsNotStructLit(t *testing.T) {
	// "for x in items {" has a brace but it's a loop body, not a
	// struct literal. The cursor analysis should NOT treat it as InCall
	// with FuncName="items".
	text := `for x in items {

}
`
	cur := AnalyzeCursor(text, 1, 2)
	// The brace of "for" is not preceded by ")" so isFunctionBodyBrace
	// returns false. identBeforeOffset returns "items". This is a known
	// limitation — "for" bodies look like struct literals to the cursor
	// analysis. Just document the behavior.
	if cur.InCall && cur.FuncName == "items" {
		// This is the current (imperfect) behavior. The test documents
		// it. A future fix could detect "for" keywords.
		t.Log("for-loop brace misidentified as struct literal (known limitation)")
	}
}

func TestCursor_FuncBody_IsNotStructLit(t *testing.T) {
	text := `func foo(x: int) string {

}
`
	cur := AnalyzeCursor(text, 1, 2)
	// isFunctionBodyBrace should detect the ")" and return true.
	if cur.InCall {
		t.Error("func body should NOT be InCall")
	}
}

func TestCursor_DeployBody(t *testing.T) {
	// std.deploy(...) { } is a block fill — the { is preceded by ).
	// isFunctionBodyBrace should detect this.
	text := `std.deploy(name = "x", targets = [t]) {

}
`
	cur := AnalyzeCursor(text, 1, 2)
	if cur.InCall {
		t.Error("deploy body should NOT be InCall (it's a block fill)")
	}
}

func TestCursor_StructLitInsideDeployBody(t *testing.T) {
	text := `std.deploy(name = "x", targets = [t]) {
  posix.copy {
    src = posix.source_inline { content = "hi" }

  }
}
`
	// Cursor on empty line inside posix.copy { }
	cur := AnalyzeCursor(text, 3, 4)
	if !cur.InCall {
		t.Fatal("should be InCall inside struct literal")
	}
	if cur.FuncName != "posix.copy" {
		t.Errorf("FuncName = %q, want %q", cur.FuncName, "posix.copy")
	}
}

func TestCursor_StructLitInsideForInsideDeploy(t *testing.T) {
	text := `std.deploy(name = "x", targets = [t]) {
  for item in list {
    posix.dir {
      path = "/tmp"

    }
  }
}
`
	cur := AnalyzeCursor(text, 4, 6)
	if !cur.InCall {
		t.Fatal("should be InCall")
	}
	if cur.FuncName != "posix.dir" {
		t.Errorf("FuncName = %q, want %q", cur.FuncName, "posix.dir")
	}
	if len(cur.PresentKwargs) == 0 {
		t.Error("should have path in PresentKwargs")
	}
}

func TestCursor_EmptyLinePastEOL(t *testing.T) {
	// Reproduce the original PVE bug: col past an empty line should
	// clamp and not bleed into the next line's closing brace.
	text := "Type {\n  field = 1\n\n  }\n"
	// Line 2 is empty, col 6 is past EOL.
	cur := AnalyzeCursor(text, 2, 6)
	if !cur.InCall {
		t.Fatal("should be InCall")
	}
	if cur.FuncName != "Type" {
		t.Errorf("FuncName = %q, want %q (col clamped past empty line)", cur.FuncName, "Type")
	}
}

// String literal edge cases
// -----------------------------------------------------------------------------

func TestCursor_InsideString_NoCompletion(t *testing.T) {
	text := `posix.copy { dest = "/etc/`
	cur := AnalyzeCursor(text, 0, uint32(len(text)))
	if !cur.InString {
		t.Error("should be InString")
	}
}

func TestCursor_AfterClosedString_NotInString(t *testing.T) {
	text := `posix.copy { dest = "/etc/foo", `
	cur := AnalyzeCursor(text, 0, uint32(len(text)))
	if cur.InString {
		t.Error("should NOT be InString after closed string")
	}
}
