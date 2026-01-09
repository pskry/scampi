package diagnostic

import (
	"time"

	"godoit.dev/doit/diagnostic/event"
	"godoit.dev/doit/render"
	"godoit.dev/doit/signal"
)

type (
	Policy struct {
		WarningsAsErrors bool
		Verbosity        signal.Verbosity
	}
	Decision struct {
		Severity signal.Severity
		Show     bool
	}
	policyEmitter struct {
		pol Policy
		out render.Displayer
	}
)

func NewEmitter(policy Policy, displayer render.Displayer) Emitter {
	return &policyEmitter{
		pol: policy,
		out: displayer,
	}
}

func (p Policy) apply(s signal.Severity) signal.Severity {
	if s == signal.Warning && p.WarningsAsErrors {
		return signal.Error
	}

	return s
}

func (p *policyEmitter) Emit(ev event.Event) {
	ev.Severity = p.pol.apply(ev.Severity)
	p.out.Emit(ev)
}

// Engine lifecycle
// ===============================================

func (p *policyEmitter) EngineStart() {
	if p.pol.Verbosity >= signal.VV {
		p.out.EngineStart(p.pol.apply(signal.Info))
	}
}

func (p *policyEmitter) EngineFinish(rs RunSummary, duration time.Duration) {
	drs := render.RunSummary{
		ChangedCount: rs.ChangedCount,
		FailedCount:  rs.FailedCount,
		TotalCount:   rs.TotalCount,
	}
	p.out.EngineFinish(p.pol.apply(signal.Important), drs, duration)
}

// Planning lifecycle
// ===============================================

func (p *policyEmitter) PlanStart() {
	if p.pol.Verbosity >= signal.VV {
		p.out.PlanStart(p.pol.apply(signal.Info))
	}
}

func (p *policyEmitter) UnitPlanned(index int, name, kind string) {
	if p.pol.Verbosity >= signal.VV {
		p.out.UnitPlanned(p.pol.apply(signal.Debug), index, name, kind)
	}
}

func (p *policyEmitter) PlanFinish(unitCount int, duration time.Duration) {
	if p.pol.Verbosity >= signal.VV {
		p.out.PlanFinish(p.pol.apply(signal.Info), unitCount, duration)
	}
}

func (p *policyEmitter) PlanError(index int, name, kind string, diag Diagnostic) {
	p.out.PlanError(p.pol.apply(signal.Error), index, name, kind, toRenderTempl(diag))
}

// Action lifecycle
// ===============================================

func (p *policyEmitter) ActionStart(name string) {
	if p.pol.Verbosity >= signal.V {
		p.out.ActionStart(p.pol.apply(signal.Notice), name)
	}
}

func (p *policyEmitter) ActionFinish(name string, changed bool, duration time.Duration) {
	if changed {
		p.out.ActionFinish(p.pol.apply(signal.Important), name, changed, duration)
		return
	}

	if p.pol.Verbosity >= signal.V {
		p.out.ActionFinish(p.pol.apply(signal.Info), name, changed, duration)
	}
}

func (p *policyEmitter) ActionError(name string, err error) {
	p.out.ActionError(p.pol.apply(signal.Error), name, err)
}

// OpCheck lifecycle
// ===============================================

func (p *policyEmitter) OpCheckStart(action, op string) {
	if p.pol.Verbosity >= signal.VV {
		p.out.OpCheckStart(p.pol.apply(signal.Debug), action, op)
	}
}

func (p *policyEmitter) OpCheckSatisfied(action, op string) {
	if p.pol.Verbosity >= signal.VV {
		p.out.OpCheckSatisfied(p.pol.apply(signal.Debug), action, op)
	}
}

func (p *policyEmitter) OpCheckUnsatisfied(action, op string) {
	if p.pol.Verbosity >= signal.V {
		p.out.OpCheckUnsatisfied(p.pol.apply(signal.Notice), action, op)
	}
}

func (p *policyEmitter) OpCheckUnknown(action, op string, err error) {
	p.out.OpCheckUnknown(p.pol.apply(signal.Warning), action, op, err)
}

// OpExecute lifecycle
// ===============================================

func (p *policyEmitter) OpExecuteStart(action, op string) {
	if p.pol.Verbosity >= signal.VV {
		p.out.OpExecuteStart(p.pol.apply(signal.Debug), action, op)
	}
}

func (p *policyEmitter) OpExecuteFinish(action, op string, changed bool, duration time.Duration) {
	if changed {
		if p.pol.Verbosity >= signal.VV {
			p.out.OpExecuteFinish(p.pol.apply(signal.Info), action, op, changed, duration)
		}
		return
	}

	if p.pol.Verbosity >= signal.VV {
		p.out.OpExecuteFinish(p.pol.apply(signal.Debug), action, op, changed, duration)
	}
}

func (p *policyEmitter) OpExecuteError(action, op string, err error) {
	p.out.OpExecuteError(p.pol.apply(signal.Error), action, op, err)
}

// Errors
// ===============================================

func (p *policyEmitter) UserError(diag Diagnostic) {
	p.out.UserError(
		p.pol.apply(signal.Error),
		toRenderTempl(diag),
	)
}

func (p *policyEmitter) InternalError(message string, err error) {
	p.out.InternalError(
		p.pol.apply(signal.Error),
		render.Template{
			Text: "legacy.message",
		},
	)
}

func toRenderTempl(diag Diagnostic) render.Template {
	t := diag.Template()
	return render.Template{
		Name: t.Name,
		Text: t.Text,
		Hint: t.Hint,
		Help: t.Help,
		Data: diag,
	}
}
