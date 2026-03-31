// SPDX-License-Identifier: GPL-3.0-only

package mod

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Fetch clones dep into <cacheDir>/<dep.Path>@<dep.Version>/.
// If the destination already exists, Fetch is a no-op.
// When vanity import resolution maps the module to a subdirectory
// within a repo, only that subdirectory is kept in the cache.
// On success the .git directory is removed.
func Fetch(dep Dependency, cacheDir string) error {
	dest := filepath.Join(cacheDir, dep.Path+"@"+dep.Version)

	if _, err := os.Stat(dest); err == nil {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return &FetchError{
			ModPath: dep.Path,
			Version: dep.Version,
			Detail:  fmt.Sprintf("could not create cache directory: %v", err),
			Hint:    "check that " + cacheDir + " is writable",
		}
	}

	rm := resolveModule(dep.Path)

	cloneDest := dest
	if rm.Subdir != "" {
		var err error
		cloneDest, err = os.MkdirTemp("", "scampi-fetch-*")
		if err != nil {
			return &FetchError{
				ModPath: dep.Path,
				Version: dep.Version,
				Detail:  fmt.Sprintf("could not create temp dir: %v", err),
			}
		}
		defer func() { _ = os.RemoveAll(cloneDest) }()
	}

	//nolint:gosec // args are derived from the parsed module manifest, not user input
	cmd := exec.Command(
		"git",
		"clone",
		"--depth=1",
		"--branch",
		dep.Version,
		"--single-branch",
		rm.URL,
		cloneDest,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		_ = os.RemoveAll(cloneDest)
		return &FetchError{
			ModPath: dep.Path,
			Version: dep.Version,
			Detail:  firstLine(out),
			Hint:    "check that version " + dep.Version + " exists in " + rm.URL,
		}
	}

	if rm.Subdir != "" {
		subdirPath := filepath.Join(cloneDest, rm.Subdir)
		info, err := os.Stat(subdirPath)
		if err != nil || !info.IsDir() {
			return &FetchError{
				ModPath: dep.Path,
				Version: dep.Version,
				Detail: fmt.Sprintf(
					"subdirectory %s not found in repository",
					rm.Subdir,
				),
				Hint: "check that the module path matches a directory in " + rm.URL,
			}
		}
		if err := os.Rename(subdirPath, dest); err != nil {
			return &FetchError{
				ModPath: dep.Path,
				Version: dep.Version,
				Detail:  fmt.Sprintf("could not extract subdirectory: %v", err),
			}
		}
	} else {
		if err := os.RemoveAll(filepath.Join(dest, ".git")); err != nil {
			_ = os.RemoveAll(dest)
			return &FetchError{
				ModPath: dep.Path,
				Version: dep.Version,
				Detail:  fmt.Sprintf("could not remove .git directory: %v", err),
				Hint:    "check permissions on " + dest,
			}
		}
	}

	return nil
}

// firstLine returns the first non-empty line of b, trimmed of whitespace.
func firstLine(b []byte) string {
	first, rest, found := bytes.Cut(b, []byte{'\n'})
	line := string(bytes.TrimSpace(first))
	if line != "" || !found {
		return line
	}
	return string(bytes.TrimSpace(rest))
}
