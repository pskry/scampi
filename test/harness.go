// SPDX-License-Identifier: GPL-3.0-only

package test

import (
	"os"
	"path/filepath"
)

type ExpectedDiagnostics struct {
	Abort       bool                 `json:"abort"`
	Diagnostics []ExpectedDiagnostic `json:"diagnostics"`
}

type ExpectedDiagnostic struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Scope    string `json:"scope"`
	Severity string `json:"severity"`

	Source *ExpectedSource `json:"source,omitempty"`
	Step   *ExpectedStep   `json:"step,omitempty"`
}

type ExpectedSource struct {
	StartLine int `json:"start_line"`
	StartCol  int `json:"start_col"`
	EndLine   int `json:"end_line"`
	EndCol    int `json:"end_col"`
}

type ExpectedStep struct {
	Index int    `json:"index"`
	Kind  string `json:"kind"`
}

func absPath(p string) string {
	r, err := filepath.Abs(p)
	if err != nil {
		panic(err)
	}

	return r
}

func readDirOrDie(name string) []os.DirEntry {
	res, err := os.ReadDir(name)
	if err != nil {
		panic(err)
	}

	return res
}

func readDirSafe(name string) ([]os.DirEntry, error) {
	return os.ReadDir(name)
}

func readFileSafe(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func readOrDie(name string) []byte {
	data, err := os.ReadFile(name)
	if err != nil {
		panic(err)
	}
	return data
}

func writeOrDie(name string, data []byte, perm os.FileMode) {
	if err := os.WriteFile(name, data, perm); err != nil {
		panic(err)
	}
}
