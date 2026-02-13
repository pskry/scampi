// SPDX-License-Identifier: GPL-3.0-only

package star

import (
	"fmt"
	"strconv"

	"go.starlark.net/starlark"

	"godoit.dev/doit/spec"
)

// predeclared returns the global builtins available in every .star file.
func predeclared() starlark.StringDict {
	return starlark.StringDict{
		"copy":     starlark.NewBuiltin("copy", builtinCopy),
		"dir":      starlark.NewBuiltin("dir", builtinDir),
		"pkg":      starlark.NewBuiltin("pkg", builtinPkg),
		"symlink":  starlark.NewBuiltin("symlink", builtinSymlink),
		"template": starlark.NewBuiltin("template", builtinTemplate),
		"target":   targetModule(),
		"deploy":   starlark.NewBuiltin("deploy", builtinDeploy),
		"env":      starlark.NewBuiltin("env", builtinEnv),
	}
}

// deploy(name, targets, steps)
// -----------------------------------------------------------------------------

func builtinDeploy(
	thread *starlark.Thread,
	_ *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	var (
		name    string
		targets *starlark.List
		steps   *starlark.List
	)
	if err := starlark.UnpackArgs("deploy", args, kwargs,
		"name", &name,
		"targets", &targets,
		"steps", &steps,
	); err != nil {
		return nil, err
	}

	targetNames, err := stringList(targets, "deploy", "targets")
	if err != nil {
		return nil, err
	}

	stepInstances, err := extractSteps(steps, "deploy")
	if err != nil {
		return nil, err
	}

	span := callSpan(thread)
	block := spec.DeployBlock{
		Name:    name,
		Targets: targetNames,
		Steps:   stepInstances,
		Source:  span,
	}

	c := threadCollector(thread)
	if err := c.AddDeploy(name, block, span); err != nil {
		return nil, err
	}

	return starlark.None, nil
}

func extractSteps(
	list *starlark.List, fn string,
) ([]spec.StepInstance, error) {
	if list == nil {
		return nil, nil
	}
	out := make([]spec.StepInstance, 0, list.Len())
	for i := 0; i < list.Len(); i++ {
		v := list.Index(i)
		step, ok := v.(*StarlarkStep)
		if !ok {
			return nil, fmt.Errorf(
				"%s: steps[%d] must be a step value, got %s",
				fn, i, v.Type(),
			)
		}
		out = append(out, step.Instance)
	}
	return out, nil
}

// env(key, default?)
// -----------------------------------------------------------------------------

func builtinEnv(
	thread *starlark.Thread,
	_ *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf(
			"env: accepts 1 or 2 positional arguments, got %d", len(args),
		)
	}
	if len(kwargs) > 0 {
		return nil, fmt.Errorf("env: does not accept keyword arguments")
	}

	key, ok := starlark.AsString(args[0])
	if !ok {
		return nil, fmt.Errorf(
			"env: key must be a string, got %s", args[0].Type(),
		)
	}

	c := threadCollector(thread)
	envVal, found := c.src.LookupEnv(key)

	// No default → required
	if len(args) == 1 {
		if !found {
			return nil, EnvVarRequiredError{
				Key:    key,
				Source: callSpan(thread),
			}
		}
		return starlark.String(envVal), nil
	}

	// Has default → coerce env value to match default's type
	dflt := args[1]
	if !found {
		return dflt, nil
	}

	return coerceEnvValue(envVal, dflt)
}

func coerceEnvValue(
	raw string, dflt starlark.Value,
) (starlark.Value, error) {
	switch dflt.(type) {
	case starlark.String:
		return starlark.String(raw), nil

	case starlark.Int:
		i, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("env: cannot parse %q as int: %w", raw, err)
		}
		return starlark.MakeInt64(i), nil

	case starlark.Bool:
		switch raw {
		case "true", "1", "yes":
			return starlark.True, nil
		case "false", "0", "no", "":
			return starlark.False, nil
		default:
			return nil, fmt.Errorf("env: cannot parse %q as bool", raw)
		}

	default:
		return starlark.String(raw), nil
	}
}
