#!/usr/bin/env bash
# SPDX-License-Identifier: GPL-3.0-only
#
# Set the milestone on an issue.
# Usage: codeberg-set-milestone.sh <api> <repo> <issue_number> <milestone_id>
set -euo pipefail
: "${CODEBERG_TOKEN:?Set CODEBERG_TOKEN to a Codeberg personal access token}"

api="$1"; repo="$2"; issue="$3"; milestone="$4"

curl -sf \
  -H "Authorization: token $CODEBERG_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"milestone\": $milestone}" \
  -X PATCH \
  "$api/repos/$repo/issues/$issue" > /dev/null

echo "#$issue → milestone $milestone"
