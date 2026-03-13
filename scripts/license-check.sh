#!/usr/bin/env bash
set -euo pipefail

header="$1"
missing=()
wrong=()
stray=()

while IFS= read -r f; do
  first=$(head -1 "$f")
  if [[ "$first" != "$header" ]]; then
    if grep -q 'SPDX-License-Identifier' "$f"; then
      wrong+=("$f  (line 1: $first)")
    else
      missing+=("$f")
    fi
  fi
  count=$(grep -c 'SPDX-License-Identifier' "$f")
  if [[ "$count" -gt 1 ]]; then
    stray+=("$f  (${count} occurrences)")
  fi
done < <(find . -name '*.go' -not -path './vendor/*')

ok=true
if [[ ${#missing[@]} -gt 0 ]]; then
  echo "✗ Missing SPDX header:"
  printf '  %s\n' "${missing[@]}"
  ok=false
fi
if [[ ${#wrong[@]} -gt 0 ]]; then
  echo "✗ SPDX header present but not on line 1:"
  printf '  %s\n' "${wrong[@]}"
  ok=false
fi
if [[ ${#stray[@]} -gt 0 ]]; then
  echo "✗ Duplicate SPDX headers:"
  printf '  %s\n' "${stray[@]}"
  ok=false
fi
if [[ "$ok" == true ]]; then
  n=$(find . -name '*.go' -not -path './vendor/*' | wc -l)
  echo "✓ All ${n// /} files have correct SPDX headers"
fi
[[ "$ok" == true ]]
