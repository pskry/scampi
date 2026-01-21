//go:generate stringer -type=EngineKind
//go:generate stringer -type=PlanKind
//go:generate stringer -type=ActionKind
//go:generate stringer -type=OpKind
//go:generate stringer -type=Kind
//go:generate stringer -type=Scope
//go:generate stringer -type=Chattiness
package event

import (
	"time"

	"godoit.dev/doit/signal"
	"godoit.dev/doit/spec"
)

// Event represents a single, immutable fact that occurred during execution.
//
// An Event describes *what happened*, not how it should be rendered.
// It is the primary integration point between the engine, diagnostics,
// policy, and renderers.
//
// Invariant:
//   - Every Event MUST be emitted with Severity and Chattiness fully populated.
//   - Renderers MUST NOT infer or guess defaults for either field.
//   - Policy MAY adjust Severity, but MUST NOT alter Chattiness.
//   - Chattiness MUST NOT be used to indicate importance or failure.
//
// These rules ensure that events are semantically complete at creation time
// and can be safely consumed by multiple renderers (CLI, JSON, UI) without
// hidden coupling or duplicated logic.
type Event struct {
	Time       time.Time
	Kind       Kind
	Scope      Scope
	Subject    Subject
	Detail     any
	Severity   signal.Severity
	Chattiness Chattiness
}
type EngineEvent struct {
	Time       time.Time
	Kind       EngineKind
	Detail     *EngineFinishedDetail
	Severity   signal.Severity
	Chattiness Chattiness
}
type PlanEvent struct {
	Time           time.Time
	Kind           PlanKind
	Subject        Subject
	StartedDetail  *PlanStartedDetail
	Detail         *PlanDetail
	FinishedDetail *PlanFinishedDetail
	Severity       signal.Severity
	Chattiness     Chattiness
}
type ActionEvent struct {
	Time       time.Time
	Kind       ActionKind
	Subject    Subject
	Detail     *ActionDetail
	Severity   signal.Severity
	Chattiness Chattiness
}
type OpEvent struct {
	Time          time.Time
	Kind          OpKind
	Subject       Subject
	CheckDetail   *OpCheckDetail
	ExecuteDetail *OpExecuteDetail
	Severity      signal.Severity
	Chattiness    Chattiness
}

type EngineKind uint8

const (
	EngineStarted EngineKind = iota
	EngineFinished
)

type PlanKind uint8

const (
	PlanStarted PlanKind = iota
	PlanFinished
	StepPlanned
	PlanProduced
)

type ActionKind uint8

const (
	ActionStarted ActionKind = iota
	ActionFinished
)

type OpKind uint8

const (
	OpCheckStarted OpKind = iota
	OpChecked

	OpExecuteStarted
	OpExecuted
)

type Kind uint8

const (
	DiagnosticRaised Kind = iota
)

type Scope uint8

const (
	ScopeEngine Scope = iota
	ScopePlan
	ScopeAction
	ScopeOp
)

// Subject is a sealed interface representing the origin context of an event.
// Use type switch to handle concrete types: EngineSubject, PlanSubject,
// ActionSubject, OpSubject.
type Subject interface {
	isSubject() // sealed marker - only types in this package can implement
	Scope() Scope
}

// EngineSubject represents engine-level context (config loading, etc.)
type EngineSubject struct {
	CfgPath string
}

// TODO: Introduce some 'BaseStepSubject' or something

// PlanSubject represents step context during planning phase.
type PlanSubject struct {
	StepIndex int
	StepKind  string
	StepDesc  string
}

// ActionSubject represents step context during execution phase.
type ActionSubject struct {
	StepIndex int
	StepKind  string
	StepDesc  string
}

// OpSubject represents op context during execution phase.
type OpSubject struct {
	StepIndex int
	StepKind  string
	StepDesc  string
	DisplayID string // derived from OpDescriber or fallback
}

func (EngineSubject) isSubject() {}
func (PlanSubject) isSubject()   {}
func (ActionSubject) isSubject() {}
func (OpSubject) isSubject()     {}

func (EngineSubject) Scope() Scope { return ScopeEngine }
func (PlanSubject) Scope() Scope   { return ScopePlan }
func (ActionSubject) Scope() Scope { return ScopeAction }
func (OpSubject) Scope() Scope     { return ScopeOp }

// Chattiness describes how noisy an event is under normal operation.
// It is orthogonal to Severity and MUST NOT be used to indicate importance.
type Chattiness uint8

const (
	Subtle Chattiness = iota
	Reserved
	Normal
	Chatty
	Yappy
)

type Template struct {
	ID   string
	Text string
	Hint string
	Help string
	Data any

	Source *spec.SourceSpan
}
