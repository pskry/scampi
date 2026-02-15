// SPDX-License-Identifier: GPL-3.0-only

package service

import (
	"fmt"

	"godoit.dev/doit/diagnostic"
	"godoit.dev/doit/diagnostic/event"
	"godoit.dev/doit/signal"
	"godoit.dev/doit/spec"
)

// InvalidStateError is raised when the state field has an unrecognized value.
type InvalidStateError struct {
	Got     string
	Allowed []string
	Source  spec.SourceSpan
}

func (e InvalidStateError) Error() string {
	return fmt.Sprintf("invalid state %q", e.Got)
}

func (e InvalidStateError) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.service.InvalidState",
		Text:   `invalid state "{{.Got}}"`,
		Hint:   `expected one of: {{join ", " .Allowed}}`,
		Data:   e,
		Source: &e.Source,
	}
}

func (InvalidStateError) Severity() signal.Severity { return signal.Error }
func (InvalidStateError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }

// ServiceCommandError is emitted when a service command (start/stop/enable/disable) fails.
type ServiceCommandError struct {
	Op     string
	Name   string
	Stderr string
	Source spec.SourceSpan
}

func (e ServiceCommandError) Error() string {
	return fmt.Sprintf("failed to %s service %s", e.Op, e.Name)
}

func (e ServiceCommandError) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.service.CommandFailed",
		Text:   `failed to {{.Op}} service {{.Name}}: {{.Stderr}}`,
		Hint:   "check that the service name is correct and the init system is available",
		Help:   "the service command exited with a non-zero status",
		Data:   e,
		Source: &e.Source,
	}
}

func (ServiceCommandError) Severity() signal.Severity { return signal.Error }
func (ServiceCommandError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }

// DaemonReloadError is emitted when daemon-reload fails.
type DaemonReloadError struct {
	Name   string
	Stderr string
	Source spec.SourceSpan
}

func (e DaemonReloadError) Error() string {
	return fmt.Sprintf("daemon-reload failed before starting service %s", e.Name)
}

func (e DaemonReloadError) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.service.DaemonReloadFailed",
		Text:   `daemon-reload failed before starting service {{.Name}}: {{.Stderr}}`,
		Hint:   "check systemd configuration and permissions",
		Data:   e,
		Source: &e.Source,
	}
}

func (DaemonReloadError) Severity() signal.Severity { return signal.Error }
func (DaemonReloadError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }
