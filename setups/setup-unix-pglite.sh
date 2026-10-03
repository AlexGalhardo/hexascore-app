#!/usr/bin/env bash
# Local setup without Docker: empty DATABASE_URL -> in-process PGlite at apps/web/.pglite. Idempotent.
set -euo pipefail
# shellcheck source=setups/_common.sh
source "$(dirname "${BASH_SOURCE[0]}")/_common.sh"

check_bun
install_deps
prepare_env pglite
migrate_and_seed
print_next_steps
printf '  Database:       PGlite at apps/web/.pglite (delete the folder to reset)\n'
printf '  Redis:          disabled (optional integration; set UPSTASH_REDIS_REST_* to enable)\n'
