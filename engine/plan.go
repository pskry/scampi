package engine

import (
	"time"

	"godoit.dev/doit/diagnostic"
	"godoit.dev/doit/diagnostic/event"
	"godoit.dev/doit/spec"
)

func Plan(cfg spec.Config, em diagnostic.Emitter) (spec.Plan, error) {
	start := time.Now()
	em.Emit(diagnostic.PlanStarted())

	var (
		plan        spec.Plan
		causes      []error
		diagResults []diagnosticResult
	)

	for i, unit := range cfg.Units {
		act, err := unit.Type.Plan(i, unit)
		if err != nil {
			dr := emitDiagnostics(
				em,
				event.Subject{
					Index: i,
					Name:  unit.Name,
					Kind:  unit.Type.Kind(),
				},
				err,
			)

			diagResults = append(diagResults, dr)
			causes = append(causes, err)
			continue
		}

		plan.Actions = append(plan.Actions, act)
		em.Emit(diagnostic.UnitPlanned(i, act.Name(), unit.Type.Kind()))
	}

	em.Emit(diagnostic.PlanFinished(
		len(plan.Actions),
		len(causes),
		time.Since(start),
	))

	for _, dr := range diagResults {
		if dr.ShouldAbort() {
			return spec.Plan{}, AbortError{
				Causes: causes,
			}
		}
	}

	return plan, nil
}
