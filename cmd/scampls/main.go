// SPDX-License-Identifier: GPL-3.0-only

// scampls is the Language Server Protocol server for scampi
// configuration files. It communicates over stdin/stdout using the
// standard LSP JSON-RPC transport.
//
// Usage:
//
//	scampls [--log FILE]
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/urfave/cli/v3"
	"scampi.dev/scampi/errs"
	"scampi.dev/scampi/lsp"
)

var version = "v0.0.0-dev"

func main() {
	app := &cli.Command{
		Name:    "scampls",
		Usage:   "Language Server Protocol server for scampi configs",
		Version: version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "log",
				Usage: "write debug log to file",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			lsp.Version = version

			var opts []lsp.Option
			if logPath := cmd.String("log"); logPath != "" {
				f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
				if err != nil {
					// bare-error: CLI boundary, not reachable through engine
					return errs.Errorf("open log: %w", err)
				}
				defer func() { _ = f.Close() }()
				opts = append(opts, lsp.WithLog(
					log.New(f, "scampls: ", log.Ltime|log.Lmicroseconds),
				))
			}

			return lsp.Serve(ctx, os.Stdin, os.Stdout, opts...)
		},
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := app.Run(ctx, os.Args); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "scampls: %s\n", err)
		os.Exit(1)
	}
}
