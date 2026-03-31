// SPDX-License-Identifier: GPL-3.0-only

package lsp

import (
	"strings"
	"unicode"
)

// CursorContext describes where the cursor is relative to known Starlark
// constructs. Used by completion, signature help, and hover.
type CursorContext struct {
	// InCall is true when the cursor is directly inside a function call's
	// parentheses (not nested inside a list/dict within that call).
	InCall bool

	// InList is true when the cursor is inside a list literal [...].
	// This includes lists nested inside calls, e.g. deploy(steps=[HERE]).
	InList bool

	// FuncName is the function being called (e.g. "copy", "target.ssh").
	// Set when InCall is true, or when InList is true and the list is a
	// kwarg value of a known call.
	FuncName string

	// PresentKwargs lists kwarg names already written in the current call.
	PresentKwargs []string

	// ActiveParam is the zero-based index of the current parameter,
	// estimated by counting commas before the cursor.
	ActiveParam int

	// WordUnderCursor is the identifier (or dotted identifier) the cursor
	// is on or immediately after.
	WordUnderCursor string
}

// AnalyzeCursor inspects the document text up to the cursor position and
// extracts context about what the user is typing.
func AnalyzeCursor(text string, line, col uint32) CursorContext {
	offset := offsetFromPosition(text, line, col)
	if offset < 0 || offset > len(text) {
		return CursorContext{}
	}

	ctx := CursorContext{
		WordUnderCursor: wordAtOffset(text, offset),
	}

	// Walk backward to find the innermost enclosing bracket. Track
	// nesting of (), [], and {} so we skip matched pairs.
	type bracket struct {
		char byte
		pos  int
	}
	var stack []bracket

	inStr := byte(0)
	for i := offset - 1; i >= 0; i-- {
		ch := text[i]

		// Simple string tracking (not escape-aware, good enough).
		if inStr != 0 {
			if ch == inStr {
				inStr = 0
			}
			continue
		}
		if ch == '"' || ch == '\'' {
			inStr = ch
			continue
		}

		switch ch {
		case ')', ']', '}':
			stack = append(stack, bracket{ch, i})
		case '(':
			if len(stack) > 0 && stack[len(stack)-1].char == ')' {
				stack = stack[:len(stack)-1]
			} else {
				return analyzeCallContext(text, i, offset, ctx)
			}
		case '[':
			if len(stack) > 0 && stack[len(stack)-1].char == ']' {
				stack = stack[:len(stack)-1]
			} else {
				return analyzeListContext(text, i, offset, ctx)
			}
		case '{':
			if len(stack) > 0 && stack[len(stack)-1].char == '}' {
				stack = stack[:len(stack)-1]
			}
		}
	}

	return ctx
}

func analyzeCallContext(text string, parenPos, offset int, ctx CursorContext) CursorContext {
	funcName := identBeforeOffset(text, parenPos)
	if funcName == "" {
		return ctx
	}

	ctx.InCall = true
	ctx.FuncName = funcName

	inside := text[parenPos+1 : offset]
	ctx.PresentKwargs = extractKwargNames(inside)
	ctx.ActiveParam = countTopLevelCommas(inside)

	return ctx
}

func analyzeListContext(text string, bracketPos, _ int, ctx CursorContext) CursorContext {
	ctx.InList = true

	// Check if this list is a kwarg value inside a call, e.g. steps=[HERE].
	// Scan backward from '[' to find '=' then the kwarg name, then the '('.
	pos := bracketPos - 1
	for pos >= 0 && text[pos] == ' ' {
		pos--
	}
	if pos >= 0 && text[pos] == '=' {
		// Found '=', now find the enclosing '(' to get the function name.
		// Walk backward from the '=' skipping the kwarg name.
		depth := 0
		for i := pos - 1; i >= 0; i-- {
			switch text[i] {
			case ')':
				depth++
			case '(':
				if depth > 0 {
					depth--
				} else {
					ctx.FuncName = identBeforeOffset(text, i)
					return ctx
				}
			}
		}
	}

	return ctx
}

// offsetFromPosition converts a 0-based line/col to a byte offset.
func offsetFromPosition(text string, line, col uint32) int {
	off := 0
	for l := uint32(0); l < line; l++ {
		idx := strings.IndexByte(text[off:], '\n')
		if idx < 0 {
			return len(text)
		}
		off += idx + 1
	}
	off += int(col)
	if off > len(text) {
		off = len(text)
	}
	return off
}

// wordAtOffset extracts the identifier (possibly dotted) at or just before offset.
func wordAtOffset(text string, offset int) string {
	if offset > len(text) {
		offset = len(text)
	}

	// Expand left.
	start := offset
	for start > 0 {
		r := rune(text[start-1])
		if r == '.' || r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) {
			start--
		} else {
			break
		}
	}

	// Expand right.
	end := offset
	for end < len(text) {
		r := rune(text[end])
		if r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) {
			end++
		} else {
			break
		}
	}

	if start == end {
		return ""
	}
	return text[start:end]
}

// identBeforeOffset extracts a (possibly dotted) identifier ending just before pos.
func identBeforeOffset(text string, pos int) string {
	end := pos
	for end > 0 && text[end-1] == ' ' {
		end--
	}
	start := end
	for start > 0 {
		r := rune(text[start-1])
		if r == '.' || r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) {
			start--
		} else {
			break
		}
	}
	if start == end {
		return ""
	}
	return text[start:end]
}

// extractKwargNames finds kwarg names (word=) in the text between ( and cursor.
func extractKwargNames(s string) []string {
	var names []string
	for {
		eq := strings.IndexByte(s, '=')
		if eq < 0 {
			break
		}
		// Check it's not == (comparison).
		if eq+1 < len(s) && s[eq+1] == '=' {
			s = s[eq+2:]
			continue
		}

		// Extract the identifier before the '='.
		nameEnd := eq
		for nameEnd > 0 && s[nameEnd-1] == ' ' {
			nameEnd--
		}
		nameStart := nameEnd
		for nameStart > 0 {
			r := rune(s[nameStart-1])
			if r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) {
				nameStart--
			} else {
				break
			}
		}
		if nameStart < nameEnd {
			names = append(names, s[nameStart:nameEnd])
		}
		s = s[eq+1:]
	}
	return names
}

// countTopLevelCommas counts commas not nested inside parens, brackets, or braces.
func countTopLevelCommas(s string) int {
	depth := 0
	count := 0
	inStr := rune(0)
	for _, r := range s {
		switch {
		case inStr != 0:
			if r == inStr {
				inStr = 0
			}
		case r == '"' || r == '\'':
			inStr = r
		case r == '(' || r == '[' || r == '{':
			depth++
		case r == ')' || r == ']' || r == '}':
			if depth > 0 {
				depth--
			}
		case r == ',' && depth == 0:
			count++
		}
	}
	return count
}
