// SPDX-License-Identifier: GPL-3.0-only

// Package star evaluates Starlark configuration files into spec.Config.
package star

import (
	"fmt"

	"go.starlark.net/starlark"

	"godoit.dev/doit/diagnostic"
	"godoit.dev/doit/diagnostic/event"
	"godoit.dev/doit/signal"
	"godoit.dev/doit/spec"
)

// StarlarkError wraps a Starlark evaluation error with source position.
type StarlarkError struct {
	Err    error
	Source spec.SourceSpan
}

func (e StarlarkError) Error() string { return e.Err.Error() }
func (e StarlarkError) Unwrap() error { return e.Err }

func (e StarlarkError) EventTemplate() event.Template {
	return event.Template{
		ID:     "star.EvalError",
		Text:   "{{.Err}}",
		Data:   e,
		Source: &e.Source,
	}
}

func (e StarlarkError) Severity() signal.Severity {
	return signal.Error
}

func (e StarlarkError) Impact() diagnostic.Impact {
	return diagnostic.ImpactAbort
}

// wrapStarlarkError converts a starlark.EvalError into a StarlarkError with
// source position extracted from the backtrace, or wraps any other error.
func wrapStarlarkError(err error) StarlarkError {
	var evalErr *starlark.EvalError
	if ok := isEvalError(err, &evalErr); ok && evalErr != nil {
		bt := evalErr.CallStack
		if len(bt) > 0 {
			return StarlarkError{
				Err:    err,
				Source: posToSpan(bt[len(bt)-1].Pos),
			}
		}
	}
	return StarlarkError{Err: err}
}

func isEvalError(err error, target **starlark.EvalError) bool {
	for err != nil {
		if ee, ok := err.(*starlark.EvalError); ok {
			*target = ee
			return true
		}
		if u, ok := err.(interface{ Unwrap() error }); ok {
			err = u.Unwrap()
		} else {
			return false
		}
	}
	return false
}

// DuplicateTargetError is raised when a target name is registered twice.
type DuplicateTargetError struct {
	Name   string
	Source spec.SourceSpan
}

func (e DuplicateTargetError) Error() string {
	return fmt.Sprintf("duplicate target %q", e.Name)
}

func (e DuplicateTargetError) EventTemplate() event.Template {
	return event.Template{
		ID:     "star.DuplicateTarget",
		Text:   `duplicate target "{{.Name}}"`,
		Hint:   "each target name must be unique",
		Data:   e,
		Source: &e.Source,
	}
}

func (e DuplicateTargetError) Severity() signal.Severity {
	return signal.Error
}

func (e DuplicateTargetError) Impact() diagnostic.Impact {
	return diagnostic.ImpactAbort
}

// DuplicateDeployError is raised when a deploy block name is registered twice.
type DuplicateDeployError struct {
	Name   string
	Source spec.SourceSpan
}

func (e DuplicateDeployError) Error() string {
	return fmt.Sprintf("duplicate deploy block %q", e.Name)
}

func (e DuplicateDeployError) EventTemplate() event.Template {
	return event.Template{
		ID:     "star.DuplicateDeploy",
		Text:   `duplicate deploy block "{{.Name}}"`,
		Hint:   "each deploy block name must be unique",
		Data:   e,
		Source: &e.Source,
	}
}

func (e DuplicateDeployError) Severity() signal.Severity {
	return signal.Error
}

func (e DuplicateDeployError) Impact() diagnostic.Impact {
	return diagnostic.ImpactAbort
}

// MissingArgError is raised when a required argument is not provided.
type MissingArgError struct {
	Func   string
	Arg    string
	Source spec.SourceSpan
}

func (e MissingArgError) Error() string {
	return fmt.Sprintf("%s() missing required argument %q", e.Func, e.Arg)
}

func (e MissingArgError) EventTemplate() event.Template {
	return event.Template{
		ID:     "star.MissingArg",
		Text:   `{{.Func}}() missing required argument "{{.Arg}}"`,
		Data:   e,
		Source: &e.Source,
	}
}

func (e MissingArgError) Severity() signal.Severity {
	return signal.Error
}

func (e MissingArgError) Impact() diagnostic.Impact {
	return diagnostic.ImpactAbort
}

// EnvVarRequiredError is raised when a required env var is unset.
type EnvVarRequiredError struct {
	Key    string
	Source spec.SourceSpan
}

func (e EnvVarRequiredError) Error() string {
	return fmt.Sprintf("required environment variable %q is not set", e.Key)
}

func (e EnvVarRequiredError) EventTemplate() event.Template {
	return event.Template{
		ID:     "star.EnvVarRequired",
		Text:   `required environment variable "{{.Key}}" is not set`,
		Hint:   "set the variable or provide a default value",
		Data:   e,
		Source: &e.Source,
	}
}

func (e EnvVarRequiredError) Severity() signal.Severity {
	return signal.Error
}

func (e EnvVarRequiredError) Impact() diagnostic.Impact {
	return diagnostic.ImpactAbort
}
