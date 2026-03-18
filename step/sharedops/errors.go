// SPDX-License-Identifier: GPL-3.0-only

package sharedops

import (
	"errors"
	"fmt"

	"scampi.dev/scampi/diagnostic"
	"scampi.dev/scampi/diagnostic/event"
	"scampi.dev/scampi/signal"
	"scampi.dev/scampi/spec"
	"scampi.dev/scampi/target"
)

type UnknownUserError struct {
	User   string
	Source spec.SourceSpan
	Err    error
}

func (e UnknownUserError) Error() string {
	return fmt.Sprintf("unknown user %q: %v", e.User, e.Err)
}

func (e UnknownUserError) Unwrap() error {
	return e.Err
}

func (e UnknownUserError) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.UnknownUser",
		Text:   `unknown user "{{.User}}"`,
		Hint:   `create user "{{.User}}" with useradd or adduser before setting file owner`,
		Data:   e,
		Source: &e.Source,
	}
}

func (UnknownUserError) Severity() signal.Severity { return signal.Error }
func (UnknownUserError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }
func (e UnknownUserError) DeferredResource() spec.Resource {
	return spec.UserResource(e.User)
}

type UnknownGroupError struct {
	Group  string
	Source spec.SourceSpan
	Err    error
}

func (e UnknownGroupError) Error() string {
	return fmt.Sprintf("unknown group %q: %v", e.Group, e.Err)
}

func (e UnknownGroupError) Unwrap() error {
	return e.Err
}

func (e UnknownGroupError) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.UnknownGroup",
		Text:   `unknown group "{{.Group}}"`,
		Hint:   `create group "{{.Group}}" with groupadd or addgroup before setting file owner`,
		Data:   e,
		Source: &e.Source,
	}
}

func (UnknownGroupError) Severity() signal.Severity { return signal.Error }
func (UnknownGroupError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }
func (e UnknownGroupError) DeferredResource() spec.Resource {
	return spec.GroupResource(e.Group)
}

type PermissionDeniedError struct {
	Operation string
	Source    spec.SourceSpan
	Err       error
}

func (e PermissionDeniedError) Error() string {
	return fmt.Sprintf("%q: %v", e.Operation, e.Err)
}

func (e PermissionDeniedError) Unwrap() error {
	return e.Err
}

func (e PermissionDeniedError) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.PermissionDenied",
		Text:   `permission denied for operation "{{.Operation}}"`,
		Hint:   "run as root, or configure passwordless sudo/doas for the target user",
		Data:   e,
		Source: &e.Source,
	}
}

func (PermissionDeniedError) Severity() signal.Severity { return signal.Error }
func (PermissionDeniedError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }

type RelativePathError struct {
	Field  string
	Path   string
	Source spec.SourceSpan
}

func (e RelativePathError) Error() string {
	return fmt.Sprintf("relative path %q not allowed for %s", e.Path, e.Field)
}

func (e RelativePathError) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.RelativePath",
		Text:   `{{.Field}}: relative path not allowed`,
		Hint:   "target paths must be absolute (start with /)",
		Data:   e,
		Source: &e.Source,
	}
}

func (RelativePathError) Severity() signal.Severity { return signal.Error }
func (RelativePathError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }

// EscalationFailedError wraps a target.EscalationError with diagnostic metadata.
type EscalationFailedError struct {
	target.EscalationError
}

func (e EscalationFailedError) EventTemplate() event.Template {
	return event.Template{
		ID:   "target.EscalationFailed",
		Text: `{{.Tool}} {{.Op}} {{.Path}}: exit {{.ExitCode}}`,
		Hint: "the target user may lack passwordless sudo/doas",
		Help: "{{.Stderr}}",
		Data: e.EscalationError,
	}
}

func (EscalationFailedError) Severity() signal.Severity { return signal.Error }
func (EscalationFailedError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }

// EscalationMissingError wraps a target.NoEscalationError with diagnostic metadata.
type EscalationMissingError struct {
	target.NoEscalationError
}

func (e EscalationMissingError) EventTemplate() event.Template {
	return event.Template{
		ID:   "target.EscalationMissing",
		Text: `{{.Op}} {{.Path}}: no escalation tool found`,
		Hint: "install sudo or doas on the target, or run as root",
		Data: e.NoEscalationError,
	}
}

func (EscalationMissingError) Severity() signal.Severity { return signal.Error }
func (EscalationMissingError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }

// StagingFailedError wraps a target.StagingError with diagnostic metadata.
type StagingFailedError struct {
	target.StagingError
}

func (e StagingFailedError) EventTemplate() event.Template {
	return event.Template{
		ID:   "target.StagingFailed",
		Text: `failed to stage temp file for "{{.Path}}"`,
		Hint: "ensure /tmp is writable on the target",
		Data: e.StagingError,
	}
}

func (StagingFailedError) Severity() signal.Severity { return signal.Error }
func (StagingFailedError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }

// VerifyMissingPlaceholderError is raised at plan time when verify is set
// but does not contain the %s placeholder.
type VerifyMissingPlaceholderError struct {
	Cmd    string
	Source spec.SourceSpan
}

func (e VerifyMissingPlaceholderError) Error() string {
	return fmt.Sprintf("verify command %q must contain %%s placeholder", e.Cmd)
}

func (e VerifyMissingPlaceholderError) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.VerifyMissingPlaceholder",
		Text:   `verify command must contain %%s placeholder`,
		Hint:   `try: verify = "{{.Cmd}} %s"`,
		Data:   e,
		Source: &e.Source,
	}
}

func (VerifyMissingPlaceholderError) Severity() signal.Severity { return signal.Error }
func (VerifyMissingPlaceholderError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }

// DiagnoseTargetError wraps known target-layer errors in diagnostic types.
// Returns the original error unchanged if not a recognized target error.
func DiagnoseTargetError(err error) error {
	var noEsc target.NoEscalationError
	if errors.As(err, &noEsc) {
		return EscalationMissingError{NoEscalationError: noEsc}
	}
	var escErr target.EscalationError
	if errors.As(err, &escErr) {
		return EscalationFailedError{EscalationError: escErr}
	}
	var stagErr target.StagingError
	if errors.As(err, &stagErr) {
		return StagingFailedError{StagingError: stagErr}
	}
	return err
}
