// SPDX-License-Identifier: GPL-3.0-only

//go:generate stringer -type=EngineKind
//go:generate stringer -type=PlanKind
//go:generate stringer -type=ActionKind
//go:generate stringer -type=OpKind
//go:generate stringer -type=Chattiness
package event

import (
	"time"

	"godoit.dev/doit/signal"
	"godoit.dev/doit/spec"
)

type StepDetail struct {
	StepIndex int
	StepKind  string
	StepDesc  string
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
	Step           StepDetail
	StartedDetail  *PlanStartedDetail
	Detail         *PlanDetail
	FinishedDetail *PlanFinishedDetail
	Severity       signal.Severity
	Chattiness     Chattiness
}

type ActionEvent struct {
	Time       time.Time
	Kind       ActionKind
	Step       StepDetail
	Detail     *ActionDetail
	Severity   signal.Severity
	Chattiness Chattiness
}

type OpEvent struct {
	Time          time.Time
	Kind          OpKind
	Step          StepDetail
	DisplayID     string
	CheckDetail   *OpCheckDetail
	ExecuteDetail *OpExecuteDetail
	Severity      signal.Severity
	Chattiness    Chattiness
}

type IndexAllEvent struct {
	Time       time.Time
	Steps      []StepIndexDetail
	Severity   signal.Severity
	Chattiness Chattiness
}

type IndexStepEvent struct {
	Time       time.Time
	Doc        spec.StepDoc
	Severity   signal.Severity
	Chattiness Chattiness
}

type EngineDiagnostic struct {
	Time       time.Time
	CfgPath    string
	Detail     DiagnosticDetail
	Severity   signal.Severity
	Chattiness Chattiness
}

type PlanDiagnostic struct {
	Time       time.Time
	Step       StepDetail
	Detail     DiagnosticDetail
	Severity   signal.Severity
	Chattiness Chattiness
}

type ActionDiagnostic struct {
	Time       time.Time
	Step       StepDetail
	Detail     DiagnosticDetail
	Severity   signal.Severity
	Chattiness Chattiness
}

type OpDiagnostic struct {
	Time       time.Time
	Step       StepDetail
	DisplayID  string
	Detail     DiagnosticDetail
	Severity   signal.Severity
	Chattiness Chattiness
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

// Field is a single renderable text within a Template.
type Field struct {
	id, text string
	data     any
}

func (f Field) TemplateID() string   { return f.id }
func (f Field) TemplateText() string { return f.text }
func (f Field) TemplateData() any    { return f.data }

func (t Template) TextField() Field { return Field{t.ID + ".Text", t.Text, t.Data} }
func (t Template) HintField() Field { return Field{t.ID + ".Hint", t.Hint, t.Data} }
func (t Template) HelpField() Field { return Field{t.ID + ".Help", t.Help, t.Data} }
