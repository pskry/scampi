// SPDX-License-Identifier: GPL-3.0-only

package mod

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"scampi.dev/scampi/source"
)

// Init creates a scampi.mod file in dir with the given module path.
// If modulePath is empty, it is inferred from the git remote origin URL.
func Init(ctx context.Context, src source.Source, dir string, modulePath string) error {
	if modulePath == "" {
		inferred, err := inferModulePath(dir)
		if err != nil {
			return err
		}
		modulePath = inferred
	}

	if !isModulePath(modulePath) {
		return &InitError{
			Detail: fmt.Sprintf("invalid module path %q", modulePath),
			Hint:   "module path must be a host/path URL, e.g. codeberg.org/yourname/yourmodule",
		}
	}

	dest := filepath.Join(dir, "scampi.mod")
	meta, err := src.Stat(ctx, dest)
	if err != nil {
		return &InitStatError{Path: dest, Cause: err}
	}
	if meta.Exists {
		return &InitError{
			Detail: "scampi.mod already exists",
			Hint:   "delete it first or edit it directly",
		}
	}

	content := "module " + modulePath + "\n"
	if err := src.WriteFile(ctx, dest, []byte(content)); err != nil {
		return &InitError{
			Detail: fmt.Sprintf("could not write scampi.mod: %v", err),
			Hint:   "check directory permissions",
		}
	}

	return nil
}

func inferModulePath(dir string) (string, error) {
	out, err := exec.Command("git", "-C", dir, "remote", "get-url", "origin").Output()
	if err != nil {
		return "", &InitError{
			Detail: "could not infer module path from git remote",
			Hint:   "specify it explicitly: scampi mod init <module-path>",
		}
	}
	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return "", &InitError{
			Detail: "git remote origin is empty",
			Hint:   "specify it explicitly: scampi mod init <module-path>",
		}
	}
	return urlToModulePath(raw), nil
}

func urlToModulePath(raw string) string {
	for _, prefix := range []string{"https://", "http://", "git://"} {
		if after, ok := strings.CutPrefix(raw, prefix); ok {
			raw = after
			break
		}
	}

	if after, ok := strings.CutPrefix(raw, "git@"); ok {
		raw = strings.Replace(after, ":", "/", 1)
	}

	raw = strings.TrimSuffix(raw, ".git")
	raw = strings.TrimRight(raw, "/")

	return raw
}
