// SPDX-License-Identifier: GPL-3.0-only

package testkit

import (
	"context"
	"fmt"

	"go.starlark.net/starlark"

	"scampi.dev/scampi/errs"
	"scampi.dev/scampi/target"
)

var packageAssertionAttrs = []string{
	"is_absent",
	"is_installed",
}

// PackageAssertion is the Starlark value returned by assert_that.package(name).
type PackageAssertion struct {
	tgt       target.Target
	name      string
	collector *Collector
}

func (a *PackageAssertion) String() string        { return fmt.Sprintf("package_assertion(%s)", a.name) }
func (a *PackageAssertion) Type() string          { return "package_assertion" }
func (a *PackageAssertion) Freeze()               {}
func (a *PackageAssertion) Truth() starlark.Bool  { return starlark.True }
func (a *PackageAssertion) Hash() (uint32, error) { return 0, nil }
func (a *PackageAssertion) AttrNames() []string   { return packageAssertionAttrs }

func (a *PackageAssertion) Attr(name string) (starlark.Value, error) {
	switch name {
	case "is_installed":
		return starlark.NewBuiltin("package.is_installed", a.builtinIsInstalled), nil
	case "is_absent":
		return starlark.NewBuiltin("package.is_absent", a.builtinIsAbsent), nil
	}
	return nil, nil
}

func (a *PackageAssertion) builtinIsInstalled(
	_ *starlark.Thread,
	_ *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs("is_installed", args, kwargs, 0); err != nil {
		return nil, err
	}
	name := a.name
	tgt := a.tgt
	a.collector.Add(Assertion{
		Description: fmt.Sprintf("package %s is installed", name),
		Check: func() error {
			pm := target.Must[target.PkgManager]("package.is_installed", tgt)
			installed, err := pm.IsInstalled(context.Background(), name)
			if err != nil {
				// bare-error: assertion result
				return errs.Errorf("package %s: %s", name, err)
			}
			if !installed {
				// bare-error: assertion result
				return errs.Errorf("package %s is not installed", name)
			}
			return nil
		},
	})
	return starlark.None, nil
}

func (a *PackageAssertion) builtinIsAbsent(
	_ *starlark.Thread,
	_ *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs("is_absent", args, kwargs, 0); err != nil {
		return nil, err
	}
	name := a.name
	tgt := a.tgt
	a.collector.Add(Assertion{
		Description: fmt.Sprintf("package %s is absent", name),
		Check: func() error {
			pm := target.Must[target.PkgManager]("package.is_absent", tgt)
			installed, err := pm.IsInstalled(context.Background(), name)
			if err != nil {
				// bare-error: assertion result
				return errs.Errorf("package %s: %s", name, err)
			}
			if installed {
				// bare-error: assertion result
				return errs.Errorf("package %s is installed but should be absent", name)
			}
			return nil
		},
	})
	return starlark.None, nil
}
