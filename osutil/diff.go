package osutil

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RunDiffTool writes current and desired content to temp files and execs the
// given diff tool interactively. The tool string is split by whitespace to
// support multi-word commands like "nvim -d". The temp dir is cleaned up when
// the tool exits.
//
// diff(1) exit code 1 (files differ) is not treated as an error.
func RunDiffTool(ctx context.Context, tool, destPath string, current, desired []byte) error {
	dir, err := os.MkdirTemp("", "doit-inspect-")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(dir) }()

	base := filepath.Base(destPath)
	currentDir := filepath.Join(dir, "current")
	desiredDir := filepath.Join(dir, "desired")

	if err := os.MkdirAll(currentDir, 0o700); err != nil {
		return fmt.Errorf("creating current dir: %w", err)
	}
	if err := os.MkdirAll(desiredDir, 0o700); err != nil {
		return fmt.Errorf("creating desired dir: %w", err)
	}

	currentFile := filepath.Join(currentDir, base)
	desiredFile := filepath.Join(desiredDir, base)

	if err := os.WriteFile(currentFile, current, 0o600); err != nil {
		return fmt.Errorf("writing current file: %w", err)
	}
	if err := os.WriteFile(desiredFile, desired, 0o600); err != nil {
		return fmt.Errorf("writing desired file: %w", err)
	}

	parts := strings.Fields(tool)
	args := append(parts[1:], currentFile, desiredFile)

	cmd := exec.CommandContext(ctx, parts[0], args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// diff(1) exits 1 when files differ — not an error for us.
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil
		}
		return fmt.Errorf("diff tool %q: %w", parts[0], err)
	}

	return nil
}

// ResolveDiffTool picks a diff tool from environment variables.
// Lookup order: DOIT_DIFFTOOL → DIFFTOOL → EDITOR → "diff".
func ResolveDiffTool() string {
	for _, env := range []string{"DOIT_DIFFTOOL", "DIFFTOOL", "EDITOR"} {
		if v := os.Getenv(env); v != "" {
			return v
		}
	}
	return "diff"
}
