// SPDX-License-Identifier: GPL-3.0-only

package testkit

import (
	"context"
	"fmt"
	"io/fs"
	"strings"

	"go.starlark.net/starlark"

	"scampi.dev/scampi/errs"
	"scampi.dev/scampi/target"
)

var fileAssertionAttrs = []string{
	"absent",
	"contains",
	"exists",
	"has_content",
	"has_mode",
}

// FileAssertion is the Starlark value returned by assert_that.file(path).
type FileAssertion struct {
	tgt       target.Target
	path      string
	collector *Collector
}

func (a *FileAssertion) String() string        { return fmt.Sprintf("file_assertion(%s)", a.path) }
func (a *FileAssertion) Type() string          { return "file_assertion" }
func (a *FileAssertion) Freeze()               {}
func (a *FileAssertion) Truth() starlark.Bool  { return starlark.True }
func (a *FileAssertion) Hash() (uint32, error) { return 0, nil }
func (a *FileAssertion) AttrNames() []string   { return fileAssertionAttrs }

func (a *FileAssertion) Attr(name string) (starlark.Value, error) {
	switch name {
	case "has_content":
		return starlark.NewBuiltin("file.has_content", a.builtinHasContent), nil
	case "contains":
		return starlark.NewBuiltin("file.contains", a.builtinContains), nil
	case "exists":
		return starlark.NewBuiltin("file.exists", a.builtinExists), nil
	case "absent":
		return starlark.NewBuiltin("file.absent", a.builtinAbsent), nil
	case "has_mode":
		return starlark.NewBuiltin("file.has_mode", a.builtinHasMode), nil
	}
	return nil, nil
}

func (a *FileAssertion) builtinHasContent(
	_ *starlark.Thread,
	_ *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	var expected string
	if err := starlark.UnpackPositionalArgs("has_content", args, kwargs, 1, &expected); err != nil {
		return nil, err
	}
	path := a.path
	tgt := a.tgt
	a.collector.Add(Assertion{
		Description: fmt.Sprintf("file %s has expected content", path),
		Check: func() error {
			fs := target.Must[target.Filesystem]("file.has_content", tgt)
			data, err := fs.ReadFile(context.Background(), path)
			if err != nil {
				// bare-error: assertion result
				return errs.Errorf("file %s: %s", path, err)
			}
			if string(data) != expected {
				got := string(data)
				// bare-error: assertion result
				return errs.Errorf("file %s content mismatch:\n  got: %q\n  want: %q", path, got, expected)
			}
			return nil
		},
	})
	return starlark.None, nil
}

func (a *FileAssertion) builtinContains(
	_ *starlark.Thread,
	_ *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	var substring string
	if err := starlark.UnpackPositionalArgs("contains", args, kwargs, 1, &substring); err != nil {
		return nil, err
	}
	path := a.path
	tgt := a.tgt
	a.collector.Add(Assertion{
		Description: fmt.Sprintf("file %s contains %q", path, substring),
		Check: func() error {
			fs := target.Must[target.Filesystem]("file.contains", tgt)
			data, err := fs.ReadFile(context.Background(), path)
			if err != nil {
				// bare-error: assertion result
				return errs.Errorf("file %s: %s", path, err)
			}
			if !strings.Contains(string(data), substring) {
				// bare-error: assertion result
				return errs.Errorf("file %s does not contain %q", path, substring)
			}
			return nil
		},
	})
	return starlark.None, nil
}

func (a *FileAssertion) builtinExists(
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
		Description: fmt.Sprintf("file %s exists", path),
		Check: func() error {
			fsys := target.Must[target.Filesystem]("file.exists", tgt)
			info, err := fsys.Stat(context.Background(), path)
			if err != nil {
				// bare-error: assertion result
				return errs.Errorf("file %s does not exist", path)
			}
			if info.IsDir() {
				// bare-error: assertion result
				return errs.Errorf("%s is a directory, not a file", path)
			}
			return nil
		},
	})
	return starlark.None, nil
}

func (a *FileAssertion) builtinAbsent(
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
		Description: fmt.Sprintf("file %s absent", path),
		Check: func() error {
			fsys := target.Must[target.Filesystem]("file.absent", tgt)
			_, err := fsys.Stat(context.Background(), path)
			if err == nil {
				// bare-error: assertion result
				return errs.Errorf("file %s exists but should be absent", path)
			}
			return nil
		},
	})
	return starlark.None, nil
}

func (a *FileAssertion) builtinHasMode(
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
		Description: fmt.Sprintf("file %s has mode %s", path, modeStr),
		Check: func() error {
			fsys := target.Must[target.Filesystem]("file.has_mode", tgt)
			info, err := fsys.Stat(context.Background(), path)
			if err != nil {
				// bare-error: assertion result
				return errs.Errorf("file %s: %s", path, err)
			}
			got := info.Mode().Perm()
			if got != expected {
				// bare-error: assertion result
				return errs.Errorf("file %s mode: got %04o, want %04o", path, got, expected)
			}
			return nil
		},
	})
	return starlark.None, nil
}

// bare-error: Starlark argument validation
func parseMode(s string) (fs.FileMode, error) {
	var mode uint32
	if _, err := fmt.Sscanf(s, "%o", &mode); err != nil {
		// bare-error: argument validation
		return 0, errs.Errorf("invalid mode %q: expected octal (e.g. \"0644\")", s)
	}
	return fs.FileMode(mode), nil
}
