// SPDX-License-Identifier: GPL-3.0-only

package std

import (
	"io/fs"
	"strings"
	"testing"

	"scampi.dev/scampi/lang/ast"
	"scampi.dev/scampi/lang/check"
	"scampi.dev/scampi/lang/lex"
	"scampi.dev/scampi/lang/parse"
)

func TestStdLibCompiles(t *testing.T) {
	// Phase 1: parse and check std.scampi to build the root scope.
	stdData, err := fs.ReadFile(FS, "std.scampi")
	if err != nil {
		t.Fatalf("read std.scampi: %v", err)
	}
	stdFile := parseOrFatal(t, "std.scampi", stdData)
	stdChecker := check.New()
	stdChecker.Check(stdFile)
	if errs := stdChecker.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("std.scampi check: %s", e.Msg)
		}
		t.FailNow()
	}

	stdScope := stdChecker.FileScope()

	// Phase 2: parse and check each submodule with std available.
	err = fs.WalkDir(FS, ".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".scampi") {
			return nil
		}
		if path == "std.scampi" {
			return nil
		}
		t.Run(path, func(t *testing.T) {
			data, readErr := fs.ReadFile(FS, path)
			if readErr != nil {
				t.Fatalf("read: %v", readErr)
			}
			f := parseOrFatal(t, path, data)
			c := check.NewWithModules(map[string]*check.Scope{
				"std": stdScope,
			})
			c.Check(f)
			if errs := c.Errors(); len(errs) > 0 {
				for _, e := range errs {
					t.Errorf("check: %s", e.Msg)
				}
			}
		})
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
}

func parseOrFatal(t *testing.T, name string, data []byte) *ast.File {
	t.Helper()
	l := lex.New(name, data)
	p := parse.New(l)
	f := p.Parse()
	if errs := l.Errors(); len(errs) > 0 {
		t.Fatalf("%s lex errors: %v", name, errs)
	}
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("%s parse errors: %v", name, errs)
	}
	return f
}
