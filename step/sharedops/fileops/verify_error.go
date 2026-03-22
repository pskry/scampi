// SPDX-License-Identifier: GPL-3.0-only

package fileops

import (
	"errors"
	"fmt"

	"scampi.dev/scampi/diagnostic"
	"scampi.dev/scampi/diagnostic/event"
	"scampi.dev/scampi/spec"
	"scampi.dev/scampi/target"
)

// VerifyError is returned when a verify command exits non-zero.
type VerifyError struct {
	diagnostic.FatalError
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

// VerifyIOError is returned when verify infrastructure (temp dirs, temp files,
// running the verify command) fails due to I/O errors.
type VerifyIOError struct {
	diagnostic.FatalError
	Op     string
	Err    error
	Advice string
}

func (e VerifyIOError) Error() string {
	return e.Op + ": " + e.Err.Error()
}

func (e VerifyIOError) Unwrap() error { return e.Err }

func (e VerifyIOError) EventTemplate() event.Template {
	return event.Template{
		ID:   "builtin.VerifyIOError",
		Text: "verify failed: {{.Op}}",
		Hint: "{{.Advice}}",
		Help: "{{.Err}}",
		Data: e,
	}
}

func newVerifyIOError(op string, err error) VerifyIOError {
	return VerifyIOError{Op: op, Err: err, Advice: verifyIOAdvice(err)}
}

func verifyIOAdvice(err error) string {
	switch {
	case errors.Is(err, target.ErrPermission):
		return "the connecting user lacks write permission on the target — check ownership with ls -la"
	case errors.Is(err, target.ErrNotExist):
		return "path does not exist on target — check that parent directories are present"
	case errors.Is(err, target.ErrCommandNotFound):
		return "verify command not found — ensure it is installed on the target"
	default:
		return "check target filesystem permissions and connectivity"
	}
}
