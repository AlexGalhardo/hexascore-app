# SDD 01: Catalog

Status: reviewed (2026-10-28 consistency pass). Decisions: D-01, D-03, D-09, D-10, D-13, D-16, D-17, A-04, A-06. ADRs: 0003, 0006, 0012. Related SDDs: 02, 04, 05, 06, 07, 08, 09, 10, 11, 12.

## 1. Objective

The catalogue is the shared spine of HexaScore: games, movies, series and books, plus the people, organisations, franchises, platforms, genres, game modes, countries and languages around them (D-01). It gives users fast item and entity pages in three locales with stable URLs (D-17), turns every creator, publisher or platform into a browsable page (discovery), gives developers per-field provenance and licence (D-03), and gives scoring (SDD 02) one stable item identity, with books rated per work (D-10).

## 2. Scope

In scope (v1): `items` for four types plus book `editions` (D-10); `entities` with kinds and role-bearing relations (D-01); translations with licence (D-17 slugs, redirects); per-field provenance and audience visibility (D-03); cover keys; popularity prior; read models for web, app and paid API; admin merge; sitemap data; data quality asserts; legacy migration (A-04).

Out of scope:

- Score formula and denormalised score columns: SDD 02 (referenced only).
- Trending and similar items: SDD 08 (D-02). Search ranking: SDD 07 (D-16); this SDD only fixes the columns it indexes.
- Ingestion connectors, normalisation, cover processing jobs: SDD 06.
- Community edits to catalogue fields: SDD 04 (writes through the provenance model defined here).
- Critic scores: removed (D-09). Edition-level contributors (translator, illustrator), engines as entities, episode-level data: after v1.0.0 (D-30); the relation model (`item_entities` roles, `entity_links`) leaves room for them without a schema rewrite.

## 3. Functional requirements

- FR-01-1: Every item has exactly one `type` in `game|movie|series|book`, a unique `wikidata_qid`, and a slug unique per type. Verify: integration test (Postgres) inserting duplicates expects unique violations.
- FR-01-2: A book item is a work; it has 0..n editions, each with optional ISBN-13 that, when present, passes the checksum. Ratings and statuses reference `items.id`, never `editions.id`. Verify: integration test; DB CHECK rejects `9780306406158` (bad digit) and accepts `9780306406157`.
- FR-01-3: An edition can only reference an item of type `book`. Verify: integration test expecting FK violation for a movie id.
- FR-01-4: Entity kinds are fixed (§5.2); an organisation's role lives on the relation, not the entity. Verify: role/kind unit test; DQ-08.
- FR-01-5: Item detail lists related entities grouped by role in `position` order; cast is capped at 20 with `castTotal`. Verify: integration test with 30 cast rows.
- FR-01-6: Names resolve requested locale → `en` → `default_name`, reporting the locale used. Verify: resolver unit test.
- FR-01-7: Slugs are the same in every locale; a changed slug leaves a redirect, and the old URL answers 301 to the new one. Verify: Playwright test on `/pt-br/games/<old>`.
- FR-01-8: Admin merge of item A into B (same type) is atomic and leaves id and slug redirects (§7.5). Verify: integration test; `GET /api/v1/games/{A}` → 301 to `{B}`; the audit event holds the jsonb snapshot of both items (§7.5).
- FR-01-9: The paid API never emits a field whose source is not in `CATALOG_API_SOURCES`; the site and app may show it with attribution. Verify: contract test over a fixture with TMDB/IGDB-sourced fields, run with default config and with `tmdb` added.
- FR-01-10: Listings filter by genre slug, platform slug, year, min/max score (stars 0–10, D-13), `min_<category>` (A-06 keys) and sort by `-popularity` (default), `-score`, `score`, `-release_date`, `release_date`, `-rating_count`, with opaque cursor pagination. Verify: integration tests per filter; Schemathesis.
- FR-01-11: Entity detail shows item counts per type and an average score per type when at least 5 items of that type have a visible score. Verify: integration test at 4 and 5 scored items.
- FR-01-12: Sitemaps list every `public` item and every entity with at least 3 linked `public` items (open issues 01-8, 09-8), with `hreflang` alternates for en, pt-br, es and `x-default`; entity pages below the threshold render `noindex,follow` until they qualify. Verify: integration test parsing the XML at 2 and 3 linked items, plus a robots-meta assertion.
- FR-01-13: `vukem catalog:check` runs the data quality asserts of §7.6 and exits non-zero on any `error`-level violation. Verify: integration test with seeded violations.
- FR-01-14: Items whose `visibility` is not `public` are excluded from trending (SDD 08), search (SDD 07), sitemaps, listings, the paid API and the app API (404 `not-found`); `hidden` item pages are visible to admins only, `takedown` item pages answer 410 on the web (open issues 08-3, 14-3). Verify: integration test per surface with one `hidden` and one `takedown` fixture.
- FR-01-15: `release_status` and `release_precision` are always set; SDD 02's release gate reads them (open issue 02-1). Verify: CHECK constraint test; DQ-13.

