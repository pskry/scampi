package render

import (
	"regexp"

	"godoit.dev/doit/diagnostic/event"
	"godoit.dev/doit/spec"
)

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

type Template struct {
	Name string
	Text string
	Hint string
	Help string

	Data any

	Source *spec.SourceSpan
}

type RunSummary struct {
	ChangedCount int
	FailedCount  int
	TotalCount   int
}

type Displayer interface {
	Emit(e event.Event)
	Close()
}

func s(n int) string {
	if n == 1 {
		return ""
	}

	return "s"
}

func visibleLen(s string) int {
	return len(ansiRe.ReplaceAllString(s, ""))
}

func elide(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if visibleLen(s) <= maxLen {
		return s
	}
	if maxLen <= 1 {
		return "…"
	}
	return s[:maxLen-1] + "…"
}

func fitLine(s string, width int) string {
	if width <= 0 {
		return s
	}
	if visibleLen(s) <= width {
		return s
	}
	return elide(s, width)
}
