#!/usr/bin/env bash
# Load + quota/rate-limit test of the public API v1 with k6 (Docker, pinned image).
# Usage: QA_BASE_URL=https://<deployment> [QA_DOCKER_BASE_URL=…] QA_API_KEY=hx_live_… [QA_FREE_API_KEY=…]
#        [VERCEL_AUTOMATION_BYPASS_SECRET=…] [RATE=8] qa/k6/run.sh
# QA_API_KEY: Pro/Business key (~600 requests per run). QA_FREE_API_KEY: a Free key reserved for this test
# (its lifetime quota is spent on the first run; later runs only check the 402).
set -euo pipefail

: "${QA_BASE_URL:?set QA_BASE_URL (deployment origin, e.g. http://localhost:3100)}"
: "${QA_API_KEY:?set QA_API_KEY (API key of a paid plan)}"
# Docker Desktop has no host network: point the container at http://host.docker.internal:<port>.
docker_base_url=${QA_DOCKER_BASE_URL:-$QA_BASE_URL}

export MSYS_NO_PATHCONV=1
root=$(cd "$(dirname "$0")/../.." && (pwd -W 2>/dev/null || pwd))
mkdir -p "$root/qa/reports/k6"

exec docker run --rm --network host --user "$(id -u):$(id -g)" \
	-e BASE_URL="$docker_base_url" -e QA_API_KEY -e QA_FREE_API_KEY -e VERCEL_AUTOMATION_BYPASS_SECRET -e RATE \
	-v "$root/qa:/qa" grafana/k6:2.3.0 run --summary-export /qa/reports/k6/summary.json /qa/k6/api.js
