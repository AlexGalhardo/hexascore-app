#!/usr/bin/env bash
# Seeds a LOCAL load-test database (never production): K6_KEYS Business-plan users with one hx_live_ key each, and
# K6_LISTS users with one public list of 25 games each. Idempotent per run through a random suffix.
#   PSQL="docker exec -i hx-senior-postgres-1 psql -U hexascore -d hexascore" qa/k6/seed.sh > /tmp/k6.env
# Prints shell exports for the k6 scripts: API_KEYS (comma separated) and LIST_PATHS (comma separated /u/<nick>/lists/<id>).
set -euo pipefail

PSQL=${PSQL:?set PSQL to a psql command line connected to the target database}
KEYS=${K6_KEYS:-10}
LISTS=${K6_LISTS:-10}
RUN=$(openssl rand -hex 3)
# usernames must match ^[a-z]{4,16}$
LET=$(printf %s "$RUN" | tr '0-9' 'g-p')
letter() { printf "\x$(printf %x $((96 + $1)))"; }

q() { $PSQL -v ON_ERROR_STOP=1 -At -q -c "$1"; }

keys=()
for i in $(seq 1 "$KEYS"); do
	key="hx_live_$(openssl rand -base64 64 | tr -dc 'A-Za-z0-9' | head -c 32)"
	hash=$(printf %s "$key" | sha256sum | cut -d' ' -f1)
	q "WITH u AS (
		INSERT INTO users (name, email, password, email_verified_at, username, created_at, updated_at)
		VALUES ('k6 api $i', 'k6-api-$RUN-$i@example.test', 'x', now(), 'kapi$LET$(letter "$i")', now(), now()) RETURNING id
	), s AS (
		INSERT INTO subscriptions (user_id, stripe_subscription_id, status, price_id, plan, state, current_period_end, created_at, updated_at)
		SELECT id, 'sub_k6_$RUN$i', 'active', 'price_k6', 'business', 'active', now() + interval '30 days', now(), now() FROM u
	) INSERT INTO api_keys (id, user_id, name, prefix, hash) SELECT gen_random_uuid(), id, 'k6', '${key:0:12}', '$hash' FROM u" >/dev/null
	keys+=("$key")
done

paths=()
for i in $(seq 1 "$LISTS"); do
	row=$(q "WITH u AS (
		INSERT INTO users (name, email, password, email_verified_at, username, created_at, updated_at)
		VALUES ('k6 list $i', 'k6-list-$RUN-$i@example.test', 'x', now(), 'klst$LET$(letter "$i")', now(), now()) RETURNING id, username
	), l AS (
		INSERT INTO lists (id, user_id, type, title, is_public, item_count)
		SELECT gen_random_uuid(), id, 'game', 'Load list $i', true, 25 FROM u RETURNING id, user_id
	), li AS (
		INSERT INTO list_items (list_id, item_id, position)
		SELECT l.id, it.id, it.rn FROM l, (
			SELECT id, row_number() OVER (ORDER BY id) AS rn FROM items WHERE type = 'game' AND visibility = 'public' LIMIT 25
		) it
	) SELECT '/u/' || u.username || '/lists/' || l.id FROM u, l")
	paths+=("$row")
done

IFS=,
echo "export API_KEYS='${keys[*]}'"
echo "export LIST_PATHS='${paths[*]}'"
