// SPDX-License-Identifier: GPL-3.0-only

package user

import (
	"fmt"
	"strings"

	"scampi.dev/scampi/diagnostic"
	"scampi.dev/scampi/diagnostic/event"
	"scampi.dev/scampi/signal"
	"scampi.dev/scampi/spec"
)

// InvalidStateError is raised when the state field has an unrecognized value.
type InvalidStateError struct {
	Got     string
	Allowed []string
	Source  spec.SourceSpan
}

func (e InvalidStateError) Error() string {
	return fmt.Sprintf("invalid state %q (expected one of: %s)", e.Got, strings.Join(e.Allowed, ", "))
}

func (e InvalidStateError) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.user.InvalidState",
		Text:   `invalid state "{{.Got}}"`,
		Hint:   `expected one of: {{join ", " .Allowed}}`,
		Data:   e,
		Source: &e.Source,
	}
}

func (InvalidStateError) Severity() signal.Severity { return signal.Error }
func (InvalidStateError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }

// UserCreateError is emitted when creating a user fails.
type UserCreateError struct {
	Name   string
	Err    error
	Source spec.SourceSpan
}

func (e UserCreateError) Error() string {
	return fmt.Sprintf("failed to create user %q: %v", e.Name, e.Err)
}

func (e UserCreateError) Unwrap() error { return e.Err }

func (e UserCreateError) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.user.CreateFailed",
		Text:   `failed to create user "{{.Name}}"`,
		Hint:   "check that the username is valid and no conflicting user exists",
		Data:   e,
		Source: &e.Source,
	}
}

func (UserCreateError) Severity() signal.Severity { return signal.Error }
func (UserCreateError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }

// UserModifyError is emitted when modifying a user fails.
type UserModifyError struct {
	Name   string
	Err    error
	Source spec.SourceSpan
}

func (e UserModifyError) Error() string {
	return fmt.Sprintf("failed to modify user %q: %v", e.Name, e.Err)
}

func (e UserModifyError) Unwrap() error { return e.Err }

func (e UserModifyError) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.user.ModifyFailed",
		Text:   `failed to modify user "{{.Name}}"`,
		Hint:   "check that the user exists and the target values are valid",
		Data:   e,
		Source: &e.Source,
	}
}

func (UserModifyError) Severity() signal.Severity { return signal.Error }
func (UserModifyError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }

// UserDeleteError is emitted when deleting a user fails.
type UserDeleteError struct {
	Name   string
	Err    error
	Source spec.SourceSpan
}

func (e UserDeleteError) Error() string {
	return fmt.Sprintf("failed to delete user %q: %v", e.Name, e.Err)
}

func (e UserDeleteError) Unwrap() error { return e.Err }

func (e UserDeleteError) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.user.DeleteFailed",
		Text:   `failed to delete user "{{.Name}}"`,
		Hint:   "check that no running processes belong to this user",
		Data:   e,
		Source: &e.Source,
	}
}

func (UserDeleteError) Severity() signal.Severity { return signal.Error }
func (UserDeleteError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }
