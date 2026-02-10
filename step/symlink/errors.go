package symlink

import (
	"fmt"

	"godoit.dev/doit/diagnostic"
	"godoit.dev/doit/diagnostic/event"
	"godoit.dev/doit/signal"
	"godoit.dev/doit/spec"
)

type LinkDirMissingError struct {
	Path   string
	Source spec.SourceSpan
	Err    error
}

func (e LinkDirMissingError) Error() string {
	return fmt.Sprintf("link directory %q does not exist", e.Path)
}

func (e LinkDirMissingError) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.symlink.LinkDirMissing",
		Text:   `link directory "{{.Path}}" does not exist`,
		Hint:   "create the parent directory before creating the symlink",
		Help:   "the symlink action does not create directories automatically",
		Data:   e,
		Source: &e.Source,
	}
}

func (LinkDirMissingError) Severity() signal.Severity { return signal.Error }
func (LinkDirMissingError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }

type LinkReadError struct {
	Path   string
	Source spec.SourceSpan
	Err    error
}

func (e LinkReadError) Error() string {
	return fmt.Sprintf("cannot read link %q: %v", e.Path, e.Err)
}

func (e LinkReadError) Unwrap() error {
	return e.Err
}

func (e LinkReadError) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.symlink.LinkReadError",
		Text:   `cannot read link "{{.Path}}"`,
		Hint:   "check file permissions and ensure the path is accessible",
		Data:   e,
		Source: &e.Source,
	}
}

func (LinkReadError) Severity() signal.Severity { return signal.Error }
func (LinkReadError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }

type NotASymlinkError struct {
	Path   string
	Source spec.SourceSpan
}

func (e NotASymlinkError) Error() string {
	return fmt.Sprintf("path %q exists but is not a symlink", e.Path)
}

func (e NotASymlinkError) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.symlink.NotASymlink",
		Text:   `path "{{.Path}}" exists but is not a symlink`,
		Hint:   "remove the existing file or directory before creating the symlink",
		Help:   "the symlink action will not overwrite existing files or directories for safety",
		Data:   e,
		Source: &e.Source,
	}
}

func (NotASymlinkError) Severity() signal.Severity { return signal.Error }
func (NotASymlinkError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }
