# Modelo de dados


**Catálogo**
- `items`: `id` uuid v7 (público), `type` (`game|movie|series|book`), `slug` (único por tipo), `wikidata_qid` (único), `release_date`, `external_ids jsonb` (imdb, igdb, steam, tmdb, isbn13, openlibrary, hltb, goodreads, metacritic, rt, epic, gog), `meta jsonb` por tipo (genres, platforms, runtime, pages, seasons, trailers; índice GIN), stats desnormalizadas `rating_count`, `score`, `score_c1..c6` (atualizadas na mesma transação da avaliação).
- `item_translations` (`item_id`, `locale`) PK: `title`, `synopsis`, `synopsis_license`, `synopsis_source_url`, `search tsvector` gerado, índice `pg_trgm` em `title`.
- `external_scores` (`item_id`, `source`, `score`, `url`, `fetched_at`) — **nunca sai pela API**.
- `affiliate_links` (`item_id`, `store`, `url`, `active`).

**Avaliação e social**
- `ratings` (`user_id`, `item_id`) PK: `c1..c6 smallint CHECK (0..100)`, `overall` gerado = `round((c1+…+c6)/6.0)`. Nomes das categorias vivem em código.
- `item_status` (`user_id`, `item_id`, `status`) · `time_to_beat` (`user_id`, `item_id`) PK: `main_minutes`, `extras_minutes`, `completionist_minutes` (integer, nullable, CHECK 30..600000 = 0,5..10000 h, pelo menos um preenchido; índice `time_to_beat_item_idx`) · `follows` (`follower_id`, `followee_id`) · `lists`, `list_items` · `reports`.

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

**0.3 (implementado)**
- Stats desnormalizadas de `items` são recalculadas a partir de `ratings` (nunca incrementais) na mesma transação da escrita, com `SELECT … FOR UPDATE` na linha do item. Guardam a média **bruta** arredondada mesmo abaixo de 5 avaliações; o limiar `MIN_RATINGS_FOR_SCORE` é aplicado só na leitura (`lib/catalog/repository.ts`). Detalhes: skill `hexascore-ratings`.
- Migration `0003_social_moderation`: `reports.details text` e índice único parcial `reports_open_unique_idx (reporter_id, target_type, target_id) WHERE status = 'open'` (dedupe de denúncias abertas via `ON CONFLICT DO NOTHING`). `reports.target_id` é o id do usuário ou da lista; `reports.reason` ∈ `spam|offensive|harassment|impersonation|other` (zod).
- `lists.hidden` = oculta pela moderação. Ordem em `list_items.position` (append = max+1; reordenar troca com o vizinho).
- Merge de itens (`lib/admin.ts#mergeItems`): copia ratings, item_status, time_to_beat (filtrado), list_items, affiliate_links e external_scores para o alvo (`ON CONFLICT DO NOTHING` → a linha do alvo vence), repointa `source_records.item_id`, apaga a origem e recalcula as stats do alvo, numa transação.
- Tempo para zerar (migration `0004_time_to_beat`, `lib/time-to-beat.ts`): só jogos e só com `item_status = completed` (checado na escrita com `FOR UPDATE` na linha do status). Sair de `completed` (ou limpar o status) apaga o relato na mesma transação (`setItemStatus` → `"cleared"` → `updateTag(item:<id>)`). Medianas (`percentile_cont(0.5)`, ignora nulos por estilo) e contagens calculadas na leitura dentro de `getItemById` (cacheadas com a tag `item:<id>`); mediana `null` abaixo de `MIN_REPORTS_FOR_TIME_TO_BEAT` (3). Merge move relatos só onde o status final do usuário no alvo é `completed`. Exclusão de conta anonimiza o usuário e mantém os relatos (agregado anônimo, como `ratings`).
- Papel admin em `user.role` (string separada por vírgula do plugin admin do Better Auth); `bun run admin:grant <email>`.

Fora por ora: tabela de gêneros/plataformas (entra quando gênero ganhar página traduzida), `ingestion_runs` (logs do Actions bastam).
