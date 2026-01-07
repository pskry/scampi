package diagnostic

import (
	"time"
)

type RunSummary struct {
	ChangedCount int
	FailedCount  int
	TotalCount   int
}

type Emitter interface {
	// Engine lifecycle
	// ===============================================

	EngineStart()
	EngineFinish(rs RunSummary, duration time.Duration)

	// Planning lifecycle
	// ===============================================

	PlanStart()
	UnitPlanned(index int, name string, kind string)
	PlanFinish(unitCount int, duration time.Duration)

	// Action lifecycle
	// ===============================================

	ActionStart(name string)
	ActionFinish(name string, changed bool, duration time.Duration)
	ActionError(name string, err error)

	// OpCheck lifecycle
	// ===============================================

	OpCheckStart(action string, op string)
	OpCheckSatisfied(action string, op string)
	OpCheckUnsatisfied(action string, op string)
	OpCheckUnknown(action string, op string, err error)

	// OpExecute lifecycle
	// ===============================================

	OpExecuteStart(action string, op string)
	OpExecuteFinish(action string, op string, changed bool, duration time.Duration)
	OpExecuteError(action string, op string, err error)

	// Errors
	// ===============================================

	UserError(message string)
	InternalError(message string, err error)
}
