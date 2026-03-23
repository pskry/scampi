#!/usr/bin/env bash
# SPDX-License-Identifier: GPL-3.0-only
#
# Add a label to a Codeberg issue (without removing existing labels).
# Usage: codeberg-add-label.sh <api> <repo> <number> <label-name>
set -euo pipefail
: "${CODEBERG_TOKEN:?Set CODEBERG_TOKEN to a Codeberg personal access token}"

api="$1"; repo="$2"; number="$3"; label_name="$4"

# Resolve label name to ID
label_id=$(curl -sf \
  -H "Authorization: token $CODEBERG_TOKEN" \
  "$api/repos/$repo/labels?limit=50" \
  | jq -r --arg name "$label_name" '.[] | select(.name == $name) | .id')

if [[ -z "$label_id" ]]; then
  echo "error: label \"$label_name\" not found in $repo" >&2
  exit 1
fi

# POST adds labels without replacing existing ones
curl -sf -X POST \
  -H "Authorization: token $CODEBERG_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"labels\":[$label_id]}" \
  "$api/repos/$repo/issues/$number/labels" > /dev/null

echo "#$number: added \"$label_name\""
