// SPDX-License-Identifier: GPL-3.0-only

package lsp

import "testing"

func TestActiveFieldNewlineSeparator(t *testing.T) {
	// In struct literals without commas, newlines separate fields.
	// activeField must return the CURRENT field, not the first one.
	inside := "\n  id = 42\n  hostname = "
	got := activeField(inside)
	if got != "hostname" {
		t.Errorf("activeField = %q, want %q", got, "hostname")
	}
}

func TestActiveFieldNewlineSeparatorThreeFields(t *testing.T) {
	inside := "\n  id = 1\n  name = \"web\"\n  port = "
	got := activeField(inside)
	if got != "port" {
		t.Errorf("activeField = %q, want %q", got, "port")
	}
}

func TestActiveFieldWithCommas(t *testing.T) {
	// With commas, it should still work correctly.
	inside := "\n  id = 1,\n  name = "
	got := activeField(inside)
	if got != "name" {
		t.Errorf("activeField = %q, want %q", got, "name")
	}
}

func TestActiveFieldSingleField(t *testing.T) {
	// Only one field, no commas, no newlines before it.
	inside := " id = "
	got := activeField(inside)
	if got != "id" {
		t.Errorf("activeField = %q, want %q", got, "id")
	}
}

func TestAnalyzeBraceNewlineSeparator(t *testing.T) {
	// Full struct literal without commas — activeKwarg should be
	// the field on the current line, not the first field.
	text := "Type {\n  id = 42\n  name = "
	cur := AnalyzeCursor(text, 2, 9)
	if !cur.InCall {
		t.Fatal("should be in call")
	}
	if cur.FuncName != "Type" {
		t.Errorf("func = %q, want %q", cur.FuncName, "Type")
	}
	if cur.ActiveKwarg != "name" {
		t.Errorf("activeKwarg = %q, want %q", cur.ActiveKwarg, "name")
	}
}

func TestExtractFieldNamesNewlines(t *testing.T) {
	// extractFieldNames should find all field names regardless of
	// whether commas or newlines separate them.
	inside := "\n  id = 42\n  hostname = \"web\"\n  "
	names := extractFieldNames(inside)
	expected := map[string]bool{"id": true, "hostname": true}
	for _, n := range names {
		delete(expected, n)
	}
	if len(expected) > 0 {
		t.Errorf("missing field names: %v (got %v)", expected, names)
	}
}
