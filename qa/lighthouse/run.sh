#!/usr/bin/env bash
# Lighthouse CI against a running app (SDD 09 §4). BASE_URL defaults to http://127.0.0.1:3000.
# The page list is qa/lighthouse/urls.txt (DemoCatalog: home, listing, game, book, person).
set -euo pipefail

base=${BASE_URL:-http://127.0.0.1:3000}
args=()
while IFS= read -r path; do
	[ -n "$path" ] && args+=("--collect.url=${base%/}${path}")
done < "$(dirname "$0")/urls.txt"

exec bunx --bun @lhci/cli@0.15.1 autorun "${args[@]}" "$@"
