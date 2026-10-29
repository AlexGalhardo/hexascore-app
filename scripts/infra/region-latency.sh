#!/usr/bin/env bash
# Usage: scripts/infra/region-latency.sh <url> [<url>...]
# Prints p50/p95 of TCP connect and time-to-first-byte (seconds) per URL, SAMPLES requests each (default 50).
set -euo pipefail

samples="${SAMPLES:-50}"
if [ "$#" -lt 1 ]; then
	echo "usage: $0 <url> [<url>...]" >&2
	exit 2
fi

# Nearest-rank percentile; reads one number per line on stdin.
percentile() {
	sort -g | awk -v p="$1" '{ a[NR] = $1 } END { if (NR == 0) { print "nan"; exit } i = int(p / 100 * NR + 0.999); if (i < 1) i = 1; print a[i] }'
}

printf '%-70s %8s %10s %10s %10s %10s %6s\n' url samples conn_p50 conn_p95 ttfb_p50 ttfb_p95 fails
for url in "$@"; do
	tmp="$(mktemp)"
	fails=0
	for _ in $(seq "$samples"); do
		# A failed request is counted, never recorded as a fast sample.
		if ! line="$(curl -s -o /dev/null -w '%{time_connect} %{time_starttransfer}' --max-time 10 "$url")"; then
			fails=$((fails + 1))
			continue
		fi
		echo "$line" >>"$tmp"
	done
	ok=$(wc -l <"$tmp" | tr -d ' ')
	printf '%-70s %8s %10s %10s %10s %10s %6s\n' "$url" "$ok" \
		"$(awk '{print $1}' "$tmp" | percentile 50)" "$(awk '{print $1}' "$tmp" | percentile 95)" \
		"$(awk '{print $2}' "$tmp" | percentile 50)" "$(awk '{print $2}' "$tmp" | percentile 95)" "$fails"
	rm -f "$tmp"
done
