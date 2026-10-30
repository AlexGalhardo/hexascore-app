#!/usr/bin/env bash
# Fuzz one OpenAPI document with Schemathesis 4.29.0 (legacy pin).
# Usage: HX_TEST_KEY=hx_live_… [QA_BASE_URL=http://127.0.0.1:3000] qa/schemathesis/run.sh openapi/api-v1.json
# App tokens: HX_TEST_KEY=$HX_APP_TEST_TOKEN qa/schemathesis/run.sh openapi/app-v1.json
set -euo pipefail

doc=${1:-openapi/api-v1.json}
shift || true
: "${HX_TEST_KEY:?set HX_TEST_KEY}"
base=${QA_BASE_URL:-http://127.0.0.1:3000}

root=$(cd "$(dirname "$0")/../.." && (pwd -W 2>/dev/null || pwd))
mkdir -p "$root/qa/reports/schemathesis"
export MSYS_NO_PATHCONV=1

# Paths already include /api/v1 and /app/v1. Docker Desktop has no host network,
# so a loopback origin is reached as host.docker.internal.
url=${base%/}
case "$url" in
http://127.0.0.1:*|http://localhost:*)
	port=${url##*:}
	port=${port%%/*}
	url=http://host.docker.internal:${port}
	;;
esac

exec docker run --rm -e HOME=/tmp \
	-v "$root:/work" -w /work/qa/reports/schemathesis \
	schemathesis/schemathesis:4.29.0 \
	--config-file /work/qa/schemathesis/schemathesis.toml \
	run "/work/$doc" --url "$url" \
	-H "Authorization: Bearer $HX_TEST_KEY" \
	--checks all \
	--exclude-checks allow_header_conformance,unsupported_method \
	"$@"
# Those two checks stay off. --checks all would turn the toml flags back on.
# GET /{path...} is the HTML 404, so an undeclared GET under /app/v1 is 404
# rather than 405, and the 405 Allow list advertises that GET.
