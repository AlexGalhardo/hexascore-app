# ADR 0011: Search on PostgreSQL FTS + pg_trgm behind one interface

Status: accepted (2026-10-28). Decisions: brainstorming D-16; ADR 0003.

## Context

Legacy search (`legacy/apps/web/lib/catalog/repository.ts`): pg_trgm `word_similarity` on titles with a per-transaction threshold, `websearch_to_tsquery('simple')` on a generated tsvector over title + synopsis, ranked by `greatest()`; GIN indexes on the tsvector and a trigram index on titles; `unaccent` installed but unused; a suggest endpoint for autocomplete. Catalogue size ~100k items plus entities after SDD 01. Budget ceiling US$30/month (D-08) leaves no room for a search service.

## Decision

Keep search in PostgreSQL: per-locale tsvector columns (configurations `english`, `portuguese`, `spanish` with `unaccent` in the pipeline for the Latin locales), trigram similarity on titles and entity names for typo tolerance, a combined ranking that prefers exact and prefix title matches, then full-text rank, then popularity prior as tie-breaker; a separate lightweight suggest index for autocomplete (title, type, year, cover thumb) with a p95 budget of 50 ms. Entity search (people, organisations, franchises) joins the same interface. The search layer is one Go interface (`Search(ctx, query) (results, error)` plus `Suggest`) with the Postgres implementation as the only one at launch; nothing outside the package touches search SQL, so a future engine swap is contained.

## Alternatives considered

- **Meilisearch or Typesense on Railway.** Instant search, typo tolerance and facets out of the box; a second service (+US$5–15/month, over the ceiling), an index to keep in sync, a new dependency ADR. Deferred; the interface boundary makes it a contained change.
- **ParadeDB / pg_search extension.** BM25 inside Postgres, but availability on Railway's managed Postgres is unverified. Not chosen; can be evaluated behind the same interface.
- **vukem built-in search.** Not in the vukem docs; nothing to use.

## Consequences

- Positive: no extra service; transactional consistency with the catalogue; search quality adequate for 100k–1M rows with proper indexes.
- Negative: no facet counts in milliseconds at large scale; ranking tuning is manual; index size counts against the Postgres plan.
- Obligations: SDD 07 defines columns, indexes, ranking, per-locale behaviour, the k6 budget for `/search` and `/suggest`, and a relevance test set (golden queries) that runs in CI.
