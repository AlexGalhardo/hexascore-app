#!/usr/bin/env bash
# Runs every job of .github/workflows/ci.yml locally, in parallel, on HEAD.
# Playwright e2e is excluded (disabled by the owner on 2026-10-31).
#
# Usage: scripts/ci/local.sh [--only a,b] [--skip a,b] [--base REV] [--keep] [--tree]
#   --only/--skip  job names (see JOBS below)
#   --base REV     start of the commitlint range (default: last tag, else the root commit)
#   --tree         test the working tree as it is instead of a clean checkout of HEAD
#   --keep         keep the checkout, logs and QA databases after the run
# Logs: .ci-local/logs/<job>.log in the invoking checkout. Exit code 1 if any job fails.
#
# Needs: Docker with the compose Postgres on 127.0.0.1:5433 (docker compose up -d --wait),
# go, bun, golangci-lint (version pinned in ci.yml; falls back to go run), curl, openssl.
# Isolation (AGENTS.md "Local QA"): every job gets its own databases on that Postgres,
# its own host ports, containers, image tag and compose project (hexascore-qa-<id>).
# The shared compose project "hexascore" is never taken down.
set -uo pipefail

ROOT=$(cd "$(dirname "$0")/../.." && pwd)
cd "$ROOT"
export MSYS_NO_PATHCONV=1 MSYS2_ARG_CONV_EXCL='*'

JOBS=(lint test-go test-go-integration test-ts migrations vuln deps setup-windows build openapi lighthouse mobile)
only="" skip="" base="" keep=0 use_tree=0
while [ $# -gt 0 ]; do
	case "$1" in
	--only) only=${2:?}; shift 2 ;;
	--skip) skip=${2:?}; shift 2 ;;
	--base) base=${2:?}; shift 2 ;;
	--keep) keep=1; shift ;;
	--tree) use_tree=1; shift ;;
	*) sed -n '2,14p' "$0"; exit 2 ;;
	esac
done

W() { cygpath -m "$1" 2>/dev/null || echo "$1"; }
is_windows() { case "$(uname -s)" in MINGW* | MSYS* | CYGWIN*) return 0 ;; *) return 1 ;; esac; }
selected() {
	[ -z "$only" ] || case ",$only," in *",$1,"*) ;; *) return 1 ;; esac
	case ",$skip," in *",$1,"*) return 1 ;; esac
}

