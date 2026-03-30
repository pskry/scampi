// SPDX-License-Identifier: GPL-3.0-only

package testkit

import (
	"context"
	"fmt"

	"go.starlark.net/starlark"

	"scampi.dev/scampi/errs"
	"scampi.dev/scampi/target"
)

var serviceAssertionAttrs = []string{
	"is_disabled",
	"is_enabled",
	"is_running",
	"is_stopped",
}

// ServiceAssertion is the Starlark value returned by assert_that.service(name).
type ServiceAssertion struct {
	tgt       target.Target
	name      string
	collector *Collector
}

func (a *ServiceAssertion) String() string        { return fmt.Sprintf("service_assertion(%s)", a.name) }
func (a *ServiceAssertion) Type() string          { return "service_assertion" }
func (a *ServiceAssertion) Freeze()               {}
func (a *ServiceAssertion) Truth() starlark.Bool  { return starlark.True }
func (a *ServiceAssertion) Hash() (uint32, error) { return 0, nil }
func (a *ServiceAssertion) AttrNames() []string   { return serviceAssertionAttrs }

func (a *ServiceAssertion) Attr(name string) (starlark.Value, error) {
	switch name {
	case "is_running":
		return starlark.NewBuiltin("service.is_running", a.builtinIsRunning), nil
	case "is_stopped":
		return starlark.NewBuiltin("service.is_stopped", a.builtinIsStopped), nil
	case "is_enabled":
		return starlark.NewBuiltin("service.is_enabled", a.builtinIsEnabled), nil
	case "is_disabled":
		return starlark.NewBuiltin("service.is_disabled", a.builtinIsDisabled), nil
	}
	return nil, nil
}

func (a *ServiceAssertion) builtinIsRunning(
	_ *starlark.Thread,
	_ *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs("is_running", args, kwargs, 0); err != nil {
		return nil, err
	}
	name := a.name
	tgt := a.tgt
	a.collector.Add(Assertion{
		Description: fmt.Sprintf("service %s is running", name),
		Check: func() error {
			sm := target.Must[target.ServiceManager]("service.is_running", tgt)
			active, err := sm.IsActive(context.Background(), name)
			if err != nil {
				// bare-error: assertion result
				return errs.Errorf("service %s: %s", name, err)
			}
			if !active {
				// bare-error: assertion result
				return errs.Errorf("service %s is not running", name)
			}
			return nil
		},
	})
	return starlark.None, nil
}

func (a *ServiceAssertion) builtinIsStopped(
	_ *starlark.Thread,
	_ *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs("is_stopped", args, kwargs, 0); err != nil {
		return nil, err
	}
	name := a.name
	tgt := a.tgt
	a.collector.Add(Assertion{
		Description: fmt.Sprintf("service %s is stopped", name),
		Check: func() error {
			sm := target.Must[target.ServiceManager]("service.is_stopped", tgt)
			active, err := sm.IsActive(context.Background(), name)
			if err != nil {
				// bare-error: assertion result
				return errs.Errorf("service %s: %s", name, err)
			}
			if active {
				// bare-error: assertion result
				return errs.Errorf("service %s is running but should be stopped", name)
			}
			return nil
		},
	})
	return starlark.None, nil
}

func (a *ServiceAssertion) builtinIsEnabled(
	_ *starlark.Thread,
	_ *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs("is_enabled", args, kwargs, 0); err != nil {
		return nil, err
	}
	name := a.name
	tgt := a.tgt
	a.collector.Add(Assertion{
		Description: fmt.Sprintf("service %s is enabled", name),
		Check: func() error {
			sm := target.Must[target.ServiceManager]("service.is_enabled", tgt)
			enabled, err := sm.IsEnabled(context.Background(), name)
			if err != nil {
				// bare-error: assertion result
				return errs.Errorf("service %s: %s", name, err)
			}
			if !enabled {
				// bare-error: assertion result
				return errs.Errorf("service %s is not enabled", name)
			}
			return nil
		},
	})
	return starlark.None, nil
}

func (a *ServiceAssertion) builtinIsDisabled(
	_ *starlark.Thread,
	_ *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs("is_disabled", args, kwargs, 0); err != nil {
		return nil, err
	}
	name := a.name
	tgt := a.tgt
	a.collector.Add(Assertion{
		Description: fmt.Sprintf("service %s is disabled", name),
		Check: func() error {
			sm := target.Must[target.ServiceManager]("service.is_disabled", tgt)
			enabled, err := sm.IsEnabled(context.Background(), name)
			if err != nil {
				// bare-error: assertion result
				return errs.Errorf("service %s: %s", name, err)
			}
			if enabled {
				// bare-error: assertion result
				return errs.Errorf("service %s is enabled but should be disabled", name)
			}
			return nil
		},
	})
	return starlark.None, nil
}
