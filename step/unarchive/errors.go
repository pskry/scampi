// SPDX-License-Identifier: GPL-3.0-only

package unarchive

import (
	"fmt"
	"strings"

	"scampi.dev/scampi/diagnostic"
	"scampi.dev/scampi/diagnostic/event"
	"scampi.dev/scampi/signal"
	"scampi.dev/scampi/spec"
)

// UnsupportedArchiveError is raised at plan time for unknown archive extensions.
type UnsupportedArchiveError struct {
	Path   string
	Source spec.SourceSpan
}

func (e UnsupportedArchiveError) Error() string {
	return fmt.Sprintf("unsupported archive format: %q", e.Path)
}

func (e UnsupportedArchiveError) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.unarchive.UnsupportedArchive",
		Text:   `unsupported archive format: "{{.Path}}"`,
		Hint:   "supported: .tar.gz, .tgz, .tar.bz2, .tbz2, .tar.xz, .txz, .tar.zst, .tzst, .tar, .zip",
		Data:   e,
		Source: &e.Source,
	}
}

func (UnsupportedArchiveError) Severity() signal.Severity { return signal.Error }
func (UnsupportedArchiveError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }

// ArchiveNotFoundError is raised when the source archive does not exist.
type ArchiveNotFoundError struct {
	Path   string
	Source spec.SourceSpan
	Err    error
}

func (e ArchiveNotFoundError) Error() string {
	return fmt.Sprintf("source archive %q does not exist", e.Path)
}

func (e ArchiveNotFoundError) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.unarchive.ArchiveNotFound",
		Text:   `source archive "{{.Path}}" does not exist`,
		Hint:   "ensure the archive file exists and is readable",
		Data:   e,
		Source: &e.Source,
	}
}

func (ArchiveNotFoundError) Severity() signal.Severity { return signal.Error }
func (ArchiveNotFoundError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }

// ExtractionError is raised when the extract command fails.
type ExtractionError struct {
	Cmd    string
	Stderr string
	Advice string
	Source spec.SourceSpan
}

func (e ExtractionError) Error() string {
	return fmt.Sprintf("extraction failed: %s: %s", e.Cmd, e.Stderr)
}

func (e ExtractionError) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.unarchive.ExtractionFailed",
		Text:   `extraction failed: {{.Cmd}}`,
		Hint:   "{{.Advice}}",
		Help:   "{{.Stderr}}",
		Data:   e,
		Source: &e.Source,
	}
}

func extractionAdvice(stderr string) string {
	lower := strings.ToLower(stderr)
	if strings.Contains(lower, "cannot open") ||
		strings.Contains(lower, "permission denied") ||
		strings.Contains(lower, "operation not permitted") {
		return "target contains files not writable by the connecting user — check ownership with ls -la"
	}
	return "check that the archive is not corrupt and required tools are installed"
}

func (ExtractionError) Severity() signal.Severity { return signal.Error }
func (ExtractionError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }

// PartialOwnershipError is raised at plan time when owner is set without group
// or vice versa.
type PartialOwnershipError struct {
	Set     string
	Missing string
	Source  spec.SourceSpan
}

func (e PartialOwnershipError) Error() string {
	return fmt.Sprintf("%s is set but %s is empty — set both or neither", e.Set, e.Missing)
}

func (e PartialOwnershipError) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.unarchive.PartialOwnership",
		Text:   `{{.Set}} is set but {{.Missing}} is empty`,
		Hint:   `add {{.Missing}}="<value>" or remove {{.Set}}`,
		Data:   e,
		Source: &e.Source,
	}
}

func (PartialOwnershipError) Severity() signal.Severity { return signal.Error }
func (PartialOwnershipError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }
