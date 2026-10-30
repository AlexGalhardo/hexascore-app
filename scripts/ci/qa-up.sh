#!/usr/bin/env bash
# Start the QA app from an image that is already loaded. Does not build.
# Usage: APP_KEY=base64:… scripts/ci/qa-up.sh [hexascore:ci]
set -euo pipefail

image=${1:-hexascore:ci}
root=$(cd "$(dirname "$0")/../.." && (pwd -W 2>/dev/null || pwd))
cd "$root"

tag=${E2E_IMAGE_TAG:-ci}
: "${APP_KEY:?set APP_KEY}"
export E2E_IMAGE_TAG="$tag"
export COMPOSE_PROJECT_NAME="${COMPOSE_PROJECT_NAME:-hx-ci}"

docker tag "$image" "hexascore-app:${tag}"
docker compose --project-name "$COMPOSE_PROJECT_NAME" --profile app up -d --wait --no-build app
