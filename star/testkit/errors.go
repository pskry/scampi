// SPDX-License-Identifier: GPL-3.0-only

package testkit

import (
	"scampi.dev/scampi/diagnostic"
	"scampi.dev/scampi/diagnostic/event"
)

// TestPass is emitted when an assertion passes.
type TestPass struct {
	diagnostic.Info
	Description string
}

func (e TestPass) EventTemplate() event.Template {
	return event.Template{
		ID:   "test.Pass",
		Text: "{{.Description}}",
		Data: e,
	}
}

// TestFail is emitted when an assertion fails.
type TestFail struct {
	diagnostic.FatalError
	Description string
	Expected    string
	Actual      string
}

func (e TestFail) EventTemplate() event.Template {
	return event.Template{
		ID:   "test.Fail",
		Text: "{{.Description}}",
		Hint: "expected: {{.Expected}}\nactual:   {{.Actual}}",
		Data: e,
	}
}

// TestSummary is emitted at the end of a test run.
type TestSummary struct {
	diagnostic.Info
	File   string
	Passed int
	Failed int
}

func (e TestSummary) EventTemplate() event.Template {
	return event.Template{
		ID:   "test.Summary",
		Text: "{{.File}}: {{.Passed}} passed, {{.Failed}} failed",
		Data: e,
	}
}

// TestError is emitted when an assertion cannot be evaluated (infrastructure error).
type TestError struct {
	diagnostic.FatalError
	Detail string
	Hint   string
}

func (e TestError) EventTemplate() event.Template {
	return event.Template{
		ID:   "test.Error",
		Text: "{{.Detail}}",
		Hint: "{{.Hint}}",
		Data: e,
	}
}
