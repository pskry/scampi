package event

import (
	"time"

	"godoit.dev/doit/spec"
)

type EngineDetail struct {
	ChangedCount int
	FailedCount  int
	TotalCount   int
	Duration     time.Duration
	Err          error
}

type PlanDetail struct {
	UnitCount int
	Duration  time.Duration
	Problems  []PlanProblem
}
type PlanProblem struct {
	Index int
	Name  string
	Kind  string
	Err   error
}

type ActionDetail struct {
	Changed  bool
	Duration time.Duration
	Err      error
}

type OpCheckDetail struct {
	Result spec.CheckResult
	Err    error
}

type OpExecuteDetail struct {
	Changed  bool
	Duration time.Duration
	Err      error
}

type DiagnosticDetail struct {
	Template Template
}
