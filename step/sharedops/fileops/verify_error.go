// SPDX-License-Identifier: GPL-3.0-only

package fileops

import (
	"fmt"

	"scampi.dev/scampi/diagnostic"
	"scampi.dev/scampi/diagnostic/event"
	"scampi.dev/scampi/signal"
	"scampi.dev/scampi/spec"
)

// VerifyError is returned when a verify command exits non-zero.
type VerifyError struct {
	Cmd      string
	Dest     string
	ExitCode int
	Stderr   string
	Source   spec.SourceSpan
}

func (e *VerifyError) Error() string {
	return fmt.Sprintf("verify %q failed (exit %d): %s", e.Cmd, e.ExitCode, e.Stderr)
}

func (e *VerifyError) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.VerifyFailed",
		Text:   `verify command failed (exit {{.ExitCode}}): {{.Cmd}}`,
		Hint:   `the content did not pass validation — {{.Dest}} was not modified`,
		Help:   "{{.Stderr}}",
		Data:   e,
		Source: &e.Source,
	}
}

func (*VerifyError) Severity() signal.Severity { return signal.Error }
func (*VerifyError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }
