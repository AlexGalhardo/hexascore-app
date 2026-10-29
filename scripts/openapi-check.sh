#!/usr/bin/env bash
# Exports both OpenAPI documents and diffs them against the committed goldens in openapi/.
# `--write` regenerates the goldens instead. APP_KEY is a throwaway value; the export never touches the database.
set -euo pipefail

cd "$(dirname "$0")/.."

export MSYS_NO_PATHCONV=1
export APP_KEY="base64:$(head -c 32 /dev/urandom | base64)"
export DATABASE_URL="${DATABASE_URL:-postgres://hexascore:hexascore@127.0.0.1:5433/hexascore?sslmode=disable}"
unset API_HOST


status=0
for name in api app; do
	out="$(mktemp)"
	go run ./cmd/app openapi:export -prefix "/$name/v1" >"$out"
	if [ "${1:-}" = "--write" ]; then
		mv "$out" "openapi/$name-v1.json"
		continue
	fi
	if ! diff -u "openapi/$name-v1.json" "$out"; then
		echo "openapi/$name-v1.json drifted: run scripts/openapi-check.sh --write and commit" >&2
		status=1
	fi
	rm -f "$out"
done
exit "$status"
