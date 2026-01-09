package diagnostic

import (
	"time"

	"godoit.dev/doit/diagnostic/event"
	"godoit.dev/doit/signal"
	"godoit.dev/doit/spec"
)

// Engine lifecycle
// ===============================================

func EngineStarted() event.Event {
	return event.Event{
		Time:       time.Now(),
		Kind:       event.EngineStarted,
		Scope:      event.ScopeEngine,
		Severity:   signal.Info,
		Chattiness: event.Subtle,
	}
}

func EngineFinished(rs RunSummary, dur time.Duration, err error) event.Event {
	e := event.Event{
		Time:  time.Now(),
		Kind:  event.EngineFinished,
		Scope: event.ScopeEngine,
		Detail: event.EngineDetail{
			ChangedCount: rs.ChangedCount,
			FailedCount:  rs.FailedCount,
			TotalCount:   rs.TotalCount,
			Duration:     dur,
			Err:          err,
		},
	}

	switch {
	case err != nil:
		e.Severity = signal.Fatal
		e.Chattiness = event.Normal

	case rs.FailedCount > 0:
		e.Severity = signal.Error
		e.Chattiness = event.Normal

	case rs.ChangedCount > 0:
		e.Severity = signal.Important
		e.Chattiness = event.Subtle

	default:
		e.Severity = signal.Info
		e.Chattiness = event.Subtle
	}

	return e
}

// Plan lifecycle
// ===============================================

func PlanStarted() event.Event {
	return event.Event{
		Time:       time.Now(),
		Kind:       event.PlanStarted,
		Scope:      event.ScopePlan,
		Severity:   signal.Info,
		Chattiness: event.Subtle,
	}
}

func PlanFinished(unitCount int, dur time.Duration, problems []event.PlanProblem) event.Event {
	e := event.Event{
		Time:  time.Now(),
		Kind:  event.PlanFinished,
		Scope: event.ScopePlan,
		Detail: event.PlanDetail{
			UnitCount: unitCount,
			Duration:  dur,
			Problems:  problems,
		},
	}

	switch {
	case len(problems) > 0:
		e.Severity = signal.Error
		e.Chattiness = event.Normal

	case unitCount == 0:
		e.Severity = signal.Warning
		e.Chattiness = event.Normal

	default:
		e.Severity = signal.Info
		e.Chattiness = event.Subtle
	}

	return e
}

func UnitPlanned(
	index int,
	name string,
	kind string,
) event.Event {
	return event.Event{
		Time:  time.Now(),
		Kind:  event.UnitPlanned,
		Scope: event.ScopePlan,
		Subject: event.Subject{
			Index: index,
			Name:  name,
			Kind:  kind,
		},
		Severity:   signal.Debug,
		Chattiness: event.Chatty,
	}
}

// Action lifecycle
// ===============================================

func ActionStarted(name string) event.Event {
	return event.Event{
		Time:  time.Now(),
		Kind:  event.ActionStarted,
		Scope: event.ScopeAction,
		Subject: event.Subject{
			Action: name,
		},
		Severity:   signal.Notice,
		Chattiness: event.Normal,
	}
}

func ActionFinished(name string, changed bool, dur time.Duration, err error) event.Event {
	e := event.Event{
		Time:  time.Now(),
		Kind:  event.ActionFinished,
		Scope: event.ScopeAction,
		Subject: event.Subject{
			Action: name,
		},
		Detail: event.ActionDetail{
			Changed:  changed,
			Duration: dur,
			Err:      err,
		},
	}

	switch {
	case err != nil:
		e.Severity = signal.Error
		e.Chattiness = event.Normal

	case changed:
		e.Severity = signal.Important
		e.Chattiness = event.Normal

	default:
		e.Severity = signal.Info
		e.Chattiness = event.Reserved
	}

	return e
}

// Op lifecycle
// ===============================================

func OpCheckStarted(action, op string) event.Event {
	return event.Event{
		Time:  time.Now(),
		Kind:  event.OpCheckStarted,
		Scope: event.ScopeOp,
		Subject: event.Subject{
			Action: action,
			Op:     op,
		},
		Severity:   signal.Debug,
		Chattiness: event.Chatty,
	}
}

func OpChecked(action, op string, res spec.CheckResult, err error) event.Event {
	e := event.Event{
		Time:  time.Now(),
		Kind:  event.OpChecked,
		Scope: event.ScopeOp,
		Subject: event.Subject{
			Action: action,
			Op:     op,
		},
		Detail: event.OpCheckDetail{
			Result: res,
			Err:    err,
		},
	}

	switch res {
	case spec.CheckSatisfied:
		e.Severity = signal.Info
		e.Chattiness = event.Subtle

	case spec.CheckUnsatisfied:
		e.Severity = signal.Notice
		e.Chattiness = event.Normal

	case spec.CheckUnknown:
		e.Severity = signal.Warning
		e.Chattiness = event.Reserved
	}

	return e
}

func OpExecuteStarted(action, op string) event.Event {
	return event.Event{
		Time:  time.Now(),
		Kind:  event.OpExecuteStarted,
		Scope: event.ScopeOp,
		Subject: event.Subject{
			Action: action,
			Op:     op,
		},
		Severity:   signal.Info,
		Chattiness: event.Chatty,
	}
}

func OpExecuted(action, op string, changed bool, dur time.Duration, err error) event.Event {
	e := event.Event{
		Time:  time.Now(),
		Kind:  event.OpExecuted,
		Scope: event.ScopeOp,
		Subject: event.Subject{
			Action: action,
			Op:     op,
		},
		Detail: event.OpExecuteDetail{
			Changed:  changed,
			Duration: dur,
			Err:      err,
		},
	}

	switch {
	case err != nil:
		e.Severity = signal.Error
		e.Chattiness = event.Normal

	case changed:
		e.Severity = signal.Important
		e.Chattiness = event.Normal

	default:
		e.Severity = signal.Info
		e.Chattiness = event.Reserved
	}

	return e
}

type (
	RunSummary struct {
		ChangedCount int
		FailedCount  int
		TotalCount   int
	}
	Template struct {
		Name string
		Text string
		Hint string
		Help string
	}
)

type Diagnostic interface {
	Template() Template
	Severity() signal.Severity
}

type Emitter interface {
	Emit(e event.Event)

	// Engine lifecycle
	// ===============================================

	EngineStart()
	EngineFinish(rs RunSummary, dur time.Duration)

	// Planning lifecycle
	// ===============================================

	PlanStart()
	UnitPlanned(index int, name string, kind string)
	PlanFinish(unitCount int, dur time.Duration)
	PlanError(index int, name string, kind string, diag Diagnostic)

	// Action lifecycle
	// ===============================================

	ActionStart(name string)
	ActionFinish(name string, changed bool, dur time.Duration)
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
	OpExecuteFinish(action string, op string, changed bool, dur time.Duration)
	OpExecuteError(action string, op string, err error)

	// Errors
	// ===============================================

	UserError(diag Diagnostic)
	InternalError(message string, err error)
}
