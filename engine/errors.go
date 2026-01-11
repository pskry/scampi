package engine

import (
	"errors"
	"slices"

	"godoit.dev/doit/diagnostic"
	"godoit.dev/doit/diagnostic/event"
)

type AbortError struct {
	Causes []error
}

func (AbortError) Error() string {
	return "execution aborted"
}

type (
	diagnosticResult struct {
		Effects []diagnostic.Effect
	}
)

func (r *diagnosticResult) add(effect diagnostic.Effect) {
	r.Effects = append(r.Effects, effect)
}

func (r diagnosticResult) ShouldAbort() bool {
	return slices.Contains(r.Effects, diagnostic.EffectAbort)
}

func emitDiagnostics(
	em diagnostic.Emitter,
	subject event.Subject,
	err error,
) diagnosticResult {
	var res diagnosticResult

	if err == nil {
		return res
	}

	var dp diagnostic.DiagnosticProvider
	if !errors.As(err, &dp) {
		return res
	}

	for _, ev := range dp.Diagnostics(subject) {
		em.Emit(ev)

		// determine effect
		effect := diagnostic.EffectAbort // default (safe)
		if ep, ok := err.(diagnostic.EffectProvider); ok {
			effect = ep.Effect()
		}

		res.add(effect)
	}

	return res
}
