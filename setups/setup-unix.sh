#!/usr/bin/env bash
# Idempotent local setup: tools, .env, APP_KEY, local stack, JS deps, migrations.
set -euo pipefail

cd "$(dirname "$0")/.."

for tool in go bun docker vukem; do
	if ! command -v "$tool" >/dev/null 2>&1; then
		echo "setup: missing tool: $tool" >&2
		exit 1
	fi
done

if [ ! -f .env ]; then
	cp .env.example .env
fi

# vukem key:generate refuses to run without a fresh key, so only generate when APP_KEY is empty.
if ! grep -q '^APP_KEY=.' .env; then
	vukem key:generate
fi

docker compose up -d --wait
bun install --frozen-lockfile
vukem migrate
