// SPDX-License-Identifier: GPL-3.0-only

package main

import (
	"context"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"scampi.dev/scampi/diagnostic"
	"scampi.dev/scampi/engine"
	"scampi.dev/scampi/source"
	"scampi.dev/scampi/spec"
	"scampi.dev/scampi/star"
	"scampi.dev/scampi/star/testkit"
)

func testCmd() *cli.Command {
	var testPath string

	return &cli.Command{
		Name:         "test",
		Usage:        "Run Starlark test files",
		ArgsUsage:    "[path]",
		OnUsageError: onUsageError,
		Before:       requireMaxArgs(1),
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:        "path",
				Config:      cli.StringConfig{TrimSpace: true},
				Destination: &testPath,
			},
		},
		Action: func(ctx context.Context, _ *cli.Command) error {
			opts := mustGlobalOpts(ctx)
			store := diagnostic.NewSourceStore()
			displ, cleanup := withDisplayer(opts, store)
			defer cleanup()

			pol := diagnostic.Policy{Verbosity: opts.verbosity}
			em := diagnostic.NewEmitter(pol, displ)
			src := source.LocalPosixSource{}

			pattern := testPath
			if pattern == "" {
				pattern = "*_test.star"
			}

			files, _ := filepath.Glob(pattern)
			if len(files) == 0 {
				emitTestDiag(em, &testkit.TestError{
					Detail: "no test files found matching " + pattern,
					Hint:   "test files must end in _test.star",
				})
				return cli.Exit("", exitUserError)
			}

			totalPassed, totalFailed := 0, 0

			for _, f := range files {
				passed, failed, err := runTestFile(ctx, em, f, store, src)
				if err != nil {
					emitTestDiag(em, &testkit.TestError{
						Detail: err.Error(),
						Hint:   "fix the test file",
					})
					totalFailed++
					continue
				}
				totalPassed += passed
				totalFailed += failed

				emitTestInfo(em, f, &testkit.TestSummary{
					Passed: passed,
					Failed: failed,
					File:   f,
				})
			}

			if totalFailed > 0 {
				return cli.Exit("", exitUserError)
			}
			return nil
		},
	}
}

func runTestFile(
	ctx context.Context,
	em diagnostic.Emitter,
	testPath string,
	store *diagnostic.SourceStore,
	src source.Source,
) (passed, failed int, err error) {
	collector := testkit.NewCollector()

	cfg, err := star.Eval(
		ctx,
		testPath,
		store,
		src,
		star.WithTestBuiltins(collector),
	)
	if err != nil {
		return 0, 0, err
	}

	resolved, err := engine.ResolveMultiple(cfg, spec.ResolveOptions{})
	if err != nil {
		return 0, 0, err
	}

	for _, rc := range resolved {
		e, engineErr := engine.New(ctx, src, rc, em)
		if engineErr != nil {
			return 0, 0, engineErr
		}
		if applyErr := e.Apply(ctx); applyErr != nil {
			e.Close()
			return 0, 0, applyErr
		}
		e.Close()
	}

	for _, assertion := range collector.Assertions() {
		if checkErr := assertion.Check(); checkErr != nil {
			failed++
			em.EmitEngineDiagnostic(diagnostic.RaiseEngineDiagnostic(
				testPath,
				&testkit.TestFail{
					Description: assertion.Description,
					Expected:    "pass",
					Actual:      checkErr.Error(),
					Source:      assertion.Source,
				},
			))
		} else {
			passed++
		}
	}

	return passed, failed, nil
}

func emitTestDiag(em diagnostic.Emitter, d diagnostic.Diagnostic) {
	em.EmitEngineDiagnostic(diagnostic.RaiseEngineDiagnostic("", d))
}

func emitTestInfo(em diagnostic.Emitter, path string, d diagnostic.Diagnostic) {
	em.EmitEngineDiagnostic(diagnostic.RaiseEngineDiagnostic(path, d))
}
