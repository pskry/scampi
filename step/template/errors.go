package template

import (
	"fmt"

	"godoit.dev/doit/diagnostic"
	"godoit.dev/doit/diagnostic/event"
	"godoit.dev/doit/signal"
	"godoit.dev/doit/spec"
)

type EnvKeyNotInValues struct {
	EnvVar string
	Key    string
	Source spec.SourceSpan
}

func (e EnvKeyNotInValues) Error() string {
	return fmt.Sprintf("env var %q maps to key %q which is not defined in values", e.EnvVar, e.Key)
}

func (e EnvKeyNotInValues) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.template.EnvKeyNotInValues",
		Text:   `env var "{{.EnvVar}}" maps to key "{{.Key}}" which is not defined in values`,
		Hint:   "add the key to data.values or remove the env mapping",
		Help:   "all env mappings must reference keys that exist in data.values",
		Data:   e,
		Source: &e.Source,
	}
}

func (EnvKeyNotInValues) Severity() signal.Severity { return signal.Error }
func (EnvKeyNotInValues) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }

type TemplateSourceMissing struct {
	Path   string
	Source spec.SourceSpan
	Err    error
}

func (e TemplateSourceMissing) Error() string {
	return fmt.Sprintf("template source %q does not exist", e.Path)
}

func (e TemplateSourceMissing) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.template.SourceMissing",
		Text:   `template source "{{.Path}}" does not exist`,
		Hint:   "ensure the template file exists and is readable",
		Help:   "the template action cannot proceed without a readable source file",
		Data:   e,
		Source: &e.Source,
	}
}

func (TemplateSourceMissing) Severity() signal.Severity { return signal.Error }
func (TemplateSourceMissing) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }

type TemplateParseError struct {
	Err    error
	Source spec.SourceSpan
}

func (e TemplateParseError) Error() string {
	return fmt.Sprintf("template parse error: %v", e.Err)
}

func (e TemplateParseError) Unwrap() error {
	return e.Err
}

func (e TemplateParseError) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.template.ParseError",
		Text:   "template parse error: {{.Err}}",
		Hint:   "check template syntax",
		Help:   "templates use Go text/template syntax",
		Data:   e,
		Source: &e.Source,
	}
}

func (TemplateParseError) Severity() signal.Severity { return signal.Error }
func (TemplateParseError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }

type TemplateExecError struct {
	Err    error
	Source spec.SourceSpan
}

func (e TemplateExecError) Error() string {
	return fmt.Sprintf("template execution error: %v", e.Err)
}

func (e TemplateExecError) Unwrap() error {
	return e.Err
}

func (e TemplateExecError) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.template.ExecError",
		Text:   "template execution error: {{.Err}}",
		Hint:   "check that all referenced variables exist in data",
		Help:   "template execution failed, usually due to missing or mistyped variable names",
		Data:   e,
		Source: &e.Source,
	}
}

func (TemplateExecError) Severity() signal.Severity { return signal.Error }
func (TemplateExecError) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }

type DestDirMissing struct {
	Path   string
	Source spec.SourceSpan
	Err    error
}

func (e DestDirMissing) Error() string {
	return fmt.Sprintf("destination directory %q does not exist", e.Path)
}

func (e DestDirMissing) EventTemplate() event.Template {
	return event.Template{
		ID:     "builtin.template.DestDirMissing",
		Text:   `destination directory "{{.Path}}" does not exist`,
		Hint:   "create the destination directory before running this action",
		Help:   "the template action does not create directories automatically",
		Data:   e,
		Source: &e.Source,
	}
}

func (DestDirMissing) Severity() signal.Severity { return signal.Error }
func (DestDirMissing) Impact() diagnostic.Impact { return diagnostic.ImpactAbort }
