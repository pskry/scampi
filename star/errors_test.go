// SPDX-License-Identifier: GPL-3.0-only

package star

import "testing"

func TestParseUnexpectedKwarg(t *testing.T) {
	tests := []struct {
		name        string
		msg         string
		wantField   string
		wantSuggest string
	}{
		{
			name:        "with suggestion",
			msg:         `copy: unexpected keyword argument "srcc" (did you mean src?)`,
			wantField:   "srcc",
			wantSuggest: "src",
		},
		{
			name:        "no suggestion",
			msg:         `copy: unexpected keyword argument "foo"`,
			wantField:   "foo",
			wantSuggest: "",
		},
		{
			name:        "no match",
			msg:         "something completely different",
			wantField:   "",
			wantSuggest: "",
		},
		{
			name:        "empty string",
			msg:         "",
			wantField:   "",
			wantSuggest: "",
		},
		{
			name:        "suggestion with multi-word func prefix",
			msg:         `target.ssh: unexpected keyword argument "hostt" (did you mean host?)`,
			wantField:   "hostt",
			wantSuggest: "host",
		},
		{
			name:        "marker present but no quotes around field",
			msg:         `unexpected keyword argument bare`,
			wantField:   "bare",
			wantSuggest: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, suggestion := parseUnexpectedKwarg(tt.msg)
			if field != tt.wantField {
				t.Errorf("field = %q, want %q", field, tt.wantField)
			}
			if suggestion != tt.wantSuggest {
				t.Errorf("suggestion = %q, want %q", suggestion, tt.wantSuggest)
			}
		})
	}
}
