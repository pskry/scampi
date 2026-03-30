// SPDX-License-Identifier: GPL-3.0-only

package testkit

import (
	"context"
	"fmt"

	"go.starlark.net/starlark"

	"scampi.dev/scampi/errs"
	"scampi.dev/scampi/target"
)

var dirAssertionAttrs = []string{
	"absent",
	"exists",
	"has_mode",
}

// DirAssertion is the Starlark value returned by assert_that.dir(path).
type DirAssertion struct {
	tgt       target.Target
	path      string
	collector *Collector
}

func (a *DirAssertion) String() string        { return fmt.Sprintf("dir_assertion(%s)", a.path) }
func (a *DirAssertion) Type() string          { return "dir_assertion" }
func (a *DirAssertion) Freeze()               {}
func (a *DirAssertion) Truth() starlark.Bool  { return starlark.True }
func (a *DirAssertion) Hash() (uint32, error) { return 0, nil }
func (a *DirAssertion) AttrNames() []string   { return dirAssertionAttrs }

func (a *DirAssertion) Attr(name string) (starlark.Value, error) {
	switch name {
	case "exists":
		return starlark.NewBuiltin("dir.exists", a.builtinExists), nil
	case "absent":
		return starlark.NewBuiltin("dir.absent", a.builtinAbsent), nil
	case "has_mode":
		return starlark.NewBuiltin("dir.has_mode", a.builtinHasMode), nil
	}
	return nil, nil
}

func (a *DirAssertion) builtinExists(
	_ *starlark.Thread,
	_ *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs("exists", args, kwargs, 0); err != nil {
		return nil, err
	}
	path := a.path
	tgt := a.tgt
	a.collector.Add(Assertion{
		Description: fmt.Sprintf("dir %s exists", path),
		Check: func() error {
			fsys := target.Must[target.Filesystem]("dir.exists", tgt)
			info, err := fsys.Stat(context.Background(), path)
			if err != nil {
				// bare-error: assertion result
				return errs.Errorf("dir %s does not exist", path)
			}
			if !info.IsDir() {
				// bare-error: assertion result
				return errs.Errorf("%s is not a directory", path)
			}
			return nil
		},
	})
	return starlark.None, nil
}

func (a *DirAssertion) builtinAbsent(
	_ *starlark.Thread,
	_ *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs("absent", args, kwargs, 0); err != nil {
		return nil, err
	}
	path := a.path
	tgt := a.tgt
	a.collector.Add(Assertion{
		Description: fmt.Sprintf("dir %s absent", path),
		Check: func() error {
			fsys := target.Must[target.Filesystem]("dir.absent", tgt)
			_, err := fsys.Stat(context.Background(), path)
			if err == nil {
				// bare-error: assertion result
				return errs.Errorf("dir %s exists but should be absent", path)
			}
			return nil
		},
	})
	return starlark.None, nil
}

func (a *DirAssertion) builtinHasMode(
	_ *starlark.Thread,
	_ *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	var modeStr string
	if err := starlark.UnpackPositionalArgs("has_mode", args, kwargs, 1, &modeStr); err != nil {
		return nil, err
	}
	expected, err := parseMode(modeStr)
	if err != nil {
		return nil, err
	}
	path := a.path
	tgt := a.tgt
	a.collector.Add(Assertion{
		Description: fmt.Sprintf("dir %s has mode %s", path, modeStr),
		Check: func() error {
			fsys := target.Must[target.Filesystem]("dir.has_mode", tgt)
			info, err := fsys.Stat(context.Background(), path)
			if err != nil {
				// bare-error: assertion result
				return errs.Errorf("dir %s: %s", path, err)
			}
			got := info.Mode().Perm()
			if got != expected {
				// bare-error: assertion result
				return errs.Errorf("dir %s mode: got %04o, want %04o", path, got, expected)
			}
			return nil
		},
	})
	return starlark.None, nil
}
