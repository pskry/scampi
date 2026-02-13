// SPDX-License-Identifier: GPL-3.0-only

package dir

import (
	"fmt"

	"godoit.dev/doit/diagnostic"
	"godoit.dev/doit/diagnostic/event"
	"godoit.dev/doit/signal"
	"godoit.dev/doit/spec"
)

type NotADirectoryError struct {
	Path   string
	Source spec.SourceSpan
}

func (e NotADirectoryError) Error() string {
	return fmt.Sprintf("path %q exists but is not a directory", e.Path)
}

func (e NotADirectoryError) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.dir.NotADirectory",
		Text:   `path "{{.Path}}" exists but is not a directory`,
		Hint:   "the path exists but is not a directory",
		Data:   e,
		Source: &e.Source,
	}
}

func (NotADirectoryError) Severity() signal.Severity { return signal.Error }
func (NotADirectoryError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }
