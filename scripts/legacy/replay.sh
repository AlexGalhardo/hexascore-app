#!/usr/bin/env bash
# Apply hexascore:export-since SQL to a legacy-shaped database (SDD 11 §7.6 R3).
# Run only after R2 (Neon read-only off). Never against production by itself.
#   hexascore:export-since 2026-10-30T00:00:00Z | TARGET_DATABASE_URL=postgres://... LEGACY_SCHEMA=public scripts/legacy/replay.sh
set -euo pipefail

: "${TARGET_DATABASE_URL:?set TARGET_DATABASE_URL}"
schema=${LEGACY_SCHEMA:-public}

if ! printf '%s' "$schema" | grep -Eq '^[A-Za-z_][A-Za-z0-9_]*$'; then
	printf 'replay: LEGACY_SCHEMA must be a bare identifier\n' >&2
	exit 2
fi

{
	printf 'SET search_path TO %s;\nBEGIN;\n' "$schema"
	cat
	printf '\nCOMMIT;\n'
} | psql "$TARGET_DATABASE_URL" -v ON_ERROR_STOP=1 -q
