// SPDX-License-Identifier: GPL-3.0-only

package lxc

import (
	"fmt"

	"scampi.dev/scampi/diagnostic"
	"scampi.dev/scampi/diagnostic/event"
	"scampi.dev/scampi/spec"
)

type InvalidConfigError struct {
	diagnostic.FatalError
	Field  string
	Reason string
	Source spec.SourceSpan
}

func (e InvalidConfigError) Error() string {
	return fmt.Sprintf("invalid pve.lxc config: %s: %s", e.Field, e.Reason)
}

func (e InvalidConfigError) EventTemplate() event.Template {
	return event.Template{
		ID:     CodeInvalidConfig,
		Text:   `invalid pve.lxc config: {{.Field}}`,
		Hint:   "{{.Reason}}",
		Data:   e,
		Source: &e.Source,
	}
}

type CommandFailedError struct {
	diagnostic.FatalError
	Op     string
	VMID   int
	Stderr string
	Source spec.SourceSpan
}

func (e CommandFailedError) Error() string {
	return fmt.Sprintf("pve.lxc %s VMID %d failed: %s", e.Op, e.VMID, e.Stderr)
}

func (e CommandFailedError) EventTemplate() event.Template {
	return event.Template{
		ID:     CodeCommandFailed,
		Text:   `pve.lxc {{.Op}} VMID {{.VMID}} failed`,
		Help:   "{{.Stderr}}",
		Data:   e,
		Source: &e.Source,
	}
}

type TemplateNotFoundError struct {
	diagnostic.FatalError
	Template string
	Storage  string
	Source   spec.SourceSpan
}

func (e TemplateNotFoundError) Error() string {
	return fmt.Sprintf("template %q not found on storage %q and not available for download", e.Template, e.Storage)
}

func (e TemplateNotFoundError) EventTemplate() event.Template {
	return event.Template{
		ID:   CodeTemplateNotFound,
		Text: `template "{{.Template}}" not found`,
		Hint: `not on storage "{{.Storage}}" and not in pveam available — check the template name or upload it manually`,
		Data: e,
		Source: &e.Source,
	}
}

type UnsupportedStateError struct {
	diagnostic.FatalError
	State  string
	Source spec.SourceSpan
}

func (e UnsupportedStateError) Error() string {
	return fmt.Sprintf("pve.lxc state %q is not yet supported", e.State)
}

func (e UnsupportedStateError) EventTemplate() event.Template {
	return event.Template{
		ID:     CodeUnsupportedState,
		Text:   `pve.lxc state "{{.State}}" is not yet supported`,
		Hint:   "supported states: running, stopped",
		Data:   e,
		Source: &e.Source,
	}
}
