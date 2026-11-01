#!/usr/bin/env bash
# Fuzz one OpenAPI document with Schemathesis 4.29.0 (legacy pin).
# Usage: HX_TEST_KEY=hx_live_… [QA_BASE_URL=http://127.0.0.1:3000/api/v1] qa/schemathesis/run.sh openapi/api-v1.json
# App tokens: HX_TEST_KEY=$HX_APP_TEST_TOKEN [QA_BASE_URL=http://127.0.0.1:3000] qa/schemathesis/run.sh openapi/app-v1.json
set -euo pipefail

doc=${1:-openapi/api-v1.json}
shift || true
: "${HX_TEST_KEY:?set HX_TEST_KEY}"
# api-v1 paths are relative to the API host's /v1 (servers url); on the app host that is /api/v1.
# app-v1 paths already start with /app/v1.
case "$doc" in
*api-v1.json) base=${QA_BASE_URL:-http://127.0.0.1:3000/api/v1} ;;
*) base=${QA_BASE_URL:-http://127.0.0.1:3000} ;;
esac

root=$(cd "$(dirname "$0")/../.." && (pwd -W 2>/dev/null || pwd))
mkdir -p "$root/qa/reports/schemathesis"
# The image runs as a non-root user. The crash cache is created in the workdir,
# and a 755 directory owned by the runner makes that mkdir fail after a real failure.
chmod 777 "$root/qa/reports/schemathesis"
export MSYS_NO_PATHCONV=1

# Docker Desktop has no host network, so a loopback origin is reached as host.docker.internal.
url=${base%/}
host_header=()
case "$url" in
http://127.0.0.1:*|http://localhost:*)
	hostport=${url#http://}
	hostport=${hostport%%/*}
	# Keep the Host the caller asked for: production CanonicalHost 301s any other
	# Host to APP_URL, which in the scan container is its own loopback (refused).
	host_header=(-H "Host: ${hostport}")
	url=http://host.docker.internal:${hostport##*:}${url#http://"$hostport"}
	;;
esac

# Distinct X-Forwarded-For per call, or the 120/min IP budget turns a missing
# Authorization into 429. The hook file is on the /work mount.
exec docker run --rm -e HOME=/tmp \
	-e SCHEMATHESIS_HOOKS=/work/qa/schemathesis/ip_hook.py \
	--add-host=host.docker.internal:host-gateway \
	-v "$root:/work" -w /work/qa/reports/schemathesis \
	schemathesis/schemathesis:4.29.0 \
	--config-file /work/qa/schemathesis/schemathesis.toml \
	run "/work/$doc" --url "$url" \
	-H "Authorization: Bearer $HX_TEST_KEY" \
	"${host_header[@]}" \
	--exclude-checks allow_header_conformance,unsupported_method \
	"$@"
# Do not pass --checks all. It turns the toml flags back on, including the
# per-route response_schema_conformance disables for empty 204 responses.
# Every other check stays on (Schemathesis default). allow_header_conformance
# and unsupported_method stay off: GET /{path...} is the HTML 404, so an
# undeclared GET under /app/v1 is 404 rather than 405, and the 405 Allow list
# advertises that GET.
