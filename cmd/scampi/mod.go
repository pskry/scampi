// SPDX-License-Identifier: GPL-3.0-only

package main

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"

	"scampi.dev/scampi/diagnostic"
	"scampi.dev/scampi/engine"
	"scampi.dev/scampi/errs"
	"scampi.dev/scampi/mod"
)

// scampi mod
// -----------------------------------------------------------------------------

func modCmd() *cli.Command {
	return &cli.Command{
		Name:                   "mod",
		Usage:                  "Manage scampi module dependencies",
		UseShortOptionHandling: true,
		Suggest:                true,
		HideHelp:               false,
		OnUsageError:           onUsageError,
		Commands: []*cli.Command{
			modInitCmd(),
			modTidyCmd(),
		},
	}
}

// scampi mod init
// -----------------------------------------------------------------------------

func modInitCmd() *cli.Command {
	var modulePath string

	return &cli.Command{
		Name:         "init",
		Usage:        "Create a scampi.mod file in the current directory",
		ArgsUsage:    "[module-path]",
		OnUsageError: onUsageError,
		Before:       requireMaxArgs(1),
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:        "module-path",
				Config:      cli.StringConfig{TrimSpace: true},
				Destination: &modulePath,
			},
		},
		Action: func(ctx context.Context, _ *cli.Command) error {
			opts := mustGlobalOpts(ctx)

			displ, cleanup := withDisplayer(opts, nil)
			defer cleanup()

			pol := diagnostic.Policy{Verbosity: opts.verbosity}
			em := diagnostic.NewEmitter(pol, displ)

			dir, err := os.Getwd()
			if err != nil {
				panic(errs.BUG("os.Getwd failed: %w", err))
			}
			if err := mod.Init(dir, modulePath); err != nil {
				emitModDiagnostic(em, err)
				return handleEngineError("mod init", engine.AbortError{Causes: []error{err}})
			}
			emitModInfo(em, "created scampi.mod")
			return nil
		},
	}
}

// scampi mod tidy
// -----------------------------------------------------------------------------

func modTidyCmd() *cli.Command {
	return &cli.Command{
		Name:         "tidy",
		Usage:        "Sync the require block with load() calls in *.star files",
		OnUsageError: onUsageError,
		Action: func(ctx context.Context, _ *cli.Command) error {
			opts := mustGlobalOpts(ctx)

			displ, cleanup := withDisplayer(opts, nil)
			defer cleanup()

			pol := diagnostic.Policy{Verbosity: opts.verbosity}
			em := diagnostic.NewEmitter(pol, displ)

			dir, err := os.Getwd()
			if err != nil {
				panic(errs.BUG("os.Getwd failed: %w", err))
			}
			changes, err := mod.Tidy(dir)
			if err != nil {
				emitModDiagnostic(em, err)
				return handleEngineError("mod tidy", engine.AbortError{Causes: []error{err}})
			}
			if len(changes) == 0 {
				emitModInfo(em, "scampi.mod is up to date")
				return nil
			}
			for _, c := range changes {
				emitModInfo(em, c)
			}
			return nil
		},
	}
}

// Helpers
// -----------------------------------------------------------------------------

func emitModDiagnostic(em diagnostic.Emitter, err error) {
	if d, ok := err.(diagnostic.Diagnostic); ok {
		em.EmitEngineDiagnostic(diagnostic.RaiseEngineDiagnostic("", d))
	}
}

func emitModInfo(em diagnostic.Emitter, detail string) {
	em.EmitEngineDiagnostic(diagnostic.RaiseEngineDiagnostic("", &mod.ModInfo{
		Detail: detail,
	}))
}
