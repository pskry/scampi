// SPDX-License-Identifier: GPL-3.0-only

package container

import (
	"fmt"
	"strings"

	"scampi.dev/scampi/diagnostic"
	"scampi.dev/scampi/diagnostic/event"
	"scampi.dev/scampi/spec"
)

type InvalidStateError struct {
	diagnostic.FatalError
	Got     string
	Allowed []string
	Source  spec.SourceSpan
}

func (e InvalidStateError) Error() string {
	return fmt.Sprintf("invalid container state %q (allowed: %s)", e.Got, strings.Join(e.Allowed, ", "))
}

func (e InvalidStateError) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.container.InvalidState",
		Text:   `invalid container state "{{.Got}}"`,
		Hint:   `allowed: running, stopped, absent`,
		Data:   e,
		Source: &e.Source,
	}
}

type InvalidRestartError struct {
	diagnostic.FatalError
	Got     string
	Allowed []string
	Source  spec.SourceSpan
}

func (e InvalidRestartError) Error() string {
	return fmt.Sprintf("invalid restart policy %q (allowed: %s)", e.Got, strings.Join(e.Allowed, ", "))
}

func (e InvalidRestartError) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.container.InvalidRestart",
		Text:   `invalid restart policy "{{.Got}}"`,
		Hint:   `allowed: always, on-failure, unless-stopped, no`,
		Data:   e,
		Source: &e.Source,
	}
}

type EmptyImageError struct {
	diagnostic.FatalError
	Source spec.SourceSpan
}

func (e EmptyImageError) Error() string {
	return "container image is required when state is not absent"
}

func (e EmptyImageError) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.container.EmptyImage",
		Text:   "container image is required",
		Hint:   `add image = "registry/name:tag"`,
		Data:   e,
		Source: &e.Source,
	}
}

type ContainerCommandError struct {
	diagnostic.FatalError
	Op     string
	Name   string
	Stderr string
	Source spec.SourceSpan
}

func (e ContainerCommandError) Error() string {
	return fmt.Sprintf("container %s %q failed: %s", e.Op, e.Name, e.Stderr)
}

func (e ContainerCommandError) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.container.CommandFailed",
		Text:   `container {{.Op}} "{{.Name}}" failed`,
		Help:   "{{.Stderr}}",
		Data:   e,
		Source: &e.Source,
	}
}
