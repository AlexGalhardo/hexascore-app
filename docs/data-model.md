# Modelo de dados


**Catálogo**
- `items`: `id` uuid v7 (público), `type` (`game|movie|series|book`), `slug` (único por tipo), `wikidata_qid` (único), `release_date`, `external_ids jsonb` (imdb, igdb, steam, tmdb, isbn13, openlibrary, hltb, goodreads, metacritic, rt, epic, gog), `meta jsonb` por tipo (genres, platforms, runtime, pages, seasons, trailers; índice GIN), stats desnormalizadas `rating_count`, `score`, `score_c1..c6` (atualizadas na mesma transação da avaliação).
- `item_translations` (`item_id`, `locale`) PK: `title`, `synopsis`, `synopsis_license`, `synopsis_source_url`, `search tsvector` gerado, índice `pg_trgm` em `title`.
- `external_scores` (`item_id`, `source`, `score`, `url`, `fetched_at`) — **nunca sai pela API**.
- `affiliate_links` (`item_id`, `store`, `url`, `active`).

**Avaliação e social**
- `ratings` (`user_id`, `item_id`) PK: `c1..c6 smallint CHECK (0..100)`, `overall` gerado = `round((c1+…+c6)/6.0)`. Nomes das categorias vivem em código.
- `item_status` (`user_id`, `item_id`, `status`) · `follows` (`follower_id`, `followee_id`) · `lists`, `list_items` · `reports`.

**Auth** (gerado pelo Better Auth CLI): `user`, `session`, `account`, `verification`, `two_factor`, `passkey`. Extras em `user`: `nickname` único, `role`, `profile_private`, `deletion_scheduled_at`.

**API e billing**
- `api_keys`: plugin apiKey do Better Auth se servir; senão 1 tabela (`prefix`, `hash` sha256, `revoked_at`, `last_used_at`).
- `subscriptions` (`user_id`, `plan`, `provider`, `status`, `current_period_end`, `provider_ref`).
- `credit_ledger` (`user_id`, `delta`, `reason`, `ref`) — saldo = soma, append-only.
- `pix_charges` (`provider_ref` único, `amount_cents`, `purpose`, `status`).
- `stripe_customers` (`user_id` PK, `customer_id` único) — 1 Customer Stripe por usuário, reusado em assinatura e créditos.
- `processed_events` (`id` do evento, `provider`) — idempotência de webhooks e dos e-mails de lembrete PIX.
- `api_usage_daily` (`user_id`, `date`, `count`).
- Últimas 100 requests por usuário ficam no Redis (`LPUSH` + `LTRIM`), não no Postgres.

**Outros**: `contact_messages` · `source_records` (`source`, `external_id`) PK com `payload jsonb`, `hash`, `fetched_at`, `item_id`.

Fora por ora: tabela de gêneros/plataformas (entra quando gênero ganhar página traduzida), `ingestion_runs` (logs do Actions bastam).
