// SPDX-License-Identifier: GPL-3.0-only

package format

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGoldenFiles(t *testing.T) {
	entries, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}

	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".input") {
			continue
		}
		base := strings.TrimSuffix(e.Name(), ".input")
		t.Run(base, func(t *testing.T) {
			input, err := os.ReadFile(filepath.Join("testdata", base+".input"))
			if err != nil {
				t.Fatal(err)
			}
			goldenPath := filepath.Join("testdata", base+".expected")
			golden, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatal(err)
			}

			got, err := Format(input)
			if err != nil {
				t.Fatalf("Format: %v", err)
			}

			if string(got) != string(golden) {
				t.Errorf("output mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, golden)
			}
		})
	}
}

func TestIdempotent(t *testing.T) {
	entries, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}

	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".expected") {
			continue
		}
		base := strings.TrimSuffix(e.Name(), ".expected")
		t.Run(base, func(t *testing.T) {
			golden, err := os.ReadFile(filepath.Join("testdata", e.Name()))
			if err != nil {
				t.Fatal(err)
			}

			got, err := Format(golden)
			if err != nil {
				t.Fatalf("Format: %v", err)
			}

			if string(got) != string(golden) {
				t.Errorf("formatting golden file is not idempotent:\n--- got ---\n%s\n--- want ---\n%s", got, golden)
			}
		})
	}
}