ID="$(date +%H%M%S)$$"
RUN=$(mktemp -d)
LOGS="$ROOT/.ci-local/logs"
mkdir -p "$LOGS"
rm -f "$LOGS"/*.log
SHA=$(git rev-parse --short HEAD)
PGC=${LOCAL_CI_PG_CONTAINER:-hexascore-postgres-1}
PGURL=${LOCAL_CI_PG_URL:-postgres://hexascore:hexascore@127.0.0.1:5433}
PGURL_CONTAINERS=${PGURL//127.0.0.1/host.docker.internal}
IMG="hexascore:ci-local-$ID"
QA_TAG="cilocal$ID"
TREES=() DBS=() PORTS=() CONTAINERS=() PROJECTS=()

log() { printf '[ci-local] %s\n' "$*" >&2; }

pgsql() { docker exec "$PGC" psql -U hexascore -d postgres -v ON_ERROR_STOP=1 -qc "$1" >/dev/null; }
newdb() { local n="hx_ci_${ID}_$1"; pgsql "CREATE DATABASE $n"; DBS+=("$n"); echo "$n"; }
fresh_tree() {
	local t
	t=$(W "$(mktemp -d)")
	git worktree add -q --detach "$t" HEAD
	TREES+=("$t")
	echo "$t"
}
free_port() {
	local p
	while :; do
		p=$((3100 + RANDOM % 800))
		case " ${PORTS[*]:-} " in *" $p "*) continue ;; esac
		(exec 3<>/dev/tcp/127.0.0.1/"$p") 2>/dev/null && continue
		PORTS+=("$p")
		echo "$p"
		return
	done
}

cleanup() {
	trap - EXIT INT TERM
	jobs -p | xargs -r kill 2>/dev/null
	local c p d t
	for c in "${CONTAINERS[@]:-}"; do [ -n "$c" ] && docker rm -f "$c" >/dev/null 2>&1; done
	for p in "${PROJECTS[@]:-}"; do [ -n "$p" ] && docker compose -p "$p" down -v --remove-orphans >/dev/null 2>&1; done
	if [ "$keep" -eq 0 ]; then
		for d in "${DBS[@]:-}"; do [ -n "$d" ] && pgsql "DROP DATABASE IF EXISTS $d WITH (FORCE)" 2>/dev/null; done
		docker rmi "$IMG" "hexascore-app:ci$QA_TAG" >/dev/null 2>&1
		for t in "${TREES[@]:-}"; do [ -n "$t" ] && { git worktree remove --force "$t" >/dev/null 2>&1; rm -rf "$t"; }; done
		rm -rf "$RUN"
	else
		log "kept: trees ${TREES[*]:-} databases ${DBS[*]:-}"
	fi
}
trap cleanup EXIT INT TERM

# Preconditions ----------------------------------------------------------------
if ! docker exec "$PGC" pg_isready -U hexascore >/dev/null 2>&1; then
	log "Postgres container $PGC not ready; starting the shared compose Postgres"
	docker compose up -d --wait postgres >/dev/null || { log "no Postgres on 127.0.0.1:5433"; exit 2; }
fi
if [ "$use_tree" -eq 1 ]; then
	TREE=$ROOT
else
	[ -z "$(git status --porcelain)" ] || log "WARNING: uncommitted changes are NOT tested (HEAD $SHA only); use --tree to test them"
	TREE=$(fresh_tree)
fi
TREE_W=$(W "$TREE")
WF="$TREE/.github/workflows/ci.yml"
pin() { sed -n "s/^  $1: //p" "$WF" | head -1; }
GOLANGCI_LINT_VERSION=$(pin GOLANGCI_LINT_VERSION)
MARKDOWNLINT_VERSION=$(pin MARKDOWNLINT_VERSION)
OSV_SCANNER_VERSION=$(pin OSV_SCANNER_VERSION)
OSV_SCANNER_SHA256=$(pin OSV_SCANNER_SHA256)
PG_IMAGE=$(sed -n 's/.*image: \(postgres:[^ ]*\).*/\1/p' "$WF" | head -1)
for v in GOLANGCI_LINT_VERSION MARKDOWNLINT_VERSION OSV_SCANNER_VERSION OSV_SCANNER_SHA256 PG_IMAGE; do
	[ -n "${!v}" ] || { log "could not read $v from $WF"; exit 2; }
done
export PG_IMAGE
want_bun=$(sed -n 's/.*"packageManager": "bun@\([^"]*\)".*/\1/p' "$TREE/package.json")
[ "$(bun --version)" = "$want_bun" ] || log "WARNING: bun $(bun --version) installed, package.json pins $want_bun"
if [ -z "$base" ]; then
	base=$(git -C "$TREE" describe --tags --abbrev=0 2>/dev/null || git -C "$TREE" rev-list --max-parents=0 HEAD | tail -1)
fi

# golangci-lint: the pinned version, installed or through go run -------------------
if golangci-lint version 2>/dev/null | grep -q "version ${GOLANGCI_LINT_VERSION#v} "; then
	GL=(golangci-lint)
else
	GL=(go run "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$GOLANGCI_LINT_VERSION")
fi
# golangci-lint refuses to run twice at once (exit 3); another worker may hold the lock.
gl() {
	local out rc n=0
	out=$(mktemp)
	while :; do
		# && / || keeps set -e from leaving the function before the retry check.
		"${GL[@]}" "$@" >"$out" 2>&1 && rc=0 || rc=$?
		cat "$out"
		if [ "$rc" -eq 3 ] && grep -qi 'parallel golangci-lint' "$out" && [ "$n" -lt 90 ]; then
			n=$((n + 1))
			echo "golangci-lint lock held by another run, retry $n/90 in 10 s"
			sleep 10
			continue
		fi
		rm -f "$out"
		return "$rc"
	done
}

step() { printf '\n==> %s\n' "$*"; }
need_build() {
	step "waiting for the build job"
	while [ ! -f "$RUN/build.rc" ]; do sleep 2; done
	[ "$(cut -d' ' -f1 "$RUN/build.rc")" -eq 0 ] || { echo "build failed, dependent job not run"; return 1; }
}
pg_shims() {
	local d="$RUN/pgbin" t
	mkdir -p "$d"
	for t in pg_dump pg_restore; do
		if is_windows; then
			printf '@echo off\r\ndocker run --rm -i --network host -e PGHOST -e PGPORT -e PGUSER -e PGPASSWORD -e PGDATABASE -e PGSSLMODE %s %s %%*\r\n' "$PG_IMAGE" "$t" >"$d/$t.cmd"
		else
			printf '#!/bin/sh\nexec docker run --rm -i --network host -e PGHOST -e PGPORT -e PGUSER -e PGPASSWORD -e PGDATABASE -e PGSSLMODE %s %s "$@"\n' "$PG_IMAGE" "$t" >"$d/$t"
			chmod +x "$d/$t"
		fi
	done
	echo "$d"
}
app_key() { echo "base64:$(openssl rand -base64 32)"; }

# Jobs (one function per ci.yml job) -------------------------------------------
job_lint() {
	cd "$TREE"
	local db
	db=$(newdb lint)
	step golangci-lint run; gl run --build-tags integration ./...
	step golangci-lint fmt --diff; gl fmt --diff
	step go vet; go vet ./...
	step "OpenAPI goldens"; DATABASE_URL="$PGURL/$db?sslmode=disable" bash scripts/openapi-check.sh
	step "OpenAPI lint"; bun run openapi:lint
	step "bun run lint"; bun run lint
	step "bun run typecheck"; bun run typecheck
	step markdownlint; bunx "markdownlint-cli2@$MARKDOWNLINT_VERSION" "**/*.md"
	step "commitlint $base..HEAD"; bunx commitlint --from "$base" --to HEAD
}
job_test_go() {
	cd "$TREE"
	step "go test -race"; TEST_DATABASE_URL="$PGURL/postgres?sslmode=disable" go test -race -count=1 ./...
}
job_test_go_integration() {
	cd "$TREE"
	# bun is on PATH so TestPluralGolden compares Go and TypeScript output; pg_dump/pg_restore 18 come from the CI image.
	step "go test -race -tags integration"
	PATH="$(pg_shims):$PATH" TEST_DATABASE_URL="$PGURL/postgres?sslmode=disable" go test -race -count=1 -tags integration ./...
}
job_test_ts() {
	cd "$TREE"
	step "bun run test"; bun run test
	step "migration policy"; bun test scripts/ci/migration-policy.test.ts
	step "TypeScript SDK"; (cd sdks/ts && bun install --frozen-lockfile --ignore-scripts && bun run typecheck && bun test)
}
job_migrations() {
	cd "$TREE"
	local db
	db=$(newdb mig)
	export APP_ENV=development DATABASE_URL="$PGURL/$db?sslmode=disable" SESSION_DRIVER=cookie MAIL_MAILER=log APP_KEY
	APP_KEY=$(app_key)
	step migrate; go run ./cmd/app migrate
	step "rollback --step 1"; go run ./cmd/app migrate:rollback --step 1
	step migrate; go run ./cmd/app migrate
}
job_vuln() {
	cd "$TREE"
	step govulncheck; go tool govulncheck ./...
}
job_deps() {
	cd "$TREE"
	local cache=${LOCAL_CI_CACHE:-$HOME/.cache/hexascore-ci} bin
	mkdir -p "$cache"
	bin="$cache/osv-scanner-$OSV_SCANNER_VERSION"
	if ! echo "$OSV_SCANNER_SHA256  $bin" | sha256sum -c - >/dev/null 2>&1; then
		step "download osv-scanner $OSV_SCANNER_VERSION"
		curl -fsSL --retry 3 -o "$(W "$bin")" "https://github.com/google/osv-scanner/releases/download/${OSV_SCANNER_VERSION}/osv-scanner_linux_amd64"
		echo "$OSV_SCANNER_SHA256  $bin" | sha256sum -c -
		chmod +x "$bin"
	fi
	# The pinned release is a static linux/amd64 binary; run it in the Postgres image already used by CI so it works on any host.
	step osv-scanner
	docker run --rm -v "$(W "$TREE"):/src:ro" -v "$(W "$bin"):/osv-scanner:ro" -w /src "$PG_IMAGE" \
		/osv-scanner scan source --config=osv-scanner.toml -L go.mod -L bun.lock -L sdks/ts/bun.lock
	step "bun audit"
	bun audit --audit-level=high --ignore GHSA-mrhx-c7xc-m3w5 --ignore GHSA-gcqx-w48m-w2h4 --ignore GHSA-3g3m-2fjr-6pw4 --ignore GHSA-g3hv-75wg-m249
}
job_setup_windows() {
	is_windows || { echo "windows-only job (CI runs it on windows-2025); skipped on $(uname -s)"; return 77; }
	local t
	t=$(fresh_tree)
	cd "$t"
	step "install the vukem CLI (version from go.mod)"
	go install "github.com/AlexGalhardo/vukem/cmd/vukem@$(go list -m -f '{{.Version}}' github.com/AlexGalhardo/vukem)"
	step "setup-windows.ps1 -NoServices"
	powershell.exe -NoProfile -ExecutionPolicy Bypass -File ./setups/setup-windows.ps1 -NoServices
	step "go test"; go test -count=1 ./...
}
job_build() {
	cd "$TREE"
	step "docker build"; docker build --build-arg "APP_VERSION=$(git -C "$TREE" rev-parse HEAD)" -t "$IMG" .
	step "image checks"; scripts/infra/check-image.sh "$IMG"
	step "/up smoke test"
	local port name=hx-ci-$ID-smoke
	port=$(free_port)
	CONTAINERS+=("$name")
	docker run -d --name "$name" -p "$port:3000" \
		-e APP_KEY="$(app_key)" -e APP_URL="http://localhost:$port" \
		-e STRIPE_KEY=sk_test_placeholder_ci -e STRIPE_WEBHOOK_SECRET=whsec_placeholder_ci \
		-e STRIPE_PRICE_PRO=price_placeholder_pro -e STRIPE_PRICE_BUSINESS=price_placeholder_business "$IMG" >/dev/null
	local i
	for i in $(seq 1 30); do curl -fsS "localhost:$port/up" >/dev/null && break; sleep 1; done
	# No database here, so readiness is 503 on "db"; only the ssr check matters.
	if ! curl -sS "localhost:$port/up?checks=1" | grep -q '"ssr":"ok"'; then docker logs "$name"; docker rm -f "$name" >/dev/null; return 1; fi
	docker rm -f "$name" >/dev/null
}
job_openapi() {
	need_build
	cd "$TREE"
	local db
	db=$(newdb oa)
	export DATABASE_URL="$PGURL/$db?sslmode=disable"
	step "export matches the committed documents"; bash scripts/openapi-check.sh --image "$IMG"
	step Redocly; bunx redocly lint openapi/api-v1.json openapi/app-v1.json
	step "app client matches codegen"
	local out
	out=$(W "$(mktemp)")
	docker run --rm --network host -e APP_ENV=development -e "APP_KEY=$(app_key)" -e DATABASE_URL -e API_HOST- \
		"$IMG" codegen:client -prefix /app/v1 -o - >"$out"
	git diff --no-index --exit-code packages/app-client/src/api.ts "$out" || { rm -f "$out"; return 1; }
	rm -f "$out"
	bun run --filter @hexascore/app-client typecheck test
	step "Schemathesis against the image"
	local sdb
	sdb=$(newdb st)
	DATABASE_URL="$PGURL/$sdb?sslmode=disable" QA_PORT=$(free_port) bash qa/schemathesis/ci.sh "$IMG"
}
chrome_path() {
	local c
	for c in "${CHROME_PATH:-}" "/c/Program Files/Google/Chrome/Application/chrome.exe" "/c/Program Files (x86)/Google/Chrome/Application/chrome.exe" \
		"$LOCALAPPDATA/Google/Chrome/Application/chrome.exe" "$(command -v google-chrome chromium chromium-browser 2>/dev/null | head -1)"; do
		[ -n "$c" ] && [ -x "$c" ] && { echo "$c"; return; }
	done
	return 1
}
job_lighthouse() {
	need_build
	cd "$TREE"
	local port db proj=hexascore-qa-$QA_TAG chrome
	chrome=$(chrome_path) || { echo "no Chrome found; set CHROME_PATH"; return 1; }
	port=$(free_port)
	db=$(newdb qa)
	PROJECTS+=("$proj")
	export APP_KEY E2E_IMAGE_TAG="ci$QA_TAG" QA_PORT="$port" QA_DATABASE_URL="$PGURL_CONTAINERS/$db?sslmode=disable"
	APP_KEY=$(app_key)
	step "QA stack on :$port ($proj)"
	docker tag "$IMG" "hexascore-app:$E2E_IMAGE_TAG"
	local c=(docker compose -p "$proj" --profile app)
	"${c[@]}" run --rm --no-deps app-init
	"${c[@]}" up -d --wait --no-deps --no-build app
	step Lighthouse
	E2E_BASE_URL="http://127.0.0.1:$port" CHROME_PATH="$(W "$chrome")" bunx @lhci/cli@0.15.1 autorun
}
job_mobile() {
	cd "$TREE"
	step lint; bun run --cwd apps/mobile lint
	step typecheck; bun run --cwd apps/mobile typecheck
	step test; bun run --cwd apps/mobile test
}

# Run -------------------------------------------------------------------------
log "HEAD $SHA, tree $TREE_W, commitlint base $base, logs $LOGS"
step_prep() { (cd "$TREE" && bun install --frozen-lockfile --ignore-scripts); }
step_prep >"$LOGS/prepare.log" 2>&1 || { log "bun install failed, see $LOGS/prepare.log"; exit 1; }

started=()
launch() {
	local j
	for j in "$@"; do
		selected "$j" || continue
		started+=("$j")
		(
			t0=$(date +%s)
			( set -eo pipefail; "job_${j//-/_}" ) >"$LOGS/$j.log" 2>&1
			echo "$? $(($(date +%s) - t0))" >"$RUN/$j.rc"
		) &
	done
}
# Two waves: the JS test runners (vitest forks, jest 5 s timeouts) fail under the CPU load of
# go -race, golangci-lint and the Docker build, so those start once they are done.
launch test-ts mobile vuln deps migrations
for j in test-ts mobile; do
	selected "$j" && while [ ! -f "$RUN/$j.rc" ]; do sleep 2; done
done
launch build lint test-go test-go-integration setup-windows openapi lighthouse
log "running ${started[*]} (e2e excluded: owner, 2026-10-31)"
wait

printf '\n%-20s %-6s %6s  %s\n' JOB RESULT SECS LOG
fail=0
for j in "${JOBS[@]}"; do
	if ! selected "$j"; then continue; fi
	read -r rc secs <"$RUN/$j.rc"
	case $rc in 0) res=PASS ;; 77) res=SKIP ;; *) res=FAIL; fail=1 ;; esac
	printf '%-20s %-6s %6s  %s\n' "$j" "$res" "$secs" ".ci-local/logs/$j.log"
done
printf '%-20s %-6s %6s  %s\n' e2e SKIP - "disabled by the owner (RUN_E2E)"
for j in "${started[@]}"; do
	read -r rc _ <"$RUN/$j.rc"
	if [ "$rc" -ne 0 ] && [ "$rc" -ne 77 ]; then printf '\n--- %s (last 25 lines) ---\n' "$j"; tail -n 25 "$LOGS/$j.log"; fi
done
[ "$fail" -eq 0 ] && echo "ALL GREEN on $SHA" || echo "FAILED on $SHA"
exit "$fail"
