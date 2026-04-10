#!/usr/bin/env bash
# SPDX-License-Identifier: GPL-3.0-only
# Shared dev runner — rebuilds a binary if Go source changed, then execs it.
# Usage: dev-run.sh <binary> [args...]
#   binary: "scampi" or "scampls"
set -euo pipefail

BIN="$1"; shift
SRCDIR="$(cd "$(dirname "$0")/.." && pwd)"
OUTDIR="$SRCDIR/build/bin"
BINPATH="$OUTDIR/$BIN"

# Rebuild if binary is missing or any source file is newer.
# Includes .go files AND .scampi stubs under std/ — those are
# embedded into the binary at compile time, so editing a stub
# requires a rebuild for the change to take effect.
needs_build=false
if [[ ! -f "$BINPATH" ]]; then
  needs_build=true
else
  newer=$(find "$SRCDIR" \
    \( -name '*.go' -o -path "$SRCDIR/std/*.scampi" \) \
    -newer "$BINPATH" -print -quit 2>/dev/null)
  if [[ -n "$newer" ]]; then
    needs_build=true
  fi
fi

if [[ "$needs_build" == "true" ]]; then
  version="$(cd "$SRCDIR" && git describe --tags --always --dirty 2>/dev/null || echo dev)"
  ldflags="-s -w -X main.version=$version"
  mkdir -p "$OUTDIR"
  (cd "$SRCDIR" && go build -ldflags "$ldflags" -o "$OUTDIR/$BIN" "./cmd/$BIN") >&2
fi

exec "$BINPATH" "$@"
