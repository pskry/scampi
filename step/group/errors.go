// SPDX-License-Identifier: GPL-3.0-only

package group

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
		ID:     "builtin.group.InvalidState",
		Text:   `invalid state "{{.Got}}"`,
		Hint:   `expected one of: {{join ", " .Allowed}}`,
		Data:   e,
		Source: &e.Source,
	}
}

func (InvalidStateError) Severity() signal.Severity { return signal.Error }
func (InvalidStateError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }

// GroupCreateError is emitted when creating a group fails.
type GroupCreateError struct {
	Name   string
	Err    error
	Source spec.SourceSpan
}

func (e GroupCreateError) Error() string {
	return fmt.Sprintf("failed to create group %q: %v", e.Name, e.Err)
}

func (e GroupCreateError) Unwrap() error { return e.Err }

func (e GroupCreateError) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.group.CreateFailed",
		Text:   `failed to create group "{{.Name}}"`,
		Hint:   "check that the group name is valid and no conflicting group exists",
		Data:   e,
		Source: &e.Source,
	}
}

func (GroupCreateError) Severity() signal.Severity { return signal.Error }
func (GroupCreateError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }

// GroupDeleteError is emitted when deleting a group fails.
type GroupDeleteError struct {
	Name   string
	Err    error
	Source spec.SourceSpan
}

func (e GroupDeleteError) Error() string {
	return fmt.Sprintf("failed to delete group %q: %v", e.Name, e.Err)
}

func (e GroupDeleteError) Unwrap() error { return e.Err }

func (e GroupDeleteError) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.group.DeleteFailed",
		Text:   `failed to delete group "{{.Name}}"`,
		Hint:   "check that no users have this as their primary group",
		Data:   e,
		Source: &e.Source,
	}
}

func (GroupDeleteError) Severity() signal.Severity { return signal.Error }
func (GroupDeleteError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }
