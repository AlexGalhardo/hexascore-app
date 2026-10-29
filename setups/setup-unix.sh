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

# Pinned to the version in .github/workflows/ci.yml (GOLANGCI_LINT_VERSION). Install docs: https://golangci-lint.run/docs/welcome/install/
GOLANGCI_LINT_VERSION=2.15.0

golangci_lint_version() {
	golangci-lint version 2>/dev/null | sed -n 's/.*has version \([0-9.]*\).*/\1/p'
}

if [ "$(golangci_lint_version)" != "$GOLANGCI_LINT_VERSION" ]; then
	go install "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v${GOLANGCI_LINT_VERSION}"
fi
if [ "$(golangci_lint_version)" != "$GOLANGCI_LINT_VERSION" ]; then
	echo "setup: golangci-lint is not ${GOLANGCI_LINT_VERSION}; make sure $(go env GOPATH)/bin is on PATH" >&2
	exit 1
fi

docker compose up -d --wait
bun install --frozen-lockfile
vukem migrate
