package spec

import (
	"context"

	"godoit.dev/doit/source"
	"godoit.dev/doit/target"
)

type (
	Config struct {
		Units   []UnitInstance
		Sources SourceStore
	}
	UnitInstance struct {
		Name   string
		Type   UnitType
		Config any
		Source SourceSpan
		Fields map[string]FieldSpan
	}
	UnitType interface {
		Kind() string
		NewConfig() any
		Plan(idx int, unit UnitInstance) (Action, error)
	}
	FieldSpan struct {
		Field SourceSpan
		Value SourceSpan
	}
	SourceSpan struct {
		Filename string
		Line     int
		StartCol int
		EndCol   int
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
		Action() string
		Check(ctx context.Context, src source.Source, tgt target.Target) (CheckResult, error)
		Execute(ctx context.Context, src source.Source, tgt target.Target) (Result, error)
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
