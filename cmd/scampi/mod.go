// SPDX-License-Identifier: GPL-3.0-only

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
			modAddCmd(),
			modDownloadCmd(),
			modUpdateCmd(),
			modVerifyCmd(),
			modCacheCmd(),
			modCleanCmd(),
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

// scampi mod add
// -----------------------------------------------------------------------------

func modAddCmd() *cli.Command {
	var moduleArg string

	return &cli.Command{
		Name:         "add",
		Usage:        "Add a dependency to scampi.mod",
		ArgsUsage:    "<module[@version]>",
		OnUsageError: onUsageError,
		Before:       requireArgs(1),
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:        "module",
				Config:      cli.StringConfig{TrimSpace: true},
				Destination: &moduleArg,
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

			modPath, version := parseModArg(moduleArg)
			cacheDir := mod.DefaultCacheDir()

			resolved, err := mod.Add(modPath, version, dir, cacheDir)
			if err != nil {
				emitModDiagnostic(em, err)
				return handleEngineError("mod add", engine.AbortError{Causes: []error{err}})
			}
			emitModInfo(em, fmt.Sprintf("added %s@%s", modPath, resolved))
			return nil
		},
	}
}

// scampi mod download
// -----------------------------------------------------------------------------

func modDownloadCmd() *cli.Command {
	return &cli.Command{
		Name:         "download",
		Usage:        "Download all dependencies listed in scampi.mod",
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

			modFile := filepath.Join(dir, "scampi.mod")
			data, err := os.ReadFile(modFile)
			if err != nil {
				e := &mod.TidyError{Detail: "could not read scampi.mod: " + err.Error(), Hint: "run: scampi mod init"}
				emitModDiagnostic(em, e)
				return handleEngineError("mod download", engine.AbortError{Causes: []error{e}})
			}

			m, err := mod.Parse(modFile, data)
			if err != nil {
				emitModDiagnostic(em, err)
				return handleEngineError("mod download", engine.AbortError{Causes: []error{err}})
			}

			sumFile := filepath.Join(dir, "scampi.sum")
			sums, err := mod.ReadSum(sumFile)
			if err != nil {
				emitModDiagnostic(em, err)
				return handleEngineError("mod download", engine.AbortError{Causes: []error{err}})
			}

			cacheDir := mod.DefaultCacheDir()
			updated := false

			for _, dep := range m.Require {
				if err := mod.Fetch(dep, cacheDir); err != nil {
					emitModDiagnostic(em, err)
					return handleEngineError("mod download", engine.AbortError{Causes: []error{err}})
				}

				dest := filepath.Join(cacheDir, dep.Path+"@"+dep.Version)
				if err := mod.ValidateEntryPoint(dep, dest); err != nil {
					emitModDiagnostic(em, err)
					return handleEngineError("mod download", engine.AbortError{Causes: []error{err}})
				}

				hash, err := mod.ComputeHash(dest)
				if err != nil {
					emitModDiagnostic(em, err)
					return handleEngineError("mod download", engine.AbortError{Causes: []error{err}})
				}

				key := dep.Path + " " + dep.Version
				if sums[key] == "" {
					sums[key] = hash
					updated = true
				}

				emitModInfo(em, fmt.Sprintf("downloaded %s@%s", dep.Path, dep.Version))
			}

			if updated {
				if err := mod.WriteSum(sumFile, sums); err != nil {
					emitModDiagnostic(em, err)
					return handleEngineError("mod download", engine.AbortError{Causes: []error{err}})
				}
			}

			if len(m.Require) == 0 {
				emitModInfo(em, "all modules up to date")
			}

			return nil
		},
	}
}

// scampi mod update
// -----------------------------------------------------------------------------

