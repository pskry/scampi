// SPDX-License-Identifier: GPL-3.0-only

package star

import (
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"

	"scampi.dev/scampi/spec"
	"scampi.dev/scampi/step/container"
)

// containerModule builds the `container` namespace (container.instance).
func containerModule() *starlarkstruct.Module {
	return &starlarkstruct.Module{
		Name: "container",
		Members: starlark.StringDict{
			"instance": starlark.NewBuiltin("container.instance", builtinContainerInstance),
		},
	}
}

// container.instance(name, image, state?, restart?, ports?, env?, desc?, on_change?)
func builtinContainerInstance(
	thread *starlark.Thread,
	_ *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	var (
		name        string
		image       string
		state       = "running"
		restart     = "unless-stopped"
		portsVal    *starlark.List
		envVal      *starlark.Dict
		desc        string
		onChangeVal starlark.Value
	)
	if err := starlark.UnpackArgs("container.instance", args, kwargs,
		"name", &name,
		"image?", &image,
		"state?", &state,
		"restart?", &restart,
		"ports?", &portsVal,
		"env?", &envVal,
		"desc?", &desc,
		"on_change?", &onChangeVal,
	); err != nil {
		return nil, err
	}

	hookIDs, err := unpackOnChange(thread, onChangeVal, "container.instance")
	if err != nil {
		return nil, err
	}

	var ports []string
	if portsVal != nil {
		ports, err = stringList(portsVal, "container.instance", "ports")
		if err != nil {
			return nil, err
		}
	}

	var env map[string]string
	if envVal != nil {
		env, err = starlarkDictToStringMap(envVal, "container.instance env")
		if err != nil {
			return nil, err
		}
	}

	span := callSpan(thread)
	fields := kwargsFieldSpans(thread, "name", "image", "state", "restart", "ports", "env")

	return &StarlarkStep{
		Instance: spec.StepInstance{
			Desc: desc,
			Type: container.Instance{},
			Config: &container.InstanceConfig{
				Desc: desc, Name: name, Image: image,
				State: state, Restart: restart, Ports: ports,
				Env: env,
			},
			OnChange: hookIDs,
			Source:   span,
			Fields:   fields,
		},
	}, nil
}
