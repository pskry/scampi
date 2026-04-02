#!/usr/bin/env bash
# Dev wrapper for scampls — always builds fresh via go run.
# Point your Neovim LSP config cmd at this script.
cd /Users/pskry/dev/scampi || exit 1
exec go run ./cmd/scampls "$@"
