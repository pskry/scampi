// SPDX-License-Identifier: GPL-3.0-only

package lsp

import (
	"testing"

	"go.lsp.dev/protocol"
)

func TestScanHexColors(t *testing.T) {
	src := []byte(`module main

let a = "#2a5a8e"
let b = "not a color"
let c = "#abcDEF"
let d = "#abc"
let e = "prefix #ff0000 suffix"
`)
	f, _ := Parse("t.scampi", src)
	got := scanHexColors(src, f)
	if len(got) != 2 {
		t.Fatalf("got %d colors, want 2 (one per full-string match)", len(got))
	}
	// Order matches AST walk order — declarations top to bottom.
	if got[0].Color.Red == 0 && got[0].Color.Green == 0 && got[0].Color.Blue == 0 {
		t.Errorf("first color came back zero — parsing failed")
	}
}

func TestHexFromColor(t *testing.T) {
	cases := []struct {
		c    protocol.Color
		want string
	}{
		{protocol.Color{Red: 0, Green: 0, Blue: 0, Alpha: 1}, "#000000"},
		{protocol.Color{Red: 1, Green: 1, Blue: 1, Alpha: 1}, "#ffffff"},
		{
			c: protocol.Color{
				Red:   float64(0x2a) / 255,
				Green: float64(0x5a) / 255,
				Blue:  float64(0x8e) / 255,
				Alpha: 1,
			},
			want: "#2a5a8e",
		},
	}
	for _, tc := range cases {
		got := hexFromColor(tc.c)
		if got != tc.want {
			t.Errorf("hexFromColor(%v) = %q, want %q", tc.c, got, tc.want)
		}
	}
}
