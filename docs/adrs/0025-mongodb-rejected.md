# ADR 0025: MongoDB is rejected; raw source payloads stay in Postgres `jsonb`

Status: accepted (2026-11-02, architect audit Phase 4). Refines ADR 0003. Related: SDD 06 §5 (`source_records`, `source_record_history`), RFC-001 (catalogue worker), D-27 (US$20 hard limit).

## Context

The only document-shaped workload in HexaScore is the ingestion layer: raw payloads from Wikidata, Wikipedia, Open Library, TMDB, IGDB, Steam and GOG, kept for provenance, diffing and re-normalisation. Today they live in `source_records` (one row per source × target, `jsonb` payload, SHA-256 content hash, `checked_at`) and `source_record_history`, with retention windows applied by `ingest:prune` (SDD 06 §5, `routes/schedule.go`).

Size estimate at the 1 M-item target (RFC-001 §6): about 7 source records per item on average, 2 to 8 KB each after trimming (Wikidata entities are trimmed to the properties the normaliser reads, SDD 06 §7.3), so roughly 7 M rows and 25 to 40 GB before history. That is the dominant storage cost of the whole platform, whatever the engine.

What MongoDB would add: schemaless writes (already true of `jsonb`), horizontal sharding (not needed below a few hundred GB), aggregation pipelines (Postgres has `jsonb_path_query`, GIN indexes and SQL), change streams (`LISTEN/NOTIFY` or the `api_changes` table already exist). What it would cost: a second database service on Railway with its own volume (the same 25 to 40 GB, billed per GB), a second backup pipeline (ADR 0014 covers Postgres only), a second driver in the app (`vukem/drivers/mongo` exists but nothing in the app uses it), and two sources of truth for provenance.

Postgres-only measures that keep the raw-payload table cheap:

- Store only the trimmed payload (already done) and compress large text with `TOAST` (automatic) or `lz4` column compression (`ALTER TABLE … SET COMPRESSION lz4`, Postgres 14+).
- Partition `source_records` and `source_record_history` by source (or by month for history) so pruning is `DROP PARTITION` instead of `DELETE`.
- Index only `(source, external_id)`, `(target_id)`, `(checked_at)` and the content hash; no GIN on the payload unless a query needs it.
- Keep history only for fields that changed (diff rows), which SDD 06 already specifies.
- Move cold history older than 90 days to the bucket as NDJSON (`export` writer exists) if the volume approaches 70 % of the plan.

## Decision

MongoDB is **rejected** for the foreseeable scope (1 M items). `source_records` and `source_record_history` stay in Postgres `jsonb` with the measures above adopted as backlog items of RFC-001 (partitioning and column compression before the 250 k milestone, cold export before 500 k).

## Alternatives considered

- **MongoDB for raw payloads.** Rejected: duplicates the cost of the largest table, adds a backup and restore path, and solves no problem Postgres has at this size.
- **Keep raw payloads in the S3 bucket only.** Partially adopted as the cold tier; hot rows stay in Postgres because the normaliser diffs them on every run.
- **Drop raw payloads, keep only normalised items.** Rejected: provenance per field (ADR 0006) and safe re-normalisation need the originals.

## Consequences

- Positive: one database, one backup, one driver; the Postgres volume is the single capacity number to watch (cost runbook step 4).
- Negative: the Postgres volume grows with the catalogue; the 10 GB volume must be resized before the 250 k milestone (RFC-001 §6 gives the projection), which is an owner action on Railway.
- Reassessment trigger: `source_records` above 150 GB, or a query pattern that needs cross-document joins the SQL layer cannot express in under 100 ms p95. Either reopens this ADR; the answer is more likely partitioning or ClickHouse-style analytics than MongoDB.
- Rollback: none needed; nothing is adopted.
