#!/usr/bin/env bash
# SPDX-License-Identifier: GPL-3.0-only
#
# bench/bench.sh — hyperfine timings of scampi vs ansible against
# the LXCs created by bench/provision.sh.
#
# Each tool runs N times in two phases:
#   - cold: target rolled back to the `pristine` snapshot before
#     every run (uncounted prep). Each cold timing measures a true
#     from-scratch deploy.
#   - warm: one untimed warmup, then N timed re-runs against an
#     already-converged target (idempotent / no-op path).
#
# Tunables (env vars; can also be set in bench/.env):
#   RUNS=N            — timed runs per (tool × phase). Default 10.
#   PHASES="cold warm" — drop one to skip a phase.

set -euo pipefail

# shellcheck source=bench/lib.sh
source "$(dirname "$0")/lib.sh"

require hyperfine
require_env_basics
require_scampi
require_ansible
export_ssh_env

mkdir -p "$RESULTS_DIR"
TS=$(date +%Y%m%d-%H%M%S)
META="$RESULTS_DIR/$TS.metadata.txt"

# ----- record run metadata --------------------------------------------------

{
    echo "# bench run metadata"
    echo "timestamp:    $TS"
    echo "host:         $(uname -a)"
    echo "scampi:       $("$SCAMPI" version 2>&1 | head -1)"
    echo "ansible:      $("$ANSIBLE" --version 2>&1 | head -1)"
    echo "hyperfine:    $(hyperfine --version 2>&1 | head -1)"
    echo "runs:         $RUNS"
    echo "phases:       $PHASES"
    echo "pve_host:     $PVE_HOST"
    echo "pve_node:     $PVE_NODE"
    echo "vmids:        ${BENCH_VMID_BASE} $((BENCH_VMID_BASE + 1)) $((BENCH_VMID_BASE + 2))"
    echo "ips:          ${BENCH_IP_PREFIX}.${BENCH_IP_BASE} ${BENCH_IP_PREFIX}.$((BENCH_IP_BASE + 1)) ${BENCH_IP_PREFIX}.$((BENCH_IP_BASE + 2))"
    echo "snapshot:     $SNAPSHOT_NAME"
} | tee "$META"

# ----- helpers --------------------------------------------------------------

run_cold() {
    local tool="$1" cmd="$2"
    local out_json="$RESULTS_DIR/$TS.cold.$tool.json"
    local out_md="$RESULTS_DIR/$TS.cold.$tool.md"
    echo
    echo "=== $tool / cold ==="
    hyperfine \
        --runs "$RUNS" \
        --warmup 0 \
        --prepare "bash bench/_cold_prep.sh" \
        --command-name "$tool-cold" \
        --export-json "$out_json" \
        --export-markdown "$out_md" \
        "$cmd"
}

run_warm() {
    local tool="$1" cmd="$2"
    local out_json="$RESULTS_DIR/$TS.warm.$tool.json"
    local out_md="$RESULTS_DIR/$TS.warm.$tool.md"
    echo
    echo "=== $tool / warm ==="
    hyperfine \
        --runs "$RUNS" \
        --warmup 1 \
        --command-name "$tool-warm" \
        --export-json "$out_json" \
        --export-markdown "$out_md" \
        "$cmd"
}

# Per-tool command that hyperfine times. Each runs via hyperfine's
# own shell, which inherits cwd (repo root, set by lib.sh) plus the
# exported env. ANSIBLE_CONFIG is set explicitly so we don't need to
# cd into bench/ansible just to pick up ansible.cfg.
tool_cmd() {
    case "$1" in
        scampi)  echo "$SCAMPI apply bench/scampi/deploy.scampi" ;;
        ansible) echo "ANSIBLE_CONFIG=bench/ansible/ansible.cfg $ANSIBLE -i bench/ansible/inventory.ini bench/ansible/site.yml" ;;
        *)       echo "unknown tool: $1" >&2; exit 1 ;;
    esac
}

# ----- main loop ------------------------------------------------------------

for phase in $PHASES; do
    for tool in scampi ansible; do
        cmd=$(tool_cmd "$tool")
        case "$phase" in
            cold) run_cold "$tool" "$cmd" ;;
            warm) run_warm "$tool" "$cmd" ;;
        esac
    done
done

echo
echo "done. results under $RESULTS_DIR/$TS.*"
echo "  - .metadata.txt        — versions + tunables for this run"
echo "  - .{cold,warm}.*.json  — hyperfine raw + statistics"
echo "  - .{cold,warm}.*.md    — hyperfine markdown summary"
