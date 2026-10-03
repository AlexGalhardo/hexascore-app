#!/usr/bin/env bash
# OWASP ZAP baseline (spider + passive) of the site and active scan of /api/v1 (Docker, pinned image).
# Usage: QA_BASE_URL=https://<deployment> [QA_DOCKER_BASE_URL=…] [QA_API_KEY=hx_live_…]
#        [VERCEL_AUTOMATION_BYPASS_SECRET=…] qa/zap/run.sh
# Without QA_API_KEY the API scan only sees the 401 surface. Reports: qa/reports/zap/zap-report.{html,json}.
set -euo pipefail

: "${QA_BASE_URL:?set QA_BASE_URL (deployment origin, e.g. http://localhost:3100)}"
# Docker Desktop has no host network: point the container at http://host.docker.internal:<port>.
target=${QA_DOCKER_BASE_URL:-$QA_BASE_URL}

export MSYS_NO_PATHCONV=1
root=$(cd "$(dirname "$0")/../.." && (pwd -W 2>/dev/null || pwd))
mkdir -p "$root/qa/reports/zap"
# The image runs as user `zap` (uid 1000), which must be able to write the reports.
chmod a+rwx "$root/qa/reports/zap"

exec docker run --rm --network host \
	-e TARGET_URL="${target%/}" -e QA_API_KEY="${QA_API_KEY:-}" \
	-e VERCEL_AUTOMATION_BYPASS_SECRET="${VERCEL_AUTOMATION_BYPASS_SECRET:-}" \
	-v "$root/qa/zap/plan.yaml:/zap/plan.yaml:ro" \
	-v "$root/apps/web/public/openapi.json:/zap/openapi.json:ro" \
	-v "$root/qa/reports/zap:/zap/wrk:rw" \
	ghcr.io/zaproxy/zaproxy:2.17.0 zap.sh -cmd -autorun /zap/plan.yaml
