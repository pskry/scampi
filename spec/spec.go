package spec

import (
	"context"

	"godoit.dev/doit/target"
)

type (
	Config struct {
		Units []CfgUnit
	}
	CfgUnit struct {
		Kind   string
		Name   string
		Impl   KindImpl
		Config any
	}

	KindImpl interface {
		Kind() string
		NewConfig() any
		Plan(idx int, cfg any) (Action, error)
	}
	Plan struct {
		Actions []Action
	}
	Action interface {
		Name() string
		Ops() []Op
	}
	Op interface {
		Name() string
		Check(ctx context.Context, tgt target.Target) (CheckResult, error)
		Execute(ctx context.Context, tgt target.Target) (Result, error)
		DependsOn() []Op
	}
	Result struct {
		Changed bool
	}
	CheckResult uint8
)

const (
	CheckUnknown CheckResult = iota
	CheckSatisfied
	CheckUnsatisfied
)
