#!/usr/bin/env bash
# Builds database/concurrent/<name>.sql outside a transaction (CREATE INDEX
# CONCURRENTLY cannot run in a migration, VF-017) after dropping an INVALID
# leftover of a failed earlier build. Run after `migrate`, as a pre-deploy step.
#
#   PSQL='psql "$DATABASE_URL"' scripts/concurrent-index.sh items_title_trgm
#
# WATCH_LOCKS=1 samples pg_locks during the build and reports whether an
# AccessExclusiveLock on the indexed table was ever granted.
set -euo pipefail

name=${1:?usage: concurrent-index.sh <index-name>}
[[ $name =~ ^[a-z0-9_]+$ ]] || { echo "invalid index name: $name" >&2; exit 2; }

file="$(dirname "$0")/../database/concurrent/$name.sql"
[[ -f $file ]] || { echo "no such file: $file" >&2; exit 2; }

PSQL=${PSQL:-psql "$DATABASE_URL"}
run() { eval "$PSQL"' -X -q -v ON_ERROR_STOP=1 "$@"'; }

if [[ $(run -Atc "select count(*) from pg_index i join pg_class c on c.oid = i.indexrelid where c.relname = '$name' and not i.indisvalid") != 0 ]]; then
	echo "dropping INVALID leftover $name"
	run -c "DROP INDEX CONCURRENTLY IF EXISTS $name"
fi

table=$(grep -oiE ' ON +[a-z0-9_]+ +USING' "$file" | head -1 | awk '{print $2}')
if [[ ${WATCH_LOCKS:-} == 1 ]]; then
	# A server-side sampler (one DO block, ends when the index is valid or after
	# 120 s) avoids a round trip per sample, which would miss a short build.
	sampler=$(mktemp)
	run -f - >"$sampler" 2>&1 <<SQL &
DO \$\$
DECLARE n int := 0; excl int := 0; modes text[] := '{}'; m text; deadline timestamptz := clock_timestamp() + interval '120 seconds';
BEGIN
  LOOP
    FOR m IN SELECT l.mode FROM pg_locks l JOIN pg_class c ON c.oid = l.relation
             WHERE c.relname = '$table' AND l.granted AND l.pid <> pg_backend_pid() LOOP
      IF m = 'AccessExclusiveLock' THEN excl := excl + 1; END IF;
      IF NOT (m = ANY(modes)) THEN modes := modes || m; END IF;
    END LOOP;
    n := n + 1;
    EXIT WHEN clock_timestamp() > deadline OR EXISTS (SELECT 1 FROM pg_index i JOIN pg_class c ON c.oid = i.indexrelid WHERE c.relname = '$name' AND i.indisvalid);
    PERFORM pg_sleep(0.02);
  END LOOP;
  RAISE NOTICE 'lock samples=% AccessExclusiveLock_samples=% modes_seen=%', n, excl, modes;
END \$\$;
SQL
	watcher=$!
	sleep 0.3
	run -f - <"$file"
	wait "$watcher"
	grep NOTICE "$sampler" || cat "$sampler"
	rm -f "$sampler"
else
	run -f - <"$file"
fi

valid=$(run -Atc "select i.indisvalid from pg_index i join pg_class c on c.oid = i.indexrelid where c.relname = '$name'")
echo "index $name indisvalid=$valid"


[[ $valid == t ]]
