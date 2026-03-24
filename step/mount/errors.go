// SPDX-License-Identifier: GPL-3.0-only

package mount

import (
	"fmt"

	"scampi.dev/scampi/diagnostic"
	"scampi.dev/scampi/diagnostic/event"
	"scampi.dev/scampi/spec"
)

type MissingFieldError struct {
	diagnostic.FatalError
	Field  string
	Source spec.SourceSpan
}

func (e MissingFieldError) Error() string {
	return fmt.Sprintf("mount: %s is required", e.Field)
}

func (e MissingFieldError) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.mount.MissingField",
		Text:   "mount: {{.Field}} is required",
		Data:   e,
		Source: &e.Source,
	}
}

type InvalidStateError struct {
	diagnostic.FatalError
	Got    string
	Source spec.SourceSpan
}

func (e InvalidStateError) Error() string {
	return fmt.Sprintf("invalid state %q", e.Got)
}

func (e InvalidStateError) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.mount.InvalidState",
		Text:   `invalid state "{{.Got}}"`,
		Hint:   `expected one of: "mounted", "unmounted", "absent"`,
		Data:   e,
		Source: &e.Source,
	}
}

type InvalidTypeError struct {
	diagnostic.FatalError
	Got    string
	Source spec.SourceSpan
}

func (e InvalidTypeError) Error() string {
	return fmt.Sprintf("unsupported filesystem type %q", e.Got)
}

func (e InvalidTypeError) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.mount.InvalidType",
		Text:   `unsupported filesystem type "{{.Got}}"`,
		Hint:   `supported types: nfs, nfs4, cifs, ceph, ext4, xfs, btrfs, tmpfs, glusterfs`,
		Data:   e,
		Source: &e.Source,
	}
}

type MountCommandError struct {
	diagnostic.FatalError
	Op     string
	Dest   string
	Stderr string
}

func (e MountCommandError) Error() string {
	return fmt.Sprintf("%s %s failed: %s", e.Op, e.Dest, e.Stderr)
}

func (e MountCommandError) EventTemplate() event.Template {
	return event.Template{
		ID:   "builtin.mount.CommandFailed",
		Text: "{{.Op}} {{.Dest}} failed",
		Hint: "{{.Stderr}}",
		Data: e,
	}
}

type MissingToolError struct {
	diagnostic.FatalError
	FsType string
	Source spec.SourceSpan
}

func (e MissingToolError) Error() string {
	return fmt.Sprintf("mount type %q requires tools not found on target", e.FsType)
}

func (e MissingToolError) EventTemplate() event.Template {
	return event.Template{
		ID:   "builtin.mount.MissingTool",
		Text: `mount type "{{.FsType}}" requires tools not found on target`,
		Hint: `{{if or (eq .FsType "nfs") (eq .FsType "nfs4")}}` +
			`add a pkg step for nfs-common (Debian/Ubuntu) or nfs-utils (RHEL/Fedora)` +
			`{{else if eq .FsType "cifs"}}add a pkg step for cifs-utils` +
			`{{else if eq .FsType "ceph"}}add a pkg step for ceph-common` +
			`{{else if eq .FsType "glusterfs"}}add a pkg step for glusterfs-client` +
			`{{else}}ensure the required filesystem tools are installed via a pkg step{{end}}`,
		Data:   e,
		Source: &e.Source,
	}
}
