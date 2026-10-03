#!/usr/bin/env sh
# Local stand-in for the GitHub Actions CI (the workflows stay in .github/ but are disabled on GitHub while the
# Actions billing is blocked). Runs from the husky pre-push hook; `bun run ci:local` runs it by hand.
# Mirrors .github/workflows/ci.yml (lint, typecheck, test, openapi check, build) and python-ci.yml (when
# webscrappings/ changed). Escape hatch for an emergency push: CI_LOCAL_SKIP=1 git push (never routine).
set -eu

if [ "${CI_LOCAL_SKIP:-}" = "1" ]; then
	echo "[ci:local] skipped (CI_LOCAL_SKIP=1)"
	exit 0
fi

step() { printf '\n[ci:local] %s\n' "$1"; }
started=$(date +%s)

step "lint (biome ci)"
bun run lint:ci

step "typecheck"
bun run typecheck

step "test"
bun run test

step "openapi check"
bun run openapi:check

# Same dummy-but-valid env as the CI build job (apps/web/lib/env.ts validates it with zod).
step "build"
export NEXT_PUBLIC_APP_NAME=HexaScore NEXT_PUBLIC_APP_URL=http://localhost:3000 BETTER_AUTH_URL=http://localhost:3000
export BETTER_AUTH_SECRET=ci-only-better-auth-secret-0123456789abcdef CRON_SECRET=ci-only-cron-secret-0123456789abcdef0123
if ! bun run build; then
	# `bun --bun next build` can crash its workers on Windows (exit 3/9) even on a clean tree; retry under Node.
	echo "[ci:local] bun build failed, retrying with Node (bunx next build)"
	(cd apps/web && bunx next build)
fi

base=$(git rev-parse --verify -q origin/dev || git rev-parse --verify -q origin/main || echo "")
if [ -n "$base" ] && git diff --name-only "$base"...HEAD | grep -q '^webscrappings/'; then
	if command -v uv >/dev/null 2>&1; then
		step "python (ruff + pytest)"
		(cd webscrappings && uv run ruff check . && uv run ruff format --check . && uv run pytest -q)
	else
		echo "[ci:local] webscrappings/ changed but uv is not installed: python checks skipped"
	fi
fi

printf '\n[ci:local] all checks passed in %ss\n' "$(($(date +%s) - started))"
