// SPDX-License-Identifier: GPL-3.0-only

package check

import (
	"io/fs"
	"strings"

	"scampi.dev/scampi/lang/ast"
	"scampi.dev/scampi/lang/lex"
	"scampi.dev/scampi/lang/parse"
)

// BootstrapStd parses and type-checks the standard library stubs from
// the given filesystem. Returns a map of module leaf names to their
// checked scopes: "std", "posix", "rest", "container".
//
// The std module is checked first (it has no imports). Submodules are
// checked with std available as an import.
func BootstrapStd(fsys fs.FS) (map[string]*Scope, error) {
	// Phase 1: parse and check std.scampi (no imports needed).
	stdFile, err := parseStub(fsys, "std.scampi")
	if err != nil {
		return nil, err
	}
	stdChecker := NewWithModules(nil)
	stdChecker.Check(stdFile)
	if errs := stdChecker.Errors(); len(errs) > 0 {
		return nil, errs[0]
	}
	stdScope := stdChecker.FileScope()

	modules := map[string]*Scope{
		"std": stdScope,
	}

	// Phase 2: parse and check each submodule with std available.
	err = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".scampi") || path == "std.scampi" {
			return nil
		}
		f, parseErr := parseStub(fsys, path)
		if parseErr != nil {
			return parseErr
		}
		c := NewWithModules(map[string]*Scope{"std": stdScope})
		c.Check(f)
		if errs := c.Errors(); len(errs) > 0 {
			return errs[0]
		}

		// Module name is the directory name (leaf of path).
		dir := path[:strings.LastIndex(path, "/")]
		if i := strings.LastIndex(dir, "/"); i >= 0 {
			dir = dir[i+1:]
		}
		modules[dir] = c.FileScope()
		return nil
	})
	if err != nil {
		return nil, err
	}

	return modules, nil
}

func parseStub(fsys fs.FS, path string) (*ast.File, error) {
	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil, err
	}
	l := lex.New(path, data)
	p := parse.New(l)
	f := p.Parse()
	if errs := l.Errors(); len(errs) > 0 {
		return nil, errs[0]
	}
	if errs := p.Errors(); len(errs) > 0 {
		return nil, errs[0]
	}
	return f, nil
}
