#!/usr/bin/env bash
# SPDX-License-Identifier: GPL-3.0-only
#
# Show full details for a Codeberg issue.
# Usage: codeberg-show-issue.sh <api> <repo> <number> [--refresh]
set -euo pipefail

api="$1"; repo="$2"; number="$3"; shift 3
dir="$(cd "$(dirname "$0")" && pwd)"

"$dir/codeberg-fetch.sh" "$api/repos/$repo/issues/$number" "$@" \
  | jq -r '"#\(.number): \(.title)\nState: \(.state)\nLabels: \(.labels | map(.name) | join(", "))\nURL: \(.html_url)\n\n\(.body)"'

comments=$("$dir/codeberg-fetch.sh" "$api/repos/$repo/issues/$number/comments" "$@")

count=$(echo "$comments" | jq 'length')
if [[ "$count" -gt 0 ]]; then
  echo ""
  echo "--- comments ($count) ---"
  echo "$comments" | jq -r '.[] | "\n@\(.user.login) (\(.created_at)):\n\(.body)"'
fi
