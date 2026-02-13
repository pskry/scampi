// SPDX-License-Identifier: GPL-3.0-only

package star

import (
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"

	"godoit.dev/doit/spec"
)

func posToSpan(pos syntax.Position) spec.SourceSpan {
	return spec.SourceSpan{
		Filename:  pos.Filename(),
		StartLine: int(pos.Line),
		StartCol:  int(pos.Col),
		EndLine:   int(pos.Line),
		EndCol:    int(pos.Col),
	}
}

// callSpan returns a SourceSpan for the call site of the current builtin.
// It looks at the second-to-last entry in the call stack (the caller).
func callSpan(thread *starlark.Thread) spec.SourceSpan {
	stack := thread.CallStack()
	if len(stack) < 2 {
		return spec.SourceSpan{}
	}
	return posToSpan(stack[len(stack)-2].Pos)
}

// kwargsFieldSpans produces a FieldSpan map from kwargs. Each field gets the
// call-site position. TODO: resolve per-kwarg source positions from the
// Starlark AST so diagnostics point at the offending field, not the call site.
func kwargsFieldSpans(pos spec.SourceSpan, names ...string) map[string]spec.FieldSpan {
	fields := make(map[string]spec.FieldSpan, len(names))
	for _, name := range names {
		fields[name] = spec.FieldSpan{Value: pos}
	}
	return fields
}
