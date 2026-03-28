#!/usr/bin/env bash
# SPDX-License-Identifier: GPL-3.0-only
#
# List Codeberg milestones.
# Usage: codeberg-list-milestones.sh <api> <repo> [--closed]
set -euo pipefail
: "${CODEBERG_TOKEN:?Set CODEBERG_TOKEN to a Codeberg personal access token}"

api="$1"; repo="$2"; shift 2
state="open"
for arg in "$@"; do
  case "$arg" in
    --closed) state="closed" ;;
  esac
done

curl -sf \
  -H "Authorization: token $CODEBERG_TOKEN" \
  "$api/repos/$repo/milestones?state=$state&limit=50" \
  | jq -r '.[] | "#\(.id)  \(.title)  [\(.open_issues) open, \(.closed_issues) closed]"'
