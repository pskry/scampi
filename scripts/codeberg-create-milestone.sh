#!/usr/bin/env bash
# SPDX-License-Identifier: GPL-3.0-only
#
# Create a Codeberg milestone.
# Usage: codeberg-create-milestone.sh <api> <repo> <title> [description]
set -euo pipefail
: "${CODEBERG_TOKEN:?Set CODEBERG_TOKEN to a Codeberg personal access token}"

api="$1"; repo="$2"; title="$3"; desc="${4:-}"

body=$(jq -n --arg t "$title" --arg d "$desc" '{title: $t, description: $d}')

resp=$(curl -sf \
  -H "Authorization: token $CODEBERG_TOKEN" \
  -H "Content-Type: application/json" \
  -d "$body" \
  "$api/repos/$repo/milestones")

id=$(echo "$resp" | jq -r '.id')
echo "milestone #$id: $title → https://codeberg.org/$repo/milestone/$id"