func modUpdateCmd() *cli.Command {
	var moduleArg string

	return &cli.Command{
		Name:         "update",
		Usage:        "Update a dependency to its latest stable version",
		ArgsUsage:    "<module>",
		OnUsageError: onUsageError,
		Before:       requireArgs(1),
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:        "module",
				Config:      cli.StringConfig{TrimSpace: true},
				Destination: &moduleArg,
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

			modPath, _ := parseModArg(moduleArg)
			cacheDir := mod.DefaultCacheDir()

			resolved, err := mod.Add(modPath, "", dir, cacheDir)
			if err != nil {
				emitModDiagnostic(em, err)
				return handleEngineError("mod update", engine.AbortError{Causes: []error{err}})
			}
			emitModInfo(em, fmt.Sprintf("updated %s to %s", modPath, resolved))
			return nil
		},
	}
}

// scampi mod verify
// -----------------------------------------------------------------------------

func modVerifyCmd() *cli.Command {
	return &cli.Command{
		Name:         "verify",
		Usage:        "Verify that cached modules match their checksums in scampi.sum",
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

			modFile := filepath.Join(dir, "scampi.mod")
			data, err := os.ReadFile(modFile)
			if err != nil {
				e := &mod.TidyError{Detail: "could not read scampi.mod: " + err.Error(), Hint: "run: scampi mod init"}
				emitModDiagnostic(em, e)
				return handleEngineError("mod verify", engine.AbortError{Causes: []error{e}})
			}

			m, err := mod.Parse(modFile, data)
			if err != nil {
				emitModDiagnostic(em, err)
				return handleEngineError("mod verify", engine.AbortError{Causes: []error{err}})
			}

			sumFile := filepath.Join(dir, "scampi.sum")
			sums, err := mod.ReadSum(sumFile)
			if err != nil {
				emitModDiagnostic(em, err)
				return handleEngineError("mod verify", engine.AbortError{Causes: []error{err}})
			}

			cacheDir := mod.DefaultCacheDir()

			for _, dep := range m.Require {
				modDir := filepath.Join(cacheDir, dep.Path+"@"+dep.Version)
				if err := mod.ValidateEntryPoint(dep, modDir); err != nil {
					emitModDiagnostic(em, err)
					return handleEngineError("mod verify", engine.AbortError{Causes: []error{err}})
				}
				if err := mod.VerifyModule(dep, modDir, sums); err != nil {
					emitModDiagnostic(em, err)
					return handleEngineError("mod verify", engine.AbortError{Causes: []error{err}})
				}
			}

			emitModInfo(em, "all modules verified")
			return nil
		},
	}
}

// scampi mod cache
// -----------------------------------------------------------------------------

func modCacheCmd() *cli.Command {
	return &cli.Command{
		Name:         "cache",
		Usage:        "Print the module cache directory",
		OnUsageError: onUsageError,
		Action: func(ctx context.Context, _ *cli.Command) error {
			opts := mustGlobalOpts(ctx)

			displ, cleanup := withDisplayer(opts, nil)
			defer cleanup()

			pol := diagnostic.Policy{Verbosity: opts.verbosity}
			em := diagnostic.NewEmitter(pol, displ)

			emitModInfo(em, mod.DefaultCacheDir())
			return nil
		},
	}
}

// scampi mod clean
// -----------------------------------------------------------------------------

func modCleanCmd() *cli.Command {
	return &cli.Command{
		Name:         "clean",
		Usage:        "Remove all cached modules",
		OnUsageError: onUsageError,
		Action: func(ctx context.Context, _ *cli.Command) error {
			opts := mustGlobalOpts(ctx)

			displ, cleanup := withDisplayer(opts, nil)
			defer cleanup()

			pol := diagnostic.Policy{Verbosity: opts.verbosity}
			em := diagnostic.NewEmitter(pol, displ)

			cacheDir := mod.DefaultCacheDir()
			if err := os.RemoveAll(cacheDir); err != nil {
				e := &mod.TidyError{
					Detail: "could not remove cache directory: " + err.Error(),
					Hint:   "check that " + cacheDir + " is accessible",
				}
				emitModDiagnostic(em, e)
				return handleEngineError("mod clean", engine.AbortError{Causes: []error{e}})
			}
			emitModInfo(em, fmt.Sprintf("cache cleared: %s", cacheDir))
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

func parseModArg(s string) (string, string) {
	if i := strings.LastIndex(s, "@"); i >= 0 {
		return s[:i], s[i+1:]
	}
	return s, ""
}
