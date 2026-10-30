#!/usr/bin/env bash
# Exports both OpenAPI documents and diffs them against the committed goldens in openapi/.
# Usage: scripts/openapi-check.sh [--image hexascore:ci] [--write]
# --image runs openapi:export inside that image (the CI job, no rebuild).
# --write regenerates the goldens. The export never touches the database.
set -euo pipefail

cd "$(dirname "$0")/.."

image=
write=0
while [ $# -gt 0 ]; do
	case "$1" in
	--image)
		image=${2:?image name}
		shift 2
		;;
	--write)
		write=1
		shift
		;;
	*)
		echo "usage: scripts/openapi-check.sh [--image NAME] [--write]" >&2
		exit 2
		;;
	esac
done

export MSYS_NO_PATHCONV=1
export APP_KEY="${APP_KEY:-base64:$(head -c 32 /dev/urandom | base64)}"
export DATABASE_URL="${DATABASE_URL:-postgres://hexascore:hexascore@127.0.0.1:5433/hexascore?sslmode=disable}"
unset API_HOST

export_doc() {
	local prefix=$1
	if [ -n "$image" ]; then
		# Warmup command: no database. APP_ENV=local skips the production Stripe check.
		docker run --rm --network none \
			-e APP_ENV=local \
			-e APP_KEY="$APP_KEY" \
			-e API_HOST- \
			"$image" openapi:export -prefix "$prefix"
		return
	fi
	go run ./cmd/app openapi:export -prefix "$prefix"
}

status=0
for name in api app; do
	# Relative path: Windows git cannot open a Git-Bash /tmp file when MSYS_NO_PATHCONV is set.
	out="openapi/.export-${name}.json"
	export_doc "/$name/v1" >"$out"
	if [ "$write" -eq 1 ]; then
		mv "$out" "openapi/$name-v1.json"
		continue
	fi
	# A handler change that is not regenerated makes this exit 1.
	if ! git diff --no-index --exit-code "openapi/$name-v1.json" "$out"; then
		echo "openapi/$name-v1.json drifted: run scripts/openapi-check.sh --write and commit" >&2
		status=1
	fi
	rm -f "$out"
done
exit "$status"
