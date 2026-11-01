#!/usr/bin/env bash
# Lighthouse CI against a running app (SDD 09 §4).
# The five URLs live in lighthouserc.json (same list as urls.txt). Default host is
# http://127.0.0.1:3000. Set BASE_URL to point the same paths at another origin.
set -euo pipefail

base=${BASE_URL:-http://127.0.0.1:3000}
args=()
if [ "$base" != "http://127.0.0.1:3000" ]; then
	while IFS= read -r path; do
		[ -n "$path" ] && args+=("--collect.url=${base%/}${path}")
	done < "$(dirname "$0")/urls.txt"
fi

exec bunx --bun @lhci/cli@0.15.1 autorun "${args[@]}" "$@"
