// SPDX-License-Identifier: GPL-3.0-only

package lsp

import (
	"fmt"
	"regexp"
	"strconv"

	"go.lsp.dev/protocol"

	"scampi.dev/scampi/lang/ast"
)

// hexColorRe matches scampi string literals whose entire content is
// a `#RRGGBB` color. We deliberately don't accept short-form `#RGB`
// or 8-digit `#RRGGBBAA` for v1 — most config sites are 6-digit and
// the editor's color picker writes full 6-digit anyway.
var hexColorRe = regexp.MustCompile(`^#([0-9a-fA-F]{6})$`)

// scanHexColors walks the AST and returns a ColorInformation for
// every string literal whose contents match `#RRGGBB`. The Range
// covers the literal's inner text (the bytes between the quotes),
// which is what the editor highlights with the color swatch.
func scanHexColors(src []byte, f *ast.File) []protocol.ColorInformation {
	if f == nil {
		return nil
	}
	var out []protocol.ColorInformation
	ast.Walk(f, func(n ast.Node) bool {
		sl, ok := n.(*ast.StringLit)
		if !ok || len(sl.Parts) != 1 {
			return true
		}
		text, ok := sl.Parts[0].(*ast.StringText)
		if !ok {
			return true
		}
		m := hexColorRe.FindStringSubmatch(text.Raw)
		if m == nil {
			return true
		}
		r, _ := strconv.ParseUint(m[1][0:2], 16, 8)
		g, _ := strconv.ParseUint(m[1][2:4], 16, 8)
		b, _ := strconv.ParseUint(m[1][4:6], 16, 8)
		out = append(out, protocol.ColorInformation{
			Range: tokenSpanToRange(src, text.SrcSpan),
			Color: protocol.Color{
				Red:   float64(r) / 255,
				Green: float64(g) / 255,
				Blue:  float64(b) / 255,
				Alpha: 1.0,
			},
		})
		return true
	}, nil)
	return out
}

// hexFromColor renders an LSP Color back to the `#RRGGBB` string the
// editor will substitute when the user picks a new color from the
// inline picker. Alpha is dropped — we don't accept 8-digit input.
func hexFromColor(c protocol.Color) string {
	r := uint8(c.Red*255 + 0.5)
	g := uint8(c.Green*255 + 0.5)
	b := uint8(c.Blue*255 + 0.5)
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}
