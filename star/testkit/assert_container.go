// SPDX-License-Identifier: GPL-3.0-only

package testkit

import (
	"context"
	"fmt"

	"go.starlark.net/starlark"

	"scampi.dev/scampi/errs"
	"scampi.dev/scampi/target"
)

var containerAssertionAttrs = []string{
	"has_image",
	"is_running",
}

// ContainerAssertion is the Starlark value returned by assert_that.container(name).
type ContainerAssertion struct {
	tgt       target.Target
	name      string
	collector *Collector
}

func (a *ContainerAssertion) String() string {
	return fmt.Sprintf("container_assertion(%s)", a.name)
}
func (a *ContainerAssertion) Type() string          { return "container_assertion" }
func (a *ContainerAssertion) Freeze()               {}
func (a *ContainerAssertion) Truth() starlark.Bool  { return starlark.True }
func (a *ContainerAssertion) Hash() (uint32, error) { return 0, nil }
func (a *ContainerAssertion) AttrNames() []string   { return containerAssertionAttrs }

func (a *ContainerAssertion) Attr(name string) (starlark.Value, error) {
	switch name {
	case "is_running":
		return starlark.NewBuiltin("container.is_running", a.builtinIsRunning), nil
	case "has_image":
		return starlark.NewBuiltin("container.has_image", a.builtinHasImage), nil
	}
	return nil, nil
}

func (a *ContainerAssertion) builtinIsRunning(
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
		Description: fmt.Sprintf("container %s is running", name),
		Check: func() error {
			cm := target.Must[target.ContainerManager]("container.is_running", tgt)
			info, exists, err := cm.InspectContainer(context.Background(), name)
			if err != nil {
				// bare-error: assertion result
				return errs.Errorf("container %s: %s", name, err)
			}
			if !exists {
				// bare-error: assertion result
				return errs.Errorf("container %s does not exist", name)
			}
			if !info.Running {
				// bare-error: assertion result
				return errs.Errorf("container %s is not running", name)
			}
			return nil
		},
	})
	return starlark.None, nil
}

func (a *ContainerAssertion) builtinHasImage(
	_ *starlark.Thread,
	_ *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	var image string
	if err := starlark.UnpackPositionalArgs("has_image", args, kwargs, 1, &image); err != nil {
		return nil, err
	}
	name := a.name
	tgt := a.tgt
	a.collector.Add(Assertion{
		Description: fmt.Sprintf("container %s has image %s", name, image),
		Check: func() error {
			cm := target.Must[target.ContainerManager]("container.has_image", tgt)
			info, exists, err := cm.InspectContainer(context.Background(), name)
			if err != nil {
				// bare-error: assertion result
				return errs.Errorf("container %s: %s", name, err)
			}
			if !exists {
				// bare-error: assertion result
				return errs.Errorf("container %s does not exist", name)
			}
			if info.Image != image {
				// bare-error: assertion result
				return errs.Errorf("container %s image: got %q, want %q", name, info.Image, image)
			}
			return nil
		},
	})
	return starlark.None, nil
}
