#!/usr/bin/env bash
# Checks log lines on stdin against SDD 12 FR-12-1/FR-12-8: every line is JSON
# with `level`; request lines also carry `request_id` and `route`; no line
# contains an email address. Needs jq.
#
#   railway logs --service app | scripts/ops/check-logs.sh
set -euo pipefail

bad=0 emails=0 n=0
while IFS= read -r line; do
	[[ -z $line ]] && continue
	n=$((n + 1))
	if ! jq -e 'has("level") and (.msg != "request" or (has("request_id") and has("route")))' <<<"$line" >/dev/null 2>&1; then
		bad=$((bad + 1))
		echo "bad line: ${line:0:200}" >&2
	fi
	if grep -qE '[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}' <<<"$line"; then
		emails=$((emails + 1))
	fi
done

echo "lines=$n bad=$bad email_matches=$emails"
[[ $bad == 0 && $emails == 0 ]]
