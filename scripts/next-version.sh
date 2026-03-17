#!/usr/bin/env bash
# SPDX-License-Identifier: GPL-3.0-only
#
# Calculate the next semver tag based on issue labels since the last tag.
# Usage: next-version.sh <api> <repo> [--pre-release <stage>] [--refresh]
#
# Inspects git log since the last tag, extracts issue numbers from
# "fixes #N" / "closes #N", queries Codeberg for labels, and determines
# the bump level:
#
#   Compat/Breaking  -> major
#   Kind/Feature     -> minor
#   Kind/Enhancement -> minor
#   Kind/Bug         -> patch
#
# Options:
#   --pre-release <stage>   Append pre-release suffix (e.g. alpha, beta, rc).
#                           Auto-increments the numeric suffix if the same
#                           stage already exists for the computed version.
#   --refresh               Force refresh of cached issue data.
set -euo pipefail

api="$1"; repo="$2"; shift 2
dir="$(cd "$(dirname "$0")" && pwd)"

pre_stage=""
refresh_flag=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --pre-release) pre_stage="$2"; shift 2 ;;
    --refresh) refresh_flag="--refresh"; shift ;;
    *) echo "unknown option: $1" >&2; exit 1 ;;
  esac
done

# Find the last tag (stable or pre-release)
last_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "")

# Extract the stable base from the last tag (strip pre-release suffix)
if [[ -n "$last_tag" ]]; then
  base="${last_tag%%-*}"  # v0.1.0-alpha.1 -> v0.1.0
else
  base="v0.0.0"
fi

major="${base#v}"          # 0.1.0
major="${major%%.*}"       # 0
rest="${base#v"$major".}"  # 1.0
minor="${rest%%.*}"        # 1
patch="${rest#"$minor".}"  # 0

# Collect issue numbers from commits since last tag
if [[ -n "$last_tag" ]]; then
  range="${last_tag}..HEAD"
else
  range="HEAD"
fi

issues=$(git log "$range" --format='%s%n%b' \
  | grep -ioE '(fixes|closes|fix|close|resolves|resolve) #[0-9]+' \
  | grep -oE '[0-9]+' \
  | sort -u)

if [[ -z "$issues" ]]; then
  echo "no issues found since ${last_tag:-inception}" >&2
  bump="patch"
else
  bump="patch"
  summary=""

  for issue in $issues; do
    json=$("$dir/codeberg-fetch.sh" "$api/repos/$repo/issues/$issue" $refresh_flag)

    title=$(echo "$json" | jq -r '.title // "???"')
    labels=$(echo "$json" | jq -r '.labels[].name' 2>/dev/null || true)
    summary="${summary}  #${issue} ${title}\n"

    for label in $labels; do
      case "$label" in
        Compat/Breaking)
          bump="major"
          ;;
        Kind/Feature|Kind/Enhancement)
          [[ "$bump" != "major" ]] && bump="minor"
          ;;
      esac
    done
  done

  echo "issues since ${last_tag:-inception}:" >&2
  printf '%b' "$summary" >&2
  echo "bump: $bump" >&2
fi

# Apply bump
case "$bump" in
  major) major=$((major + 1)); minor=0; patch=0 ;;
  minor) minor=$((minor + 1)); patch=0 ;;
  patch) patch=$((patch + 1)) ;;
esac

next="v${major}.${minor}.${patch}"

# Handle pre-release suffix
if [[ -n "$pre_stage" ]]; then
  # Find existing tags for this version+stage to auto-increment
  existing=$(git tag -l "${next}-${pre_stage}.*" | sort -V | tail -1)
  if [[ -n "$existing" ]]; then
    last_num="${existing##*.}"
    next_num=$((last_num + 1))
  else
    next_num=1
  fi
  next="${next}-${pre_stage}.${next_num}"
fi

echo "$next"