## 4. Non-functional requirements

- NFR-01-1 Performance: item detail read model ≤ 6 SQL queries and p95 ≤ 30 ms DB time at 500k items / 3M `item_entities` rows; listing page (limit 25) p95 ≤ 50 ms. Verify: Go benchmark against a seeded Postgres; k6 in SDD 05/09 budgets.
- NFR-01-2 Payload: item detail JSON ≤ 32 KB (gzip ≤ 8 KB) for the 99th-percentile item. Verify: test over the 100 largest fixture items.
- NFR-01-3 Pagination bounds: `limit` 1–100, default 25 (parity with `legacy/apps/web/lib/catalog/types.ts`); cursor ≤ 512 bytes; tampered cursor → 400. Verify: unit tests.
- NFR-01-4 Licensing: the API source allowlist fails closed; a new source is invisible to the API until added to config. Verify: FR-01-9 test with an unknown source code.
- NFR-01-5 i18n: locales exactly `en`, `pt-BR`, `es`; every item has an `en` title. Verify: DQ-01.
- NFR-01-6 Storage: catalogue tables ≤ 2 GB at 500k items (D-08). Verify: monthly `pg_total_relation_size` report (SDD 11).
- NFR-01-7 Integrity: merge and slug change are atomic; no redirect chain longer than 1 hop. Verify: DQ-05.
- NFR-01-8 Freshness: `entity_stats` recomputed hourly; stale > 2 h is a DQ warning. Verify: DQ-12.

## 5. Data model

### 5.1 Decision: one `items` table

One table for four types (legacy parity, `legacy/apps/web/lib/db/schema/catalog.ts`): ratings, statuses, lists, edits, trending and search reference one key; cross-type search and entity pages are one query; type-specific scalars are few and nullable. D-10 needs no split: the work is the item, the edition a child row. Legacy `meta` jsonb is dissolved into entity relations and typed columns.

### 5.2 DDL sketch (PostgreSQL)

