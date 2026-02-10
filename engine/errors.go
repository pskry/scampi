package engine

import (
	"errors"
	"fmt"
	"runtime"

	"godoit.dev/doit/capability"
	"godoit.dev/doit/diagnostic"
	"godoit.dev/doit/diagnostic/event"
	"godoit.dev/doit/errs"
	"godoit.dev/doit/signal"
	"godoit.dev/doit/spec"
)

type AbortError struct {
	Causes []error
}

func (AbortError) Error() string {
	return "execution aborted"
}

type CapabilityMismatchError struct {
	StepIndex    int
	StepKind     string
	RequiredCaps capability.Capability
	MissingCaps  capability.Capability
	ProvidedCaps capability.Capability
	Source       spec.SourceSpan
}

func (e CapabilityMismatchError) Error() string {
	return fmt.Sprintf(
		"step %q requires %s, but target only provides %s (missing: %s)",
		e.StepKind, e.RequiredCaps, e.ProvidedCaps, e.MissingCaps,
	)
}

func (e CapabilityMismatchError) EventTemplate() event.Template {
	return event.Template{
		ID:     "engine.CapabilityMismatch",
		Text:   `step "{{.StepKind}}" requires capabilities not provided by target`,
		Hint:   "use a different target or remove incompatible steps",
		Help:   "missing:  {{.MissingCaps}}\nrequired: {{.RequiredCaps}}\nprovided: {{.ProvidedCaps}}",
		Data:   e,
		Source: &e.Source,
	}
}

func (e CapabilityMismatchError) Severity() signal.Severity {
	return signal.Error
}

func (e CapabilityMismatchError) Impact() diagnostic.Impact {
	return diagnostic.ImpactAbort
}

func panicIfNotAbortError(err error) error {
	var abort AbortError
	if errors.As(err, &abort) {
		return abort
	}
	// very cold codepath
	wrap := errs.BUG("Engine failed with non-signal error: %w", err)
	if pc, file, line, ok := runtime.Caller(1); ok {
		_ = file
		_ = line
		details := runtime.FuncForPC(pc)
		wrap = errs.BUG("%s failed with non-signal error: %w", details.Name(), err)
	}
	panic(wrap)
}

// emitScopedDiagnostic extracts diagnostic(s) from err and passes each to emit.
// Returns the max impact and whether any diagnostic was emitted.
func emitScopedDiagnostic(err error, emit func(diagnostic.Diagnostic)) (diagnostic.Impact, bool) {
	if err == nil {
		return 0, false
	}

	var ds diagnostic.Diagnostics
	if errors.As(err, &ds) {
		impact := diagnostic.ImpactNone
		for _, d := range ds {
			emit(d)
			if d.Impact() > impact {
				impact = d.Impact()
			}
		}
		return impact, true
	}

	var d diagnostic.Diagnostic
	if !errors.As(err, &d) {
		return 0, false
	}

	emit(d)
	return d.Impact(), true
}

func emitEngineDiagnostic(em diagnostic.Emitter, cfgPath string, err error) (diagnostic.Impact, bool) {
	return emitScopedDiagnostic(err, func(d diagnostic.Diagnostic) {
		em.EmitEngineDiagnostic(diagnostic.RaiseEngineDiagnostic(cfgPath, d))
	})
}

func emitPlanDiagnostic(
	em diagnostic.Emitter, stepIndex int, stepKind, stepDesc string, err error,
) (diagnostic.Impact, bool) {
	return emitScopedDiagnostic(err, func(d diagnostic.Diagnostic) {
		em.EmitPlanDiagnostic(diagnostic.RaisePlanDiagnostic(stepIndex, stepKind, stepDesc, d))
	})
}

func emitActionDiagnostic(
	em diagnostic.Emitter, stepIndex int, stepKind, stepDesc string, err error,
) (diagnostic.Impact, bool) {
	return emitScopedDiagnostic(err, func(d diagnostic.Diagnostic) {
		em.EmitActionDiagnostic(diagnostic.RaiseActionDiagnostic(stepIndex, stepKind, stepDesc, d))
	})
}

func emitOpDiagnostic(
	em diagnostic.Emitter, stepIndex int, stepKind, stepDesc, displayID string, err error,
) (diagnostic.Impact, bool) {
	return emitScopedDiagnostic(err, func(d diagnostic.Diagnostic) {
		em.EmitOpDiagnostic(diagnostic.RaiseOpDiagnostic(stepIndex, stepKind, stepDesc, displayID, d))
	})
}
