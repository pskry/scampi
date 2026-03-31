// SPDX-License-Identifier: GPL-3.0-only

package lsp

import (
	"testing"
)

func TestAnalyzeCursorTopLevel(t *testing.T) {
	text := "cop"
	cur := AnalyzeCursor(text, 0, 3)
	if cur.InCall {
		t.Error("should not be in call")
	}
	if cur.WordUnderCursor != "cop" {
		t.Errorf("word = %q, want 'cop'", cur.WordUnderCursor)
	}
}

func TestAnalyzeCursorInsideCall(t *testing.T) {
	text := `copy(src=local("./f"), dest=`
	cur := AnalyzeCursor(text, 0, uint32(len(text)))
	if !cur.InCall {
		t.Fatal("should be in call")
	}
	if cur.FuncName != "copy" {
		t.Errorf("func = %q, want 'copy'", cur.FuncName)
	}
	if len(cur.PresentKwargs) != 2 {
		t.Errorf("present kwargs = %v, want [src dest]", cur.PresentKwargs)
	}
}

func TestAnalyzeCursorDottedFunc(t *testing.T) {
	text := `target.ssh(name="web", host=`
	cur := AnalyzeCursor(text, 0, uint32(len(text)))
	if !cur.InCall {
		t.Fatal("should be in call")
	}
	if cur.FuncName != "target.ssh" {
		t.Errorf("func = %q, want 'target.ssh'", cur.FuncName)
	}
}

func TestAnalyzeCursorCommaCount(t *testing.T) {
	text := `pkg(name="nginx", state="present", `
	cur := AnalyzeCursor(text, 0, uint32(len(text)))
	if cur.ActiveParam != 2 {
		t.Errorf("active param = %d, want 2", cur.ActiveParam)
	}
}

func TestAnalyzeCursorNestedParens(t *testing.T) {
	text := `copy(src=local("./f"), `
	cur := AnalyzeCursor(text, 0, uint32(len(text)))
	if !cur.InCall {
		t.Fatal("should be in call to copy, not local")
	}
	if cur.FuncName != "copy" {
		t.Errorf("func = %q, want 'copy'", cur.FuncName)
	}
}

func TestAnalyzeCursorMultiline(t *testing.T) {
	text := "copy(\n    src=local(\"./f\"),\n    "
	// cursor is on line 2, col 4
	cur := AnalyzeCursor(text, 2, 4)
	if !cur.InCall {
		t.Fatal("should be in call")
	}
	if cur.FuncName != "copy" {
		t.Errorf("func = %q, want 'copy'", cur.FuncName)
	}
}

func TestAnalyzeCursorModulePrefix(t *testing.T) {
	text := "target."
	cur := AnalyzeCursor(text, 0, 7)
	if cur.InCall {
		t.Error("should not be in call")
	}
	if cur.WordUnderCursor != "target." {
		t.Errorf("word = %q, want 'target.'", cur.WordUnderCursor)
	}
}

func TestAnalyzeCursorInsideList(t *testing.T) {
	text := `deploy(name="x", steps=[`
	cur := AnalyzeCursor(text, 0, uint32(len(text)))
	if !cur.InList {
		t.Fatal("should be InList")
	}
	if cur.InCall {
		t.Error("should not be InCall (innermost bracket is [)")
	}
	if cur.FuncName != "deploy" {
		t.Errorf("func = %q, want 'deploy'", cur.FuncName)
	}
}

func TestAnalyzeCursorInsideListMultiline(t *testing.T) {
	text := "deploy(\n    name=\"x\",\n    steps=[\n        "
	cur := AnalyzeCursor(text, 3, 8)
	if !cur.InList {
		t.Fatal("should be InList")
	}
	if cur.InCall {
		t.Error("should not be InCall")
	}
}

func TestAnalyzeCursorInsideListAfterComma(t *testing.T) {
	text := `deploy(steps=[dir(path="/tmp"), `
	cur := AnalyzeCursor(text, 0, uint32(len(text)))
	if !cur.InList {
		t.Fatal("should be InList")
	}
}

func TestExtractKwargNames(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{`name="web", host="1.2.3.4"`, []string{"name", "host"}},
		{`x == 1, y=2`, []string{"y"}},
		{``, nil},
		{`name=`, []string{"name"}},
	}
	for _, tt := range tests {
		got := extractKwargNames(tt.input)
		if len(got) != len(tt.want) {
			t.Errorf("extractKwargNames(%q) = %v, want %v", tt.input, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("extractKwargNames(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
			}
		}
	}
}

func TestCountTopLevelCommas(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{`a, b, c`, 2},
		{`a`, 0},
		{`f(a, b), c`, 1},
		{`[1, 2], "a,b"`, 1},
	}
	for _, tt := range tests {
		got := countTopLevelCommas(tt.input)
		if got != tt.want {
			t.Errorf("countTopLevelCommas(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}
