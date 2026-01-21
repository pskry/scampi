package diagnostic

import (
	"godoit.dev/doit/diagnostic/event"
	"godoit.dev/doit/render"
	"godoit.dev/doit/signal"
)

type (
	Policy struct {
		WarningsAsErrors bool
		Verbosity        signal.Verbosity
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

func (p *policyEmitter) EmitEngineLifecycle(ev event.EngineEvent) {
	ev.Severity = p.pol.apply(ev.Severity)
	p.out.EmitEngineLifecycle(ev)
}

func (p *policyEmitter) EmitPlanLifecycle(ev event.PlanEvent) {
	ev.Severity = p.pol.apply(ev.Severity)
	p.out.EmitPlanLifecycle(ev)
}

func (p *policyEmitter) EmitActionLifecycle(ev event.ActionEvent) {
	ev.Severity = p.pol.apply(ev.Severity)
	p.out.EmitActionLifecycle(ev)
}

func (p *policyEmitter) EmitOpLifecycle(ev event.OpEvent) {
	ev.Severity = p.pol.apply(ev.Severity)
	p.out.EmitOpLifecycle(ev)
}

func (p *policyEmitter) EmitDiagnostic(ev event.Event) {
	ev.Severity = p.pol.apply(ev.Severity)
	p.out.EmitDiagnostic(ev)
}
