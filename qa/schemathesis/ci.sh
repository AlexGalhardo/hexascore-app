#!/usr/bin/env bash
# Schemathesis 4.29.0 against a loaded image. Does not build the image.
# Usage: qa/schemathesis/ci.sh [hexascore:ci]
# Postgres must already be listening (CI service, or a local one via DATABASE_URL).
set -euo pipefail

image=${1:-hexascore:ci}
root=$(cd "$(dirname "$0")/../.." && (pwd -W 2>/dev/null || pwd))
cd "$root"

port=${QA_PORT:-3000}
db=${DATABASE_URL:-postgres://hexascore:hexascore@127.0.0.1:5432/hexascore?sslmode=disable}
pg_image=${PG_IMAGE:-postgres:18.6-alpine3.24@sha256:10a5c69a83035ab5239f5620d708520e8b28f5aa1f2b20143e6024532edadfb5}
name=hx-openapi-$$
key=hx_live_CIopenAPIScanKey0000000000000000
app_id=ciscan
app_secret=ciscanSecret0123456789abcdef
app_token=hx_app_${app_id}.${app_secret}

if [ "${#key}" -ne 40 ]; then
	echo "CI API key must be 40 characters" >&2
	exit 1
fi
key_hash=$(printf '%s' "$key" | openssl dgst -sha256 | awk '{print $NF}')
app_hash=$(printf '%s' "$app_secret" | openssl dgst -sha256 | awk '{print $NF}')
prefix=${key:0:14}

export APP_KEY="${APP_KEY:-base64:$(openssl rand -base64 32)}"
app_env=(
	-e APP_ENV=production
	-e APP_KEY="$APP_KEY"
	-e APP_URL="http://127.0.0.1:${port}"
	-e DATABASE_URL="$db"
	-e RATE_LIMIT_STORE=cache
	-e CACHE_STORE=database
	-e STRIPE_KEY=sk_test_placeholder_ci
	-e STRIPE_WEBHOOK_SECRET=whsec_placeholder_ci
	-e STRIPE_PRICE_PRO=price_placeholder_pro
	-e STRIPE_PRICE_BUSINESS=price_placeholder_business
	-e MAIL_MAILER=log
)

cleanup() {
	docker rm -f "$name" >/dev/null 2>&1 || true
}
trap cleanup EXIT

docker run --rm --network host "${app_env[@]}" "$image" migrate

docker run -i --rm --network host -e PGPASSWORD=hexascore "$pg_image" \
	psql -h 127.0.0.1 -U hexascore -d hexascore -v ON_ERROR_STOP=1 <<SQL
INSERT INTO users (name, email, password, email_verified_at, created_at, updated_at)
VALUES ('CI Scan', 'schemathesis-ci@hexascore.test', 'not-a-login', now(), now(), now())
RETURNING id \gset
INSERT INTO subscriptions (user_id, stripe_subscription_id, status, price_id, plan, provider, state, current_period_end, created_at, updated_at)
VALUES (:id, 'sub_ci_openapi', 'active', 'price_business', 'business', 'stripe', 'active', now() + interval '365 days', now(), now());
INSERT INTO api_keys (id, user_id, name, prefix, hash)
VALUES (gen_random_uuid(), :id, 'ci', '${prefix}', '${key_hash}');
INSERT INTO access_tokens (id, user_id, name, hash, abilities, created_at)
VALUES ('${app_id}', :id::text, 'ci', '${app_hash}', '["*"]', now());
SQL

docker run -d --name "$name" --network host -e "PORT=${port}" "${app_env[@]}" "$image"

for _ in $(seq 1 40); do
	if curl -fsS "http://127.0.0.1:${port}/up" >/dev/null; then
		break
	fi
	sleep 1
done
curl -fsS "http://127.0.0.1:${port}/up" >/dev/null

HX_TEST_KEY="$key" QA_BASE_URL="http://127.0.0.1:${port}" qa/schemathesis/run.sh openapi/api-v1.json
HX_TEST_KEY="$app_token" QA_BASE_URL="http://127.0.0.1:${port}" qa/schemathesis/run.sh openapi/app-v1.json

# Anonymous search, 1 rps, under the 60/min IP limit. Host network reaches the app on 127.0.0.1.
docker run --rm --network host \
	-e BASE_URL="http://127.0.0.1:${port}" \
	-v "$root/qa:/qa" \
	grafana/k6:2.3.0 run /qa/k6/app-v1.js
