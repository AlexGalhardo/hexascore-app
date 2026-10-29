#!/usr/bin/env bash
# Usage: scripts/infra/check-image.sh <image>
# Production image checks (P2.A.3.1, SDD 11 §6.3): compressed size, non-root
# user, pinned base images, the tools the jobs need, and the writable paths.
set -euo pipefail

image="${1:?usage: scripts/infra/check-image.sh <image>}"
max_mb=400
fail() { echo "check-image: $*" >&2; exit 1; }

user=$(docker image inspect "$image" --format '{{.Config.User}}')
[ "$user" = "app" ] || fail "user is '$user', want app"

size=$(docker save "$image" | gzip -c | wc -c)
mb=$((size / 1024 / 1024))
[ "$mb" -le "$max_mb" ] || fail "compressed size ${mb} MB > ${max_mb} MB"

if grep -nE ':latest|^FROM [^@]+$' "$(dirname "$0")/../../Dockerfile"; then
	fail "every FROM needs a tag and a digest"
fi

docker run --rm --entrypoint sh "$image" -c '
	set -e
	pg_dump --version | grep -q "PostgreSQL) 18\."
	cwebp -version >/dev/null
	writable=$(find / -xdev -type d -writable 2>/dev/null | grep -vE "^/(proc|sys|dev)" | sort)
	for d in $writable; do
		case "$d" in /app/storage*|/tmp*|/var/tmp*) ;; *) echo "unexpected writable dir: $d" >&2; exit 1 ;; esac
	done
' || fail "runtime checks failed"

echo "check-image: ok (${mb} MB compressed, user app)"
