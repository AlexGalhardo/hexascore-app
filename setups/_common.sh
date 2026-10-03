# Shared helpers for setups/setup-unix-*.sh. Sourced, not executed.
# shellcheck shell=bash

EXPECTED_BUN="1.4.2"
LOCAL_DATABASE_URL="postgres://hexascore:hexascore@localhost:5432/hexascore"
LOCAL_UPSTASH_URL="http://localhost:8079"
LOCAL_UPSTASH_TOKEN="local-dev-token"

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="$ROOT_DIR/apps/web/.env.local"
ENV_EXAMPLE="$ROOT_DIR/.env.example"

if [ -t 1 ]; then
	C_BLUE=$'\033[34m' C_GREEN=$'\033[32m' C_YELLOW=$'\033[33m' C_RED=$'\033[31m' C_BOLD=$'\033[1m' C_RESET=$'\033[0m'
else
	C_BLUE="" C_GREEN="" C_YELLOW="" C_RED="" C_BOLD="" C_RESET=""
fi

step() { printf '\n%s==> %s%s\n' "$C_BLUE$C_BOLD" "$1" "$C_RESET"; }
ok() { printf '%s  ok%s %s\n' "$C_GREEN" "$C_RESET" "$1"; }
warn() { printf '%s  warn%s %s\n' "$C_YELLOW" "$C_RESET" "$1"; }
fail() {
	printf '%s  error%s %s\n' "$C_RED" "$C_RESET" "$1" >&2
	exit 1
}

check_bun() {
	step "Checking Bun"
	command -v bun >/dev/null 2>&1 || fail "Bun not found. Install $EXPECTED_BUN: curl -fsSL https://bun.sh/install | bash -s bun-v$EXPECTED_BUN"
	local version
	version="$(bun --version)"
	if [ "$version" = "$EXPECTED_BUN" ]; then
		ok "bun $version"
	else
		warn "bun $version found, project pins $EXPECTED_BUN (packageManager). Run: bun upgrade --version $EXPECTED_BUN"
	fi
}

install_deps() {
	step "Installing dependencies (bun install)"
	(cd "$ROOT_DIR" && bun install --frozen-lockfile) || fail "bun install failed (see output above)"
	ok "dependencies installed"
}

random_secret() {
	bun -e 'console.log(require("node:crypto").randomBytes(32).toString("hex"))'
}

# Prints the value of KEY in $ENV_FILE without surrounding quotes (empty if missing).
env_get() {
	awk -v key="$1" 'index($0, key "=") == 1 { v = substr($0, length(key) + 2); gsub(/^"|"$/, "", v); print v; exit }' "$ENV_FILE"
}

# Sets KEY="value" in $ENV_FILE (replacing or appending), writing through a temp file so a failure never leaves it half-written.
env_set() {
	local tmp
	tmp="$(mktemp "${ENV_FILE}.XXXXXX")"
	if awk -v key="$1" -v val="$2" '
		index($0, key "=") == 1 { print key "=\"" val "\""; done = 1; next }
		{ print }
		END { if (!done) print key "=\"" val "\"" }
	' "$ENV_FILE" >"$tmp"; then
		mv "$tmp" "$ENV_FILE"
	else
		rm -f "$tmp"
		fail "could not write $ENV_FILE"
	fi
}

# mode: "docker" (local Postgres + Upstash proxy) or "pglite" (no Docker).
prepare_env() {
	local mode="$1"
	step "Preparing apps/web/.env.local"
	if [ ! -f "$ENV_FILE" ]; then
		[ -f "$ENV_EXAMPLE" ] || fail ".env.example not found at $ENV_EXAMPLE"
		cp "$ENV_EXAMPLE" "$ENV_FILE"
		ok "created from .env.example"
	else
		ok "already exists, only filling empty/local values"
	fi

	local key
	for key in BETTER_AUTH_SECRET CRON_SECRET; do
		if [ -z "$(env_get "$key")" ]; then
			env_set "$key" "$(random_secret)"
			ok "generated $key"
		fi
	done

	local db_url upstash_url
	db_url="$(env_get DATABASE_URL)"
	upstash_url="$(env_get UPSTASH_REDIS_REST_URL)"
	if [ "$mode" = "docker" ]; then
		if [ -z "$db_url" ]; then
			env_set DATABASE_URL "$LOCAL_DATABASE_URL"
			ok "DATABASE_URL -> local Docker Postgres"
		elif [ "$db_url" != "$LOCAL_DATABASE_URL" ]; then
			warn "DATABASE_URL points elsewhere; keeping it (migrations will run there)"
		fi
		if [ -z "$upstash_url" ]; then
			env_set UPSTASH_REDIS_REST_URL "$LOCAL_UPSTASH_URL"
			env_set UPSTASH_REDIS_REST_TOKEN "$LOCAL_UPSTASH_TOKEN"
			ok "UPSTASH_REDIS_REST_* -> local proxy"
		fi
	else
		# Only undo values the Docker setup wrote; anything else was set on purpose by the developer.
		if [ "$db_url" = "$LOCAL_DATABASE_URL" ]; then
			env_set DATABASE_URL ""
			ok "DATABASE_URL cleared (was the local Docker URL) -> PGlite"
		elif [ -n "$db_url" ]; then
			warn "DATABASE_URL is set, so PGlite will NOT be used. Clear it in apps/web/.env.local for PGlite."
		fi
		if [ "$upstash_url" = "$LOCAL_UPSTASH_URL" ]; then
			env_set UPSTASH_REDIS_REST_URL ""
			env_set UPSTASH_REDIS_REST_TOKEN ""
			ok "UPSTASH_REDIS_REST_* cleared (local proxy needs Docker)"
		fi
	fi
}

migrate_and_seed() {
	step "Applying database migrations"
	(cd "$ROOT_DIR" && bun run db:migrate) || fail "migrations failed (see output above)"
	ok "migrations applied"

	step "Seeding"
	if [ -f "$ROOT_DIR/apps/web/lib/db/seed.ts" ]; then
		(cd "$ROOT_DIR" && bun run db:seed) || fail "seed failed (see output above)"
		ok "seed applied"
	else
		warn "no seed script yet (apps/web/lib/db/seed.ts), skipping"
	fi
}

print_next_steps() {
	step "Done"
	printf '  Start the app:  %sbun run dev%s\n' "$C_BOLD" "$C_RESET"
	printf '  Web:            http://localhost:3000\n'
	printf '  Env file:       apps/web/.env.local\n'
}
