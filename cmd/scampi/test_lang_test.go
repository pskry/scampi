// SPDX-License-Identifier: GPL-3.0-only

package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"scampi.dev/scampi/diagnostic"
	"scampi.dev/scampi/diagnostic/event"
	"scampi.dev/scampi/render"
	"scampi.dev/scampi/source"
	"scampi.dev/scampi/testkit"
)

// nopDisplayer satisfies render.Displayer with no-op writes. Used
// to keep the runner happy in unit tests where we only care about
// the structured (passed, failed, err) return values, not what
// would be printed to the terminal.
type nopDisplayer struct {
	diagnostics []event.EngineDiagnostic
}

func (d *nopDisplayer) EmitEngineLifecycle(event.EngineEvent) {}
func (d *nopDisplayer) EmitPlanLifecycle(event.PlanEvent)     {}
func (d *nopDisplayer) EmitActionLifecycle(event.ActionEvent) {}
func (d *nopDisplayer) EmitOpLifecycle(event.OpEvent)         {}
func (d *nopDisplayer) EmitIndexAll(event.IndexAllEvent)      {}
func (d *nopDisplayer) EmitIndexStep(event.IndexStepEvent)    {}
func (d *nopDisplayer) EmitInspect(event.InspectEvent)        {}
func (d *nopDisplayer) EmitLegend()                           {}
func (d *nopDisplayer) EmitEngineDiagnostic(e event.EngineDiagnostic) {
	d.diagnostics = append(d.diagnostics, e)
}
func (d *nopDisplayer) EmitPlanDiagnostic(event.PlanDiagnostic)     {}
func (d *nopDisplayer) EmitActionDiagnostic(event.ActionDiagnostic) {}
func (d *nopDisplayer) EmitOpDiagnostic(event.OpDiagnostic)         {}
func (d *nopDisplayer) Interrupt()                                  {}
func (d *nopDisplayer) Close()                                      {}

var _ render.Displayer = (*nopDisplayer)(nil)

// writeTestFile writes content to a temp file and returns the path.
func writeTestFile(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return path
}

func newTestEmitter() (diagnostic.Emitter, *nopDisplayer) {
	displ := &nopDisplayer{}
	em := diagnostic.NewEmitter(diagnostic.Policy{}, displ)
	return em, displ
}

func TestRunLangTestFile_Passing(t *testing.T) {
	src := `module main

import "std"
import "std/posix"
import "std/test"
import "std/test/matchers"

let mock = test.target_in_memory(
  name = "mock",
  initial = test.InitialState {
    dirs = ["/etc"],
  },
  expect = test.ExpectedState {
    files = {
      "/etc/foo": matchers.has_substring("hello"),
    },
  },
)

std.deploy(name = "smoke", targets = [mock]) {
  posix.copy {
    dest = "/etc/foo"
    src = posix.source_inline { content = "hello world" }
    perm = "0644"
    owner = "root"
    group = "root"
  }
}
`
	path := writeTestFile(t, "smoke_test.scampi", src)
	em, displ := newTestEmitter()
	passed, failed, err := runLangTestFile(
		context.Background(),
		em,
		path,
		source.LocalPosixSource{},
	)
	if err != nil {
		for _, d := range displ.diagnostics {
			t.Logf("diag: %+v", d)
		}
		t.Fatalf("runLangTestFile: %v", err)
	}
	if passed != 1 || failed != 0 {
		t.Errorf("want 1 passed / 0 failed, got %d/%d (diags=%d)",
			passed, failed, len(displ.diagnostics))
	}
}

func TestRunLangTestFile_Failing(t *testing.T) {
	src := `module main

import "std"
import "std/posix"
import "std/test"
import "std/test/matchers"

let mock = test.target_in_memory(
  name = "mock",
  initial = test.InitialState {
    dirs = ["/etc"],
  },
  expect = test.ExpectedState {
    files = {
      "/etc/foo": matchers.has_exact_content("EXPECTED"),
    },
  },
)

std.deploy(name = "fail", targets = [mock]) {
  posix.copy {
    dest = "/etc/foo"
    src = posix.source_inline { content = "ACTUAL" }
    perm = "0644"
    owner = "root"
    group = "root"
  }
}
`
	path := writeTestFile(t, "fail_test.scampi", src)
	em, displ := newTestEmitter()
	passed, failed, err := runLangTestFile(
		context.Background(),
		em,
		path,
		source.LocalPosixSource{},
	)
	if err != nil {
		t.Fatalf("runLangTestFile: %v", err)
	}
	if passed != 0 || failed != 1 {
		t.Errorf("want 0 passed / 1 failed, got %d/%d", passed, failed)
	}
	if len(displ.diagnostics) == 0 {
		t.Errorf("expected at least one TestFail diagnostic")
	}
}

func TestDetectLangSyntax(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want bool
	}{
		{"plain module", "module main\nlet x = 1\n", true},
		{"comment then module", "// header\nmodule main\n", true},
		{"block comment then module", "/* header */\nmodule main\n", true},
		{"blank lines then module", "\n\nmodule main\n", true},
		{"starlark assignment", "x = 1\nprint(x)\n", false},
		{"starlark function call", "deploy(name=\"x\")\n", false},
		{"empty file", "", false},
		{"only comments", "// nothing else\n# also nothing\n", false},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := detectLangSyntax([]byte(tt.src)); got != tt.want {
				t.Errorf("detectLangSyntax(%q) = %v, want %v", tt.src, got, tt.want)
			}
		})
	}
}

// TestkitImportsTopLevel ensures the testkit package's diagnostic
// types are reachable from the cmd binary so the dispatch path can
// emit them. Pure compile-time check; this would fail in lint not
// at test time, but having an explicit assertion is cheap.
func TestTestkitTypesReachable(t *testing.T) {
	var _ = testkit.TestFail{}
	var _ = testkit.TestSummary{}
	var _ = testkit.TestError{}
	if !strings.Contains(testkit.TestFail{Description: "x"}.Description, "x") {
		t.Fatal("TestFail not addressable")
	}
}
