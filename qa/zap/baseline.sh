#!/usr/bin/env bash
# OWASP ZAP baseline (spider + passive). Image tag 2.17.0, pinned by digest.
# Usage: qa/zap/baseline.sh http://127.0.0.1:3000
# Exit 0 when no High (FAIL) alerts remain. Mediums are triaged in docs/qa/scans.md.
set -euo pipefail

web=${1:?usage: qa/zap/baseline.sh BASE_URL}
root=$(cd "$(dirname "$0")/../.." && (pwd -W 2>/dev/null || pwd))
mkdir -p "$root/qa/reports/zap"
chmod a+rwx "$root/qa/reports/zap" || true
export MSYS_NO_PATHCONV=1

# ghcr.io/zaproxy/zaproxy:2.17.0 index, inspected 2026-10-30.
# https://github.com/zaproxy/zaproxy/releases/tag/v2.17.0
image=ghcr.io/zaproxy/zaproxy@sha256:030d3df57735dacbb83464854758e674dd9ad01d79d12a129a8e58b1fd384238

# Docker Desktop has no host network. Rewrite a loopback origin for the container.
target=$web
case "$target" in
http://127.0.0.1:*|http://localhost:*)
	port=${target##*:}
	port=${port%%/*}
	target=http://host.docker.internal:${port}
	;;
esac

exec docker run --rm \
	-v "$root/qa/zap/rules.tsv:/zap/wrk/rules.tsv:ro" \
	-v "$root/qa/reports/zap:/zap/wrk:rw" \
	"$image" \
	zap-baseline.py -t "$target" -I \
	-c /zap/wrk/rules.tsv \
	-J zap-report.json -r zap-report.html \
	-m 5
