package spec

import "context"

type (
	Config struct {
		Tasks []CfgTask
	}
	CfgTask struct {
		Kind   string
		Spec   Spec
		Config any
	}

	Spec interface {
		Kind() string
		Schema() string
		NewConfig() any
		Plan(idx int, cfg any) (RtTask, error)
	}
	RtPlan struct {
		Tasks []RtTask
	}
	RtTask interface {
		Name() string
		Ops() []Op
	}
	Op interface {
		Name() string
		Execute(ctx context.Context) (Result, error)
	}
	Result struct {
		Changed bool
	}
)
