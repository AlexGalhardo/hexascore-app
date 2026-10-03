# Changelog

All notable changes to this project are documented here.
Format: [Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/). Versioning: [SemVer](https://semver.org/).

## [Unreleased]

### Added

- `PLAN.md` with decisions and milestones; reference docs in `docs/` (architecture, data model, API, frontend, design system, payments, ingestion, security, testing, process).
- `AGENTS.md` and `CLAUDE.md` agent guides.
- Catalog (0.1): rating categories per type (`lib/categories.ts`), pure catalog repository (`lib/catalog/`) with keyset pagination, filters, `pg_trgm` + `tsvector` search, facets, external scores, affiliate links and public ratings; tagged `unstable_cache` layer (`catalog`, `item:<id>`).
- Pages: landing highlights per section, `/games` `/movies` `/series` `/books` with URL filters and cursor pagination, item page (score badge, hexagonal radar, users' scores dialog, external scores, click-to-load trailers, details, where to buy, synopsis attribution, `RatingPanel` slot), `/search`, navbar autocomplete.
- Route handlers: `GET /api/items/[id]/ratings`, `GET /api/search/suggest`, `POST /api/cron/revalidate` (Bearer `CRON_SECRET`).
- `components/score/` (`ScoreBadge`, `HexRadar`, `ScoreBar`) and i18n namespaces `categories`, `catalog`, `item`, `search` (en, pt-BR, es).
- Deterministic dev seed `bun run db:seed` (`@faker-js/faker` 10.6.0): 50 users, 400 items, ratings; refuses `VERCEL_ENV=production`.
- PGlite test helper `createTestDatabase()` and repository tests.

### Fixed

- i18n proxy matcher only matched `/` (unescaped dot), so locale-less routes like `/games` returned 404.
- PGlite under Next: listed in `serverExternalPackages` (bundling broke its WASM path) and in-memory during `next build` (parallel workers could not share `.pglite`).

### Removed

- Files inherited from the previous project (RespondeAê): old plan, docs, setup scripts, tests, REST client samples, graphify output, `vercel.json`.
