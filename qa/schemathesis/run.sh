#!/usr/bin/env bash
# Fuzzes the public API v1 from apps/web/public/openapi.json with Schemathesis (Docker, pinned image).
# Usage: QA_BASE_URL=https://<deployment> [QA_DOCKER_BASE_URL=…] QA_API_KEY=hx_live_… [VERCEL_AUTOMATION_BYPASS_SECRET=…] qa/schemathesis/run.sh [extra st run args]
# QA_API_KEY must belong to a paid plan (Free = 100 requests for life and 1 req/s); a run spends ~2-3k requests.
set -euo pipefail

: "${QA_BASE_URL:?set QA_BASE_URL (deployment origin, e.g. http://localhost:3100)}"
# Docker Desktop has no host network: point the container at http://host.docker.internal:<port>.
docker_base_url=${QA_DOCKER_BASE_URL:-$QA_BASE_URL}
: "${QA_API_KEY:?set QA_API_KEY (API key of a paid plan)}"

export MSYS_NO_PATHCONV=1
root=$(cd "$(dirname "$0")/../.." && (pwd -W 2>/dev/null || pwd))
mkdir -p "$root/qa/reports/schemathesis"

headers=(-H "Authorization: Bearer $QA_API_KEY")
if [ -n "${VERCEL_AUTOMATION_BYPASS_SECRET:-}" ]; then
	headers+=(-H "x-vercel-protection-bypass: $VERCEL_AUTOMATION_BYPASS_SECRET")
fi

# Fail fast on a wrong key or a Free key instead of fuzzing 401s/402s.
usage=$(curl -fsS --max-time 15 "${QA_BASE_URL%/}/api/v1/me/usage" "${headers[@]}") || {
	echo "QA_API_KEY rejected by ${QA_BASE_URL}/api/v1/me/usage" >&2
	exit 1
}
case "$usage" in *'"plan":"free"'*) echo "QA_API_KEY is on the Free plan; use a Pro or Business key" >&2 && exit 1 ;; esac

exec docker run --rm --network host --user "$(id -u):$(id -g)" -e HOME=/tmp \
	-v "$root:/work" -w /work/qa/reports/schemathesis \
	schemathesis/schemathesis:4.29.0 --config-file /work/qa/schemathesis/schemathesis.toml \
	run /work/apps/web/public/openapi.json --url "${docker_base_url%/}/api/v1" "${headers[@]}" "$@"
