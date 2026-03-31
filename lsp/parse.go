// SPDX-License-Identifier: GPL-3.0-only

package lsp

import (
	"go.lsp.dev/protocol"
	"go.starlark.net/syntax"
)

// fileOptions mirrors the options in star/eval.go — Set, While, and
// Recursion are enabled beyond core Starlark.
var fileOptions = &syntax.FileOptions{
	Set:       true,
	While:     true,
	Recursion: true,
}

// Parse parses Starlark source and returns the AST and any diagnostics.
// It uses the same FileOptions as the engine evaluator so parse errors
// match what the engine would report.
func Parse(filename string, content []byte) (*syntax.File, []protocol.Diagnostic) {
	f, err := fileOptions.Parse(filename, content, 0)
	if err != nil {
		return nil, syntaxErrors(err)
	}
	return f, nil
}

func syntaxErrors(err error) []protocol.Diagnostic {
	if errs, ok := err.(syntax.Error); ok {
		return []protocol.Diagnostic{syntaxDiag(errs)}
	}
	if errList, ok := err.(*syntax.Error); ok {
		return []protocol.Diagnostic{syntaxDiag(*errList)}
	}

	// Starlark returns a list of errors for parse failures.
	if el, ok := err.(interface{ Unwrap() []error }); ok {
		var diags []protocol.Diagnostic
		for _, e := range el.Unwrap() {
			if se, ok := e.(syntax.Error); ok {
				diags = append(diags, syntaxDiag(se))
			}
		}
		if len(diags) > 0 {
			return diags
		}
	}

	// Fallback: unknown error shape, report at line 0.
	return []protocol.Diagnostic{{
		Range:    protocol.Range{},
		Severity: protocol.DiagnosticSeverityError,
		Source:   "scampi",
		Message:  err.Error(),
	}}
}

func syntaxDiag(e syntax.Error) protocol.Diagnostic {
	// Starlark positions are 1-based; LSP positions are 0-based.
	line := uint32(0)
	if e.Pos.Line > 0 {
		line = uint32(e.Pos.Line - 1)
	}
	col := uint32(0)
	if e.Pos.Col > 0 {
		col = uint32(e.Pos.Col - 1)
	}

	return protocol.Diagnostic{
		Range: protocol.Range{
			Start: protocol.Position{Line: line, Character: col},
			End:   protocol.Position{Line: line, Character: col},
		},
		Severity: protocol.DiagnosticSeverityError,
		Source:   "scampi",
		Message:  e.Msg,
	}
}
