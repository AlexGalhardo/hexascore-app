#!/usr/bin/env bash
# Local setup with Docker: Postgres 18 + Redis + Upstash REST proxy (infra/docker-compose.yaml). Idempotent.
set -euo pipefail
# shellcheck source=setups/_common.sh
source "$(dirname "${BASH_SOURCE[0]}")/_common.sh"

COMPOSE=(docker compose -f "$ROOT_DIR/infra/docker-compose.yaml")

check_bun

step "Checking Docker"
command -v docker >/dev/null 2>&1 || fail "Docker not found. Install Docker Desktop/Engine, or use setups/setup-unix-pglite.sh"
docker info >/dev/null 2>&1 || fail "Docker daemon is not running. Start Docker and run this script again."
docker compose version >/dev/null 2>&1 || fail "Docker Compose v2 plugin not found ('docker compose')."
ok "docker $(docker version --format '{{.Server.Version}}')"

install_deps
prepare_env docker

step "Starting containers (Postgres, Redis, Upstash REST proxy)"
"${COMPOSE[@]}" up -d --wait --wait-timeout 120 ||
	fail "containers did not become healthy. Inspect: docker compose -f infra/docker-compose.yaml logs (is port 5432/6379/8079 already in use?)"
ok "postgres, redis and Upstash REST proxy healthy"

migrate_and_seed
print_next_steps
printf '  Postgres:       %s\n' "$LOCAL_DATABASE_URL"
printf '  Upstash REST:   %s (token %s)\n' "$LOCAL_UPSTASH_URL" "$LOCAL_UPSTASH_TOKEN"
printf '  Stop:           docker compose -f infra/docker-compose.yaml down   (add -v to drop data)\n'