Migrations use vukem `schema` (https://vukem.dev/docs/database/migrations "Migrations"). Generated columns, GIN/trigram indexes, SQL functions and composite foreign keys go through `s.Raw`; whether `t.JSON` maps to `jsonb` on Postgres is [NOT VERIFIED], so `jsonb` columns are created with `s.Raw`. Enumerations are `text` + `t.Check` because the Postgres mapping of `t.Enum` is not documented [NOT VERIFIED]. Ids are `uuid` v7 set in Go (`uuid.NewV7`, https://vukem.dev/docs/database/migrations "UUIDs"; "Models" says to set non-integer keys before `Create`).

```sql
create table data_sources (
  code text primary key,                 -- wikidata, wikipedia, openlibrary, tmdb, igdb, steam, gog, hexascore
  name text not null,
  licence text not null,                 -- SPDX-like: CC0-1.0, CC-BY-SA-4.0, proprietary-tmdb, ...
  attribution text,                      -- text shown on site/app
  url text not null
);

create table items (
  id uuid primary key,
  type text not null check (type in ('game','movie','series','book')),
  slug text not null check (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$' and length(slug) <= 120),
  wikidata_qid text not null unique check (wikidata_qid ~ '^Q[1-9][0-9]*$'),
  original_title text,
  release_date date,                        -- month/year precision: stored as the 1st of the period
  release_precision text not null default 'unknown'
    check (release_precision in ('day','month','year','unknown')),
  release_status text not null default 'released'
    check (release_status in ('announced','released','cancelled')),  -- SDD 02 release gate (02-1)
  visibility text not null default 'public'
    check (visibility in ('public','hidden','takedown')),            -- replaces trending_exclusions (08-3)
  runtime_minutes int check (runtime_minutes between 1 and 6000),   -- movie
  season_count smallint, episode_count int,                          -- series
  page_count int check (page_count between 1 and 20000),             -- book (work-level, typical)
  primary_edition_id uuid,                                           -- book cover fallback
  trailers jsonb not null default '[]',                              -- [{youtubeId,title}]
  external_ids jsonb not null default '{}',
  provenance jsonb not null default '{}',  -- {"release_date":"wikidata","runtime_minutes":"tmdb",...}
  cover_key text, cover_thumb_key text, cover_source text references data_sources(code),
  cover_attribution jsonb,                 -- {author,url} when the licence needs it
  popularity_prior real not null default 0, -- Wikidata sitelinks count; trending is SDD 08
  -- score columns (score_raw, score_c1..c6, rating_count, confidence): owned by SDD 02
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  unique (type, slug),
  unique (id, type)                        -- target of the editions composite FK
);
create index items_type_popularity_idx on items (type, popularity_prior desc, id);
create index items_type_release_idx on items (type, release_date, id);
create index items_external_ids_gin on items using gin (external_ids jsonb_path_ops);
create index items_public_idx on items (type, popularity_prior desc, id) where visibility = 'public';

create table item_translations (
  item_id uuid not null references items(id) on delete cascade,
  locale text not null check (locale in ('en','pt-BR','es')),
  title text not null check (length(title) between 1 and 500),
  title_source text not null references data_sources(code),
  synopsis text,
  synopsis_source text references data_sources(code),
  synopsis_licence text,                   -- CC-BY-SA-4.0 for Wikipedia
  synopsis_source_url text,
  primary key (item_id, locale),
  check (synopsis is null or (synopsis_source is not null and synopsis_licence is not null))
);
-- search tsvector and trigram indexes on title: defined by SDD 07

create table item_alt_titles (            -- aliases, romanisations, former titles (07-6); filled by SDD 06
  item_id uuid not null references items(id) on delete cascade,
  title text not null check (length(title) between 1 and 500),
  lang text,                               -- BCP 47 with script (ja-Latn, ru-Latn); null when unknown
  kind text not null check (kind in ('alias','transliteration','original','former')),
  source text not null references data_sources(code),
  primary key (item_id, title)
);
-- title_norm column, trigram index and unique (item_id, title_norm): SDD 07 §5

create table item_links (                 -- item -> item relations (08-1); filled by SDD 06
  from_item_id uuid not null references items(id) on delete cascade,
  to_item_id uuid not null references items(id) on delete cascade,
  relation text not null check (relation in ('based_on')),           -- Wikidata P144
  source text not null references data_sources(code),
  primary key (from_item_id, relation, to_item_id),
  check (from_item_id <> to_item_id)
);
create index item_links_to_idx on item_links (to_item_id, relation);

create function isbn13_is_valid(t text) returns boolean language sql immutable as $$
  select t ~ '^97[89][0-9]{10}$' and (
    select sum(substr(t, i, 1)::int * case when i % 2 = 0 then 3 else 1 end)
    from generate_series(1, 13) i) % 10 = 0 $$;

-- migration order: entities is created before editions (FKs below)
create table editions (
  id uuid primary key,
  work_id uuid not null,
  work_type text not null default 'book' check (work_type = 'book'),
  foreign key (work_id, work_type) references items(id, type) on delete cascade,
  isbn13 char(13) unique check (isbn13 is null or isbn13_is_valid(isbn13)),
  wikidata_qid text unique, openlibrary_id text unique,
  title text not null, subtitle text,
  language text not null,                  -- BCP 47: pt-BR, en, es, ...
  market_country char(2),                  -- ISO 3166-1 alpha-2
  is_brazilian boolean generated always as (market_country = 'BR' or language = 'pt-BR') stored,
  publisher_id uuid references entities(id), imprint_id uuid references entities(id),
  publication_year smallint check (publication_year between 1400 and 2100),
  page_count int, format text check (format in ('hardcover','paperback','ebook','audiobook','other')),
  cover_key text, cover_thumb_key text, cover_source text references data_sources(code),
  source text not null references data_sources(code),
  created_at timestamptz not null default now(), updated_at timestamptz not null default now(),
  check (isbn13 is not null or wikidata_qid is not null or openlibrary_id is not null)
);
create index editions_work_idx on editions (work_id, is_brazilian desc, publication_year desc);

create table entities (
  id uuid primary key,
  kind text not null check (kind in ('person','organization','franchise','platform',
                                     'genre','game_mode','country','language')),
  slug text not null check (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$' and length(slug) <= 120),
  wikidata_qid text unique check (wikidata_qid ~ '^Q[1-9][0-9]*$'),  -- null = HexaScore-native
  code text,                               -- ISO 3166-1 for country, BCP 47 for language
  default_name text not null,
  external_ids jsonb not null default '{}', -- tmdb_person, tmdb_company, igdb_company, igdb_platform, openlibrary_author
  image_key text, image_source text references data_sources(code),
  born_on date, died_on date, founded_year smallint,
  provenance jsonb not null default '{}',
  created_at timestamptz not null default now(), updated_at timestamptz not null default now(),
  unique (kind, slug),
  unique (kind, code)
);

create table entity_translations (
  entity_id uuid not null references entities(id) on delete cascade,
  locale text not null check (locale in ('en','pt-BR','es')),
  name text not null, name_source text not null references data_sources(code),
  description text, description_source text references data_sources(code),
  description_licence text, description_source_url text,
  primary key (entity_id, locale)
);

create table item_entities (
  item_id uuid not null references items(id) on delete cascade,
  entity_id uuid not null references entities(id) on delete restrict,
  role text not null check (role in ('director','writer','cast','creator','author',
     'developer','publisher','production_company','network','platform','genre',
     'game_mode','franchise','series','country','original_language')),
  position smallint not null default 0,
  character_name text,                     -- cast only
  source text not null references data_sources(code),
  primary key (item_id, role, entity_id)
);
create index item_entities_entity_idx on item_entities (entity_id, role, item_id);

create table entity_links (               -- platform -> holder, imprint -> publisher, franchise -> parent
  from_id uuid not null references entities(id) on delete cascade,
  to_id uuid not null references entities(id) on delete cascade,
  relation text not null check (relation in ('platform_holder','imprint_of','part_of')),
  primary key (from_id, relation, to_id)
);

create table entity_stats (               -- hourly job, §7.4
  entity_id uuid not null references entities(id) on delete cascade,
  type text not null, item_count int not null, scored_count int not null,
  avg_score real, refreshed_at timestamptz not null,
  primary key (entity_id, type)
);

create table catalog_redirects (
  scope text not null check (scope in ('item','entity')),
  namespace text not null,                 -- item type or entity kind
  old_slug text, old_id uuid,
  target_id uuid not null,
  reason text not null check (reason in ('rename','merge')),
  created_at timestamptz not null default now(),
  check (old_slug is not null or old_id is not null)
);
create unique index catalog_redirects_slug_uq on catalog_redirects (scope, namespace, old_slug) where old_slug is not null;
create unique index catalog_redirects_id_uq on catalog_redirects (scope, old_id) where old_id is not null;

create table affiliate_links (            -- kept, optional (feature flag "affiliates", default off)
  item_id uuid not null references items(id) on delete cascade,
  store text not null check (store in ('amazon','mercadolivre','steam','gog')),
  url text not null, active boolean not null default true,
  primary key (item_id, store)
);
```

Role compatibility (enforced in Go on write and by DQ-08):

| Role | Entity kind | Item types |
| --- | --- | --- |
| director, writer, cast, creator | person | movie, series (books use `author`) |
| author | person | book |
| developer, publisher, production_company, network | organization | developer/publisher: game, book (publisher); production_company: movie, series; network: series |
| platform, game_mode | platform, game_mode | game |
| franchise, series | franchise | all |
| genre, country, original_language | genre, country, language | all |

Genres are shared across types (genre page has type tabs). The Wikidata property per role is SDD 06's mapping [NOT VERIFIED] here.

### 5.3 Invariants

- I-1: `items.slug` and `entities.slug` are never reused while a redirect with that `old_slug` exists in the same namespace.
- I-2: `catalog_redirects.target_id` always points at a live row; on a new merge, existing redirects to the source are rewritten to the new target (flatten).
- I-3: `primary_edition_id`, when set, belongs to the same work.
- I-4: A referenced entity cannot be deleted (`restrict`).

### 5.4 Migration notes from legacy (A-04)

| Legacy | Rewrite | Rule |
| --- | --- | --- |
| `items.id, type, slug, wikidata_qid, release_date, external_ids, created_at, updated_at` | same | 1:1; ids preserved so API ids and lists survive; `release_precision` = `day` when a date exists else `unknown`; `release_status` = `released` when `release_date <= today` else `announced`; `visibility` = `public` |
| `items.popularity`; `score*`, `rating_count` | `popularity_prior`; SDD 02 columns | rename; SDD 02 |
| `meta.genres, platforms, creators, countries` | `entities` + `item_entities` | resolve via QIDs in `source_records` payloads; unresolved strings become native entities (DQ-09); roleless `creators` get `creator`/`author`/`developer` by type until re-ingested |
| `meta.runtimeMinutes, pages, seasons, episodes, trailers` | typed columns, `trailers` | 1:1 |
| `meta.coverUrl, coverThumbUrl, coverAttribution` | `cover_key`, `cover_thumb_key`, `cover_source`, `cover_attribution` | Blob objects copied to the S3 bucket (SDD 11) under the legacy key `covers/<id>-<hash16>.webp` (06-4: no rewrite of ~80k objects; new covers use the same format, SDD 06) |
| book items | work + one edition per `external_ids.isbn13` | edition language/publisher from Open Library edition data when present, else `language` = `und`, publisher null, flagged by DQ-10; never blocks the migration (01-1) |
| `item_translations` | same + `title_source`, `synopsis_source` | 1:1; `synopsis_license` → `synopsis_licence`; source inferred from URL host |
| `external_scores` | dropped (D-09) | not copied; kept only in the final legacy `pg_dump` |
| `affiliate_links` | kept, optional | rows copied if any (PROGRESS.md notes `affiliate_store` values never written) |
| `source_records` | SDD 06 | — |

## 6. Contracts

### 6.1 Audiences and field visibility

Audience `site` = web Inertia props and first-party app API (`/app/v1`, A-01). Audience `api` = paid read API (SDD 05). `CATALOG_API_SOURCES` (default `wikidata,wikipedia,openlibrary,hexascore`) is loaded with vukem `config.Load` (pattern shown in https://vukem.dev/docs/digging-deeper/storage "S3"). Open Library's licence is pending A-05 [NOT VERIFIED]: until `docs/legal/open-library.md` exists, `openlibrary` is allowed in the API for identifiers and titles only (01-2), and it leaves the default if A-05 fails.

HexaScore-sourced fields (`hexascore` and `community` sources: scores, counts, trending, community edits, HexaScore ids and slugs) reach API customers under the API terms (A-07): use inside the customer's product, attribution "Data by HexaScore" with a link, no bulk redistribution, no model training without written permission; no open-data licence at launch. The provenance columns let a field be opened later without a migration.

| Field | site | api |
| --- | --- | --- |
| id, type, slug, wikidata QID, titles from allowed sources | yes | yes |
| synopsis | yes, with licence line | only if `synopsis_source` allowed; always with `licence` + `sourceUrl` |
| release date, runtime, counts | yes | only if `provenance[field]` allowed |
| cover, thumb | yes, with attribution | only if `cover_source` allowed |
| entity relations | yes | only rows whose `source` is allowed |
| external_ids | tmdb/igdb/steam/gog shown as links | only keys of allowed sources plus `wikidata`, `isbn13` |
| score, ratingCount, confidence | yes (SDD 02) | yes (SDD 02) |
| affiliate links | when flag on | never |
| popularity_prior | sort only | sort only |

Excluded fields are omitted (not `null`); `meta.omittedFields` is an array of the omitted field names (camelCase, same as SDD 05 §6.2) so buyers see coverage is partial (D-03 note).

### 6.2 Routes

| Op ID | Method, path | Auth | Output |
| --- | --- | --- | --- |
| `listItems` | GET `/api/v1/{type}` | API key | page of `ItemSummary` |
| `getItem` | GET `/api/v1/{type}/{id}` | API key | `ItemDetail`; 301 for merged id |
| `listEditions` | GET `/api/v1/books/{id}/editions` | API key | page of `Edition` |
| `listGenres` | GET `/api/v1/{type}/genres` | API key | genres with counts (legacy parity) |
| `getEntity` | GET `/api/v1/entities/{id}` | API key | `EntityDetail` |
| `listEntityItems` | GET `/api/v1/entities/{id}/items?type=&role=` | API key | page of `ItemSummary` |
| app equivalents | GET `/app/v1/...` same paths | access token | audience `site` |

Paths follow the README route-prefix convention (07-5): `/api/v1` in the app router, published as `https://api.hexascore.app/v1` by host routing (SDD 05, VF-015). Every route serves `public` items only (FR-01-14).

`{type}` path values are `games|movies|series|books` (`legacy/apps/web/lib/sections.ts`). Query: `genre`, `platform` (entity slugs), `year` (1800–2100), `min_score`/`max_score` (0–10, step 0.5), `min_<category>` (keys from `legacy/apps/web/lib/categories.ts`, A-06), `sort`, `limit`, `cursor`. RFC 9457 errors: `invalid-query` (400), `invalid-cursor` (400), `not-found` (404). Envelope per https://vukem.dev/docs/rest-apis/pagination "Envelopes" with `meta.next_cursor`/`prev_cursor`.

Inertia pages (SDD 09 owns layout): `Catalog/Index` props `{type, filters, items: ItemSummary[], cursor, total}`; `Catalog/Show` `{item: ItemDetail, editions?, ratingSummary}`; `Entity/Show` `{entity: EntityDetail, items, cursor}`. Web paths: `/{section}/{slug}`, `/people/{slug}`, `/companies/{slug}`, `/franchises/{slug}`, `/platforms/{slug}`, `/genres/{slug}`, `/game-modes/{slug}`, `/countries/{code}`, `/languages/{code}`, all with `/pt-br` and `/es` prefixes (D-17).

### 6.3 Item detail example (audience `api`, default sources)

```json
{
  "id": "0192f3a4-7b1c-7e21-9c55-3f1e2d4a6b70",
  "type": "book",
  "slug": "dom-casmurro",
  "wikidataQid": "Q1235781",
  "title": "Dom Casmurro",
  "titleLocale": "pt-BR",
  "originalTitle": "Dom Casmurro",
  "releaseDate": "1899-01-01",
  "synopsis": {
    "text": "Dom Casmurro é um romance de ...",
    "source": "wikipedia",
    "licence": "CC-BY-SA-4.0",
    "sourceUrl": "https://pt.wikipedia.org/wiki/Dom_Casmurro"
  },
  "cover": null,
  "credits": {
    "author": [{ "id": "0192f3a4-...", "kind": "person", "slug": "machado-de-assis", "name": "Machado de Assis" }]
  },
  "genres": [{ "id": "0192f3a4-...", "slug": "novel", "name": "Romance" }],
  "editions": { "total": 14, "brazilian": 9, "href": "/api/v1/books/0192f3a4-.../editions" },
  "externalIds": { "wikidata": "Q1235781", "openlibrary": "OL15843W" },
  "score": 8.4, "ratingCount": 132, "confidence": "high",
  "links": { "web": "https://hexascore.app/pt-br/books/dom-casmurro" },
  "meta": { "omittedFields": ["cover", "runtimeMinutes"], "sources": ["wikidata", "wikipedia", "openlibrary"] }
}
```

QID and Open Library id are illustrative [NOT VERIFIED]. `score`/`confidence`: SDD 02 (D-13).

## 7. Behaviour and algorithms

### 7.1 Slugs

Generated from the `en` title: NFKD, strip marks, lowercase, non-alphanumerics → `-`, trim to 120. Collision in the namespace: append release year, then `-2`, `-3`. Reserved per namespace: `genres`, `trending`, `search`, `new`. Slug changes (admin or SDD 04 approved edit) insert a `rename` redirect in the same transaction. Lookup order on a web request: live slug → redirect (301) → 404.

### 7.2 Listing and cursor

vukem `CursorPaginate` pages by primary key only and rejects `?sort` with a cursor (https://vukem.dev/docs/rest-apis/pagination "Query conventions"), so keyset paging over `(sort expression, id)` is built with `WhereRaw` (https://vukem.dev/docs/database/orm "Query builder") and an opaque cursor `{s, v, id, d}` signed with `signed.Signer` (https://vukem.dev/docs/digging-deeper/storage "Storage"). Composite-key cursors are [NOT IN VUKEM DOCS] (`vukem-feedback.md` candidate). Nulls sort last both ways (legacy `sortExpression`); score filters use SDD 02's visibility rule; `total` is counted on the first page only.

### 7.3 Read-model assembly

Item detail: 1 query item + translation (locale and `en` in one `IN`), 1 query `item_entities` joined to entities and translations, 1 query edition summary (books), SDD 02 score columns on the item row. Then an audience filter removes disallowed fields. Book cover: item cover, else `primary_edition_id` cover, else for `pt-BR` the newest Brazilian edition with a cover, else the newest edition with a cover. Editions list orders `is_brazilian desc` for every locale (D-10 "highlighted") and the UI shows a "Brazilian edition" badge.

### 7.4 Entity stats

Hourly scheduler job (https://vukem.dev/docs/digging-deeper/scheduler "Scheduler", BETA; D-08) upserts `entity_stats` in batches of 10k entities; `avg_score` = mean of visible item scores per type, shown when `scored_count >= 5`. Idempotent; restart reruns the batch.

### 7.5 Merge

Admin action, same type/kind, one transaction: `LockForUpdate` both rows ordered by id; 409 on self-merge or an already-redirected source; move relations, editions and user data (`on conflict do nothing`, target wins, legacy `mergeItems` rule); copy missing translations; union `external_ids`; insert `merge` redirects (id, slug) and rewrite redirects to the source; write the SDD 04 audit event carrying a jsonb snapshot of both items (row, translations, relations, editions, external ids) taken before the move, for manual restore (04-9; a reversible merge is post-launch); delete source; `tx.AfterCommit` enqueues SDD 02 recompute and SDD 07 reindex (https://vukem.dev/docs/database/relations "Query builder").

### 7.6 Data quality asserts

Each assert is a SQL query returning violating ids; `vukem catalog:check [--json]` is an app command (`vukem make:command`, https://vukem.dev/docs/cli/reference-make) that runs them, prints counts, and exits 1 on any `error`. Schedule: daily 04:00 UTC via the scheduler, results logged and alerted by SDD 12; also a CI step on the fixture DB and a mandatory gate after the legacy data copy (D-21).

| ID | Assert | Level |
| --- | --- | --- |
| DQ-01 | every item has an `en` translation with non-empty title | error |
| DQ-02 | every `editions.isbn13` passes `isbn13_is_valid` and is unique | error |
| DQ-03 | no orphan entities: zero `item_entities`, editions and `entity_links` refs, older than 7 days | warn |
| DQ-04 | slugs match the regex; no live slug equals a redirect `old_slug` in its namespace | error |
| DQ-05 | redirect targets exist and are not themselves redirect sources (no chains) | error |
| DQ-06 | synopsis present ⇒ source, licence and (for Wikipedia) source URL present | error |
| DQ-07 | `cover_key` present ⇒ `cover_source` present; covers from licences needing attribution have `cover_attribution` | error |
| DQ-08 | `item_entities` role compatible with entity kind and item type (§5.2 table) | error |
| DQ-09 | native entities (`wikidata_qid null`) count, listed for review | warn |
| DQ-10 | book works without editions; editions with defaulted language | warn |
| DQ-11 | `provenance` values and every `*_source` exist in `data_sources` | error |
| DQ-12 | `entity_stats.refreshed_at` older than 2 h | warn |
| DQ-13 | `released` item with null `release_date` and precision not `unknown`; `announced` item with a `day` date more than 30 days past | warn |
| DQ-14 | `item_links` 2-cycles (`a based_on b` and `b based_on a`) | warn |

## 8. Edge cases

- No `pt-BR` title: `en` shown, `titleLocale: "en"`.
- Two works share an ISBN: unique index rejects; SDD 06 logs and skips.
- Edition ISBN-10 only: converted to ISBN-13 (978 prefix + recomputed check) by SDD 06; stored only as ISBN-13.
- One person or organisation in two roles on one item: two rows; entity page groups by role.
- Slug renamed twice: both old slugs redirect directly to the live slug (flatten).
- Merged item id requested in the API: 301 with `Location`; in a list of ids (lookup) the target id is returned.
- Default config and a TMDB-only runtime: field omitted in API, listed in `omittedFields`, shown on site with TMDB attribution.
- Entity left without items: DQ-03 flags it; deletion is manual.
- Cursor from `-score` used with `sort=-popularity`: 400 `invalid-cursor`.
- `min_score=7.3`: 400 (step 0.5).

## 9. Acceptance criteria

- [ ] `vukem migrate` on empty Postgres 18 → exit 0; `vukem migrate:status --json` lists every catalog migration as ran.
- [ ] `psql -c "select isbn13_is_valid('9780306406157'), isbn13_is_valid('9780306406158')"` → `t | f`.
- [ ] `go test ./app/catalog/... -run 'TestMerge|TestRedirect|TestVisibility|TestCursor'` → PASS.
- [ ] `go test -tags=integration ./app/catalog/...` with `DATABASE_URL` set → PASS (FR-01-1..13).
- [ ] With default `CATALOG_API_SOURCES`, `curl -H "Authorization: Bearer $KEY" $API/v1/movies/$ID | jq '..|.source? // empty' | sort -u` → no `tmdb`, no `igdb`.
- [ ] `curl -sI $WEB/pt-br/games/$OLD_SLUG` → `301`, `Location: /pt-br/games/$NEW_SLUG`.
- [ ] `vukem catalog:check --json` on the migrated staging copy → `{"errors":0}`; warnings recorded in the cutover checklist.
- [ ] `curl $WEB/sitemap.xml` → index whose child files each hold ≤ 50,000 URLs with `xhtml:link` alternates for `en`, `pt-br`, `es`, `x-default`.
- [ ] Schemathesis on `listItems`, `getItem`, `getEntity`, `listEntityItems`, `listEditions` → no failures.

## 10. Test strategy

- Unit: slugs, locale resolver, cursor, audience filter, role compatibility, ISBN.
- Integration (Postgres, `-tags=integration`, https://vukem.dev/docs/database/factories-seeders "PostgreSQL"; fixtures via "Factories and seeders"): constraints, merge, redirects, filters, sorts, entity stats, DQ asserts.
- Contract: vukem OpenAPI (SDD 05) + Schemathesis; visibility test with TMDB/IGDB fixtures. E2E: Playwright + axe (SDD 09). Load: k6 (NFR-01-1).
- Migration: run on a legacy dump; compare counts per type (items, translations, covers); DQ gate.

## 11. Open issues

1. ISBN-only legacy books lack edition language/publisher. Decided (01-1): Open Library edition data when present, else language `und`, publisher null, DQ-10; never block the migration. Applied in §5.4.
2. Open Library licence (A-05) decides whether `openlibrary` stays in the default `CATALOG_API_SOURCES`. Owner: architect. Closes with the A-05 licence check (`docs/legal/open-library.md`); until then identifiers and titles only (§6.1).
3. Engines (D-01 proposed list), translators/illustrators on editions and episode data. Decided: after v1.0.0 (D-30).
4. Composite-key cursor pagination is not in vukem docs; built in the app with `WhereRaw`. Owner: vukem. Closes when `vukem-feedback.md` entry ships or is rejected.
5. Postgres mapping of `t.JSON` (json vs jsonb) and `t.Enum` is undocumented; DDL uses `s.Raw` and text + CHECK. Owner: vukem. Closes when vukem docs state the mapping.
6. Whether TMDB/IGDB display terms allow showing their fields in the mobile app (audience `site`) is [NOT VERIFIED]. Owner: architect. Decided (01-6): the app is the same service as the site (display, not redistribution), attribution in the item and about screens; the owner still owes the Twitch agreement reading (D-36).
7. Licence for HexaScore's own data. Decided: A-07 (§6.1).
8. No sitemap helper in vukem [NOT IN VUKEM DOCS]; an app handler generates it; helper filed as VF-007. Threshold decided (01-8, 09-8): ≥ 3 linked items, others `noindex,follow` (FR-01-12); revisited with Search Console data 60 days after launch.
