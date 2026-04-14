// SPDX-License-Identifier: GPL-3.0-only

package integration

import (
	"bytes"
	"errors"
	"path/filepath"
	"testing"

	"scampi.dev/scampi/diagnostic"
	"scampi.dev/scampi/engine"
	"scampi.dev/scampi/gen"
	"scampi.dev/scampi/test/harness"
)

func TestGenAPI(t *testing.T) {
	root := harness.AbsPath("../testdata/gen-api")
	entries := harness.ReadDirOrDie(root)

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		dir := filepath.Join(root, name)

		t.Run(name, func(t *testing.T) {
			specPath := findGenSpec(t, dir)
			expectStarPath := filepath.Join(dir, "expected.scampi")
			expectJSONPath := filepath.Join(dir, "expected.json")

			expect := harness.LoadExpected(t, expectJSONPath)

			rec := &harness.RecordingDisplayer{}
			em := diagnostic.NewEmitter(diagnostic.Policy{}, rec)

			var buf bytes.Buffer
			err := gen.API(specPath, "test", &buf, em, gen.APIOptions{})

			if expect.Abort {
				var abort engine.AbortError
				if !errors.As(err, &abort) {
					t.Fatalf("expected AbortError, got %v", err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			harness.AssertDiagnostics(t, rec, expect.Diagnostics, specPath)

			if !expect.Abort {
				expectedStar := harness.ReadOrDie(expectStarPath)
				if got := buf.String(); got != string(expectedStar) {
					t.Fatalf(
						"output mismatch:\n--- got ---\n%s\n--- want ---\n%s",
						got,
						expectedStar,
					)
				}
			}
		})
	}
}

func findGenSpec(t *testing.T, dir string) string {
	t.Helper()
	for _, name := range []string{"spec.yaml", "spec.json"} {
		p := filepath.Join(dir, name)
		if _, err := harness.ReadFileSafe(p); err == nil {
			return p
		}
	}
	t.Fatalf("no spec.yaml or spec.json in %s", dir)
	return ""
}
