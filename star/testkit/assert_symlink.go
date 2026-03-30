// SPDX-License-Identifier: GPL-3.0-only

package testkit

import (
	"context"
	"fmt"
	"io/fs"

	"go.starlark.net/starlark"

	"scampi.dev/scampi/errs"
	"scampi.dev/scampi/target"
)

var symlinkAssertionAttrs = []string{
	"absent",
	"points_to",
}

// SymlinkAssertion is the Starlark value returned by assert_that.symlink(path).
type SymlinkAssertion struct {
	tgt       target.Target
	path      string
	collector *Collector
}

func (a *SymlinkAssertion) String() string        { return fmt.Sprintf("symlink_assertion(%s)", a.path) }
func (a *SymlinkAssertion) Type() string          { return "symlink_assertion" }
func (a *SymlinkAssertion) Freeze()               {}
func (a *SymlinkAssertion) Truth() starlark.Bool  { return starlark.True }
func (a *SymlinkAssertion) Hash() (uint32, error) { return 0, nil }
func (a *SymlinkAssertion) AttrNames() []string   { return symlinkAssertionAttrs }

func (a *SymlinkAssertion) Attr(name string) (starlark.Value, error) {
	switch name {
	case "points_to":
		return starlark.NewBuiltin("symlink.points_to", a.builtinPointsTo), nil
	case "absent":
		return starlark.NewBuiltin("symlink.absent", a.builtinAbsent), nil
	}
	return nil, nil
}

func (a *SymlinkAssertion) builtinPointsTo(
	_ *starlark.Thread,
	_ *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	var targetPath string
	if err := starlark.UnpackPositionalArgs("points_to", args, kwargs, 1, &targetPath); err != nil {
		return nil, err
	}
	path := a.path
	tgt := a.tgt
	a.collector.Add(Assertion{
		Description: fmt.Sprintf("symlink %s points to %s", path, targetPath),
		Check: func() error {
			sl := target.Must[target.Symlink]("symlink.points_to", tgt)
			got, err := sl.Readlink(context.Background(), path)
			if err != nil {
				// bare-error: assertion result
				return errs.Errorf("symlink %s: %s", path, err)
			}
			if got != targetPath {
				// bare-error: assertion result
				return errs.Errorf("symlink %s points to %q, want %q", path, got, targetPath)
			}
			return nil
		},
	})
	return starlark.None, nil
}

func (a *SymlinkAssertion) builtinAbsent(
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
		Description: fmt.Sprintf("symlink %s absent", path),
		Check: func() error {
			sl := target.Must[target.Symlink]("symlink.absent", tgt)
			info, err := sl.Lstat(context.Background(), path)
			if err != nil {
				return nil
			}
			if info.Mode().Type() == fs.ModeSymlink {
				// bare-error: assertion result
				return errs.Errorf("symlink %s exists but should be absent", path)
			}
			return nil
		},
	})
	return starlark.None, nil
}
