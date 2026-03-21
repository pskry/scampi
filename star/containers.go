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

// container.instance(name, image, state?, restart?, ports?, desc?, on_change?)
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
		desc        string
		onChangeVal starlark.Value
	)
	if err := starlark.UnpackArgs("container.instance", args, kwargs,
		"name", &name,
		"image?", &image,
		"state?", &state,
		"restart?", &restart,
		"ports?", &portsVal,
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

	span := callSpan(thread)
	fields := kwargsFieldSpans(thread, "name", "image", "state", "restart", "ports")

	return &StarlarkStep{
		Instance: spec.StepInstance{
			Desc: desc,
			Type: container.Instance{},
			Config: &container.InstanceConfig{
				Desc: desc, Name: name, Image: image,
				State: state, Restart: restart, Ports: ports,
			},
			OnChange: hookIDs,
			Source:   span,
			Fields:   fields,
		},
	}, nil
}
