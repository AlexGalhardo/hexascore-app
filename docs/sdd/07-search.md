# SDD 07: Search

Status: reviewed (2026-10-28 consistency pass). Decisions: D-01, D-08, D-10, D-16, D-17, A-01. ADRs: 0003, 0006, 0011. Related SDDs: 01, 05, 06, 08, 09, 10, 12, 13.

## 1. Objective

Find any item (game, movie, series, book work or edition) or entity (person, organisation, franchise) in under a second from any locale, tolerating typos, missing accents and transliterated titles, with autocomplete while typing. Search is the main discovery path after trending (SDD 08). It runs in PostgreSQL (ADR 0003, ADR 0011) behind one Go interface so an engine swap touches one package.

## 2. Scope

In (v1): the `search` package and interface; per-locale FTS; trigram on item titles, alternative titles, edition titles, entity names; ranking; type filter and grouped results; entity block; suggest; offset pagination; caching; did-you-mean; analytics; golden set; k6 budgets; web, app and paid-API contracts.

Out: facet counts (ADR 0011 "Negative"); vector search; personalised ranking (A-02); search in lists/reviews; external engines (deferred, ADR 0011); trending/similar (SDD 08); listings and their cursor (SDD 01 §7.2).

## 3. Functional requirements

- FR-07-1: All search SQL lives in `internal/search`; no other package references `search_tsv`, `*_norm`, `search_suggest` or pg_trgm operators. CI grep test `TestSearchSQLConfined`.
- FR-07-2: `Search` orders hits by §7.3. Integration test + golden set (§10).
- FR-07-3: Websearch syntax (`"phrase"`, `-exclude`, `or`) works: `"the last of us" -part` excludes titles with "part". Integration test.
- FR-07-4: `acao`, `AÇÃO`, `ação` return the same top 10. Integration test.
- FR-07-5: `zleda` returns "The Legend of Zelda" in top 3 (legacy parity, `legacy/qa/e2e/specs/search.spec.ts` lines 50–55). Golden set.
- FR-07-6: Without a type filter the web page shows one group per type (games, movies, series, books), ≤ 8 hits each, plus ≤ 6 entities (legacy grouping, `legacy/apps/web/app/[locale]/search/page.tsx` lines 30–36). Playwright.
- FR-07-7: With a type filter: one list, 20 per page, ≤ 200 results reachable; beyond → 400 `invalid-query`. Integration test.
- FR-07-8: Editions match by title and exact ISBN-13; the work is returned with `matchedEdition` (D-10). Integration test.
- FR-07-9: Alternative/transliterated titles match: `gojira` finds "Godzilla". Golden set.
- FR-07-10: `Suggest` returns ≤ 8 entries (≤ 6 items, ≤ 2 entities) for ≥ 2 characters, prefix before trigram. Integration test.
- FR-07-11: Zero results with ≥ 3 characters return `didYouMean` or null. Integration test.
- FR-07-12: Searches and clicks emit events (§7.8); no raw query older than 30 days survives. Integration test with fake clock.
- FR-07-13: Paid API output obeys ADR 0006 through SDD 01's audience filter. Schemathesis + unit test.

## 4. Non-functional requirements

- NFR-07-1: `/search` p95 ≤ 300 ms at 20 rps, 60 s, one instance, cache off, 200 distinct queries (legacy budget `legacy/qa/k6/api.js` lines 26–27). k6.
- NFR-07-2: `/search/suggest` p95 ≤ 50 ms at 50 rps, cache off. k6.
- NFR-07-3: `SET LOCAL statement_timeout` 800 ms (search), 150 ms (suggest).
- NFR-07-4: Catalogue changes reach suggest within 2 min; full search on commit (trigger-maintained `search_tsv`, open issue 07-1).
- NFR-07-5: Search indexes ≤ 600 MB at 100k items + 200k editions + 300k entities.
- NFR-07-6: Input 2–100 runes, ≤ 10 terms, control characters → space (legacy `legacy/apps/web/lib/catalog/repository.ts` lines 305–316); hostile input renders escaped.
- NFR-07-7: Suggest is a keyboard-operable ARIA combobox; axe 0 serious violations on `Search/Index`.
- NFR-07-8: Query logs hold no IP; rows deleted at 30 days; aggregates only for queries seen ≥ 5 times a day (LGPD, SDD 13).
- NFR-07-9: Search pages are SSR (D-07) and `noindex, follow`.

## 5. Data model

Requires Postgres ≥ 15 (`^@` on SP-GiST, `starts_with`); vukem's suite runs `postgres:17` (<https://vukem.dev/docs/database/migrations> "PostgreSQL"); Railway's version is [NOT VERIFIED]. All DDL is `s.Raw` (<https://vukem.dev/docs/database/migrations> "Migrations"). Each migration runs in its own transaction, so `CREATE INDEX CONCURRENTLY` cannot run in one; non-transactional migrations are [NOT IN VUKEM DOCS] (`vukem-feedback.md` candidate). At launch size a plain `CREATE INDEX` locks for seconds, accepted in the cutover window (SDD 11). Extensions exist in legacy (`legacy/apps/web/lib/db/migrations/0000_extensions.sql`); `unaccent` was installed but unused.

```sql
create extension if not exists pg_trgm;
create extension if not exists unaccent;

-- unaccent() is STABLE; generated columns and index expressions need IMMUTABLE.
create function hs_unaccent(text) returns text language sql immutable parallel safe strict
  as $$ select public.unaccent('public.unaccent'::regdictionary, $1) $$;
-- Single normaliser; Go never normalises, SQL calls hs_norm($1).
create function hs_norm(text) returns text language sql immutable parallel safe strict
  as $$ select btrim(regexp_replace(lower(hs_unaccent($1)), '[[:punct:][:space:]]+', ' ', 'g')) $$;

create text search configuration hs_en (copy = pg_catalog.english);
alter text search configuration hs_en
  alter mapping for hword, hword_part, word with unaccent, english_stem;
create text search configuration hs_pt (copy = pg_catalog.portuguese);
alter text search configuration hs_pt
  alter mapping for hword, hword_part, word with unaccent, portuguese_stem;
create text search configuration hs_es (copy = pg_catalog.spanish);
alter text search configuration hs_es
  alter mapping for hword, hword_part, word with unaccent, spanish_stem;
```

`asciiword` keeps the stock mapping. Configs are replaced by a new name (`hs_pt_v2`), never altered in place: altering a mapping does not rewrite stored vectors.

Columns on SDD 01 tables (SDD 01 §5.2 delegates them here):

```sql
alter table item_translations
  add column title_norm text generated always as (hs_norm(title)) stored,
  add column search_tsv tsvector not null default '',      -- site: title A + any synopsis D
  add column search_tsv_api tsvector not null default '';  -- api: allowlisted-source text only (07-7)
-- Trigger-maintained, not generated (07-1): CASE over regconfig may fail the immutability check.
create function item_translations_tsv() returns trigger language plpgsql as $$
declare cfg regconfig := case new.locale when 'en' then 'hs_en' when 'pt-BR' then 'hs_pt' else 'hs_es' end;
begin
  new.search_tsv := setweight(to_tsvector(cfg, new.title), 'A')
                 || setweight(to_tsvector(cfg, coalesce(new.synopsis, '')), 'D');
  new.search_tsv_api := setweight(to_tsvector(cfg, new.title), 'A')     -- titles come from allowed sources
                 || setweight(to_tsvector(cfg, case when new.synopsis_source = 'wikipedia'
                                                    then coalesce(new.synopsis, '') else '' end), 'D');
  return new;
end $$;
create trigger item_translations_tsv_trg before insert or update of title, synopsis, synopsis_source, locale
  on item_translations for each row execute function item_translations_tsv();
create index item_translations_tsv_gin on item_translations using gin (search_tsv);
create index item_translations_tsv_api_gin on item_translations using gin (search_tsv_api);
-- Adding a synopsis source to CATALOG_API_SOURCES means a migration that edits the CASE above and runs search:rebuild.
create index item_translations_title_trgm on item_translations using gin (title_norm gin_trgm_ops);

-- item_alt_titles is owned by SDD 01 §5.2 (07-6) and filled by SDD 06 (FR-06-11); search adds:
alter table item_alt_titles add column title_norm text generated always as (hs_norm(title)) stored;
create unique index item_alt_titles_norm_uq on item_alt_titles (item_id, title_norm);
create index item_alt_titles_trgm on item_alt_titles using gin (title_norm gin_trgm_ops);

alter table editions add column title_norm text generated always as (hs_norm(title)) stored;
create index editions_title_trgm on editions using gin (title_norm gin_trgm_ops);
alter table entity_translations add column name_norm text generated always as (hs_norm(name)) stored;
create index entity_translations_name_trgm on entity_translations using gin (name_norm gin_trgm_ops);
alter table entities add column name_norm text generated always as (hs_norm(default_name)) stored;
create index entities_name_trgm on entities using gin (name_norm gin_trgm_ops)
  where kind in ('person','organization','franchise');

create table search_suggest (
  ref_id uuid not null,
  kind text not null check (kind in ('game','movie','series','book','person','organization','franchise')),
  locale text not null check (locale in ('en','pt-BR','es','*')),  -- '*' = alternative title
  label text not null check (length(label) <= 300),
  label_norm text not null,
  slug text not null, year smallint,
  thumb_key text,                          -- items.cover_thumb_key / entities.image_key
  popularity real not null default 0,
  primary key (ref_id, locale, label_norm)
);
create index search_suggest_prefix on search_suggest using spgist (label_norm);
create index search_suggest_trgm on search_suggest using gin (label_norm gin_trgm_ops);

create table search_dirty (ref_id uuid primary key, kind text not null,
  queued_at timestamptz not null default now());  -- AFTER triggers on the five source tables
```

Decided (07-1): the tsvectors are trigger-maintained plain columns from the first migration; no generated tsvector column. `title_norm`/`name_norm` stay generated because `hs_norm` is an `IMMUTABLE` wrapper.

Index builds on populated tables use `CREATE INDEX CONCURRENTLY`, which cannot run inside a migration transaction [NOT IN VUKEM DOCS] (VF-017): migrations create indexes while the tables are still near-empty before cutover; later concurrent builds run as a one-off pre-deploy script (`hexascore:index-concurrently`, SDD 11 pre-deploy step) outside the migration runner, idempotent via `IF NOT EXISTS` and a check for an `INVALID` index left by a failed build.

Analytics: `search_events(id uuid pk, at, search_id, surface web|app|suggest, query_norm, user_id null, locale, types text[], result_count, latency_ms, page, clicked_ref, click_position, click_group)` indexed on `at`; `search_query_daily(day, query_norm, locale, searches, zero_results, clicks)` pk `(day, query_norm, locale)`.

Legacy migration: drop `item_translations.search` (`'simple'` config, `legacy/apps/web/lib/db/migrations/0001_init.sql` line 108) and indexes (lines 327–328) after the new columns exist; `search:rebuild` fills `search_suggest` after the A-04 copy.

## 6. Contracts

```go
type Searcher interface {
	Search(ctx context.Context, q Query) (Results, error)
	Suggest(ctx context.Context, q SuggestQuery) ([]Suggestion, error)
	SearchEntities(ctx context.Context, q EntityQuery) (EntityResults, error)
}
// Audience "api" ranks only on allowlisted-source text: search_tsv_api, titles from allowed sources and
// item_alt_titles rows whose source is allowed; "site" uses search_tsv and every alt title (07-7).
type Query struct {
	Text     string
	Locale   string   // "en" | "pt-BR" | "es" | "" = all
	Types    []string // empty = all, grouped
	Page     int      // 1..10
	PerPage  int      // site 20; API 1..50
	Audience string   // "site" | "api" (SDD 01 §6.1)
}
```

Input is parsed with `web.Parse(c, schema)`; schema failures on JSON routes are `422 application/problem+json` (<https://vukem.dev/docs/basics/validation> "Validation"; legacy answered 400). Page beyond 200 results is 400 `invalid-query` (RFC 9457).

| Op ID | Route | Auth | Output |
| --- | --- | --- | --- |
| web | GET `/search?q=&type=&locale=&page=` (+ `/pt-br`, `/es`, D-17) | none | Inertia `Search/Index` |
| web | GET `/search/suggest?q=&type=&locale=` | none | `{data: Suggestion[]}`; legacy `/api/search/suggest` → 308 |
| web | POST `/search/events` | CSRF | 204 |
| app | GET `/app/v1/search`, `/app/v1/search/suggest`; POST `/app/v1/search/events` | optional token (A-01) | audience `site` |
| `searchItems` | GET `/api/v1/search?q=&type=&locale=&limit=&page=` | API key | `{data: ItemSummary[], meta}` |
| `searchEntities` | GET `/api/v1/search/entities?q=&kind=&locale=&limit=&page=` | API key | `{data: EntitySummary[], meta}` |

`Search/Index` props: `query {q, type, locale, page}`, `groups [{type, items: SearchHit[], total, moreHref}]` (no type), `results` as `inertia.Scroll(v, meta)` with partial reload `only: ['results']` (type set; <https://vukem.dev/docs/frontend/inertia> "Inertia v3 (BETA)", Props and Requests rows), `entities EntityHit[]`, `didYouMean`, `searchId`, `tooShort`. The `Scroll` meta for offset pages not built by `orm` is [NOT VERIFIED]. `SearchHit` = SDD 01 `ItemSummary` + `matchedTitle` (other-locale or alternative title), `matchedEdition {id, title, isbn13}`, optional `highlight`.

`Suggestion` = legacy shape (`legacy/apps/web/app/api/search/suggest/route.ts`) `{id, type, title, year, coverThumbUrl, path}` + `kind`; `score` dropped (SDD 02 owns live scores).

API: `limit` 1–50 (legacy `SEARCH_MAX`, `legacy/apps/web/lib/api/schemas.ts` line 303), `page·limit ≤ 200`. Displayed titles come only from allowed sources, falling back to another allowed-source title; a hit with none is dropped (ADR 0006 §1).

## 7. Behaviour and algorithms

### 7.1 Query parsing

1. `\p{Cc}` → space, NFC, trim, collapse; > 100 runes → 400; < 2 → `tooShort`, empty 200.
2. Raw text goes to `websearch_to_tsquery(config, $1)`, which never raises on bad syntax. An empty tsquery (stop words only, e.g. "the") still runs the trigram/exact branches.
3. Trigram/exact text = query minus quotes and `-term` tokens; empty → zero results.
4. Text matching `^97[89][0-9]{10}$` after stripping hyphens/spaces: exact `editions.isbn13` lookup, its work at rank 1.

### 7.2 Locale

Search locale = `locale` param, else UI locale (URL prefix on web, `Accept-Language` on app), else all. With a locale, FTS runs on that locale's rows with its config; "all" OR-s three branches. Trigram/exact/prefix always span all locales, alternative and edition titles, so a pt-BR user typing an English title finds it. Display title: UI locale, fallback `en`.

### 7.3 Ranking

Per item, over all candidate rows:

- `E` = 1 if any `norm = hs_norm(q)`; `P` = 1 if any `starts_with(norm, hs_norm(q))`
- `S` = max `word_similarity`; θ = 0.2 (legacy `WORD_SIMILARITY_THRESHOLD`, `repository.ts` lines 302–303, set per transaction, line 321)
- `F` = max `ts_rank_cd(search_tsv, tsq, 32)` ∈ [0, 1)
- `Pop` = `least(ln(1 + popularity_prior) / ln(1 + p99), 1)`, p99 cached daily

```text
rank = 10·E + 5·P + (S ≥ θ ? 2 + S : 0) + F + 0.3·Pop
```

Tiers are strict: non-exact max 9.3 < 10; non-prefix max 4.3 < 5; FTS-only max 1.3 < 2.2. Ties: `popularity_prior desc, release_date desc nulls last, id`. Legacy ranked by `greatest(word_similarity, ts_rank)` without exact/prefix tiers (`repository.ts` lines 323–347).

```sql
-- $1 raw q; $2 types text[] or null; $3 p99; $4 limit; $5 offset (locale pt-BR shown)
select set_config('pg_trgm.word_similarity_threshold', '0.2', true);
with c as (
  select item_id, title_norm n, ts_rank_cd(search_tsv, websearch_to_tsquery('hs_pt', $1), 32) f
    from item_translations
   where locale = 'pt-BR' and search_tsv @@ websearch_to_tsquery('hs_pt', $1)
  union all select item_id, title_norm, 0 from item_translations where hs_norm($1) <% title_norm
  union all select item_id, title_norm, 0 from item_alt_titles where hs_norm($1) <% title_norm
  union all select work_id, title_norm, 0 from editions where hs_norm($1) <% title_norm
), s as (
  select item_id, bool_or(n = hs_norm($1)) e, bool_or(starts_with(n, hs_norm($1))) p,
         max(word_similarity(hs_norm($1), n)) sim, max(f) f
    from c group by item_id
)
select i.id, 10 * e::int + 5 * p::int + case when sim >= 0.2 then 2 + sim else 0 end + f
       + 0.3 * least(ln(1 + i.popularity_prior) / ln(1 + $3), 1) as rank
  from s join items i on i.id = s.item_id
 where $2::text[] is null or i.type = any($2)
 order by rank desc, i.popularity_prior desc, i.release_date desc nulls last, i.id
 limit $4 offset $5;
```

Grouped mode adds `row_number() over (partition by i.type …) <= 8` and `count(*) over (partition by i.type)` for `total`. SQL is written with `?` and `database.Rebind`, run via `db.QueryContext` (<https://vukem.dev/docs/database/orm> "Query builder"). Trigram branches are skipped when the normalised query is under 3 characters.

### 7.4 Entities

Same E/P/S tiers (no FTS) over `entity_translations.name_norm` and `entities.name_norm`, kinds person/organization/franchise; tie-break `entity_stats.item_count` (SDD 01 §7.4).

### 7.5 Suggest

Query 1: `label_norm ^@ hs_norm($1) and locale in ($ui, '*')`, `order by popularity desc limit 24`. Under 8 distinct refs: query 2 with `label_norm % hs_norm($1)` (similarity threshold 0.3), `order by similarity desc, popularity desc limit 24`. Go dedupes by `ref_id` (UI-locale label preferred), caps 6 items + 2 entities. Job `search:suggest-sync` (`EveryMinute`, `WithoutOverlapping`, <https://vukem.dev/docs/digging-deeper/scheduler> "Scheduler") drains ≤ 5,000 `search_dirty` rows with `FOR UPDATE SKIP LOCKED` (<https://vukem.dev/docs/database/orm> "Lock" row), upserts/deletes suggest rows and the dirty rows in one transaction; idempotent. `search:rebuild` builds `search_suggest_new` and swaps by rename; weekly and after backfills (SDD 06).

### 7.6 Caching

Anonymous requests only (no session, no API key). `cache.NewMemory(10000)` with `cache.Remember` (<https://vukem.dev/docs/digging-deeper/cache> "Cache", singleflight on misses); memory is loss-tolerant (ADR 0003). Key `search:v1:{epoch}:{sha256(text|locale|types|page|perPage|audience)}`; `epoch` is bumped when the suggest sync applies changes. TTL 120 s search, 300 s suggest. `ETag` via `web.ETagOf(v)` and `c.NotModified` (<https://vukem.dev/docs/rest-apis/overview> "Cross-cutting", Caching row), `Cache-Control: public, max-age=60` (legacy suggest header); logged-in requests get `private, no-store`.

### 7.7 Highlighting and zero results

Titles are highlighted client-side. `ts_headline` snippets only with `highlight=1` on web, for the 20 hits on the page; it re-parses each document (0.2–1 ms per row [NOT VERIFIED]). Zero results: `didYouMean` = top suggest label with `similarity ≥ 0.3`, popularity tie-break; the page also shows trending (SDD 08).

### 7.8 Analytics and privacy

Events `search.performed` and `search.clicked` (position 1-based within its group). After the click is stored, the in-process event `SearchResultClicked{ItemID uuid, QueryHash string /* sha256(query_norm|locale), hex */, Position int /* 1-based */, Locale string, At time.Time}` is dispatched (<https://vukem.dev/docs/digging-deeper/events>); SDD 08 consumes it by this name (open issue 08-5). It never carries the query text or a user id. Web clicks via `navigator.sendBeacon` to POST `/search/events`. Events buffer in memory and flush every 10 s with `orm.CreateMany`; a crash loses ≤ 10 s. Paid-API queries are counted (SDD 05), never logged with text. E-mail, phone and CPF patterns are stored as `[redacted]`. Daily job: aggregate into `search_query_daily` (rows with `searches ≥ 5`, kept 13 months), delete `search_events` older than 30 days. Consumers: SDD 08 (clicks), SDD 12 (zero-result rate).

### 7.9 Accents, pt-BR, transliteration

"ação", "Ação", "acao" all normalise to `acao`, and the FTS configs unaccent before stemming, so they share lexemes. The Portuguese snowball stemmer has accented-suffix rules (`-ção`), so unaccenting first may weaken "ação"/"ações" conflation.

Decided (07-2, P1.E.2.2, measured on Postgres 18.6): `hs_pt` keeps `unaccent` before `portuguese_stem`; only the unaccented form is indexed, no second vector. `to_tsvector('hs_pt', w)` against stock `portuguese`: `ação` → `'aca'` vs `'açã'`, `acao` → `'aca'` vs `'aca'` (stock never conflates the unaccented spelling users type); `ações` → `'aco'`, `acoes` → `'aco'`. The cost is `-ção`/`-ções` plural conflation: `informação` and `informações` share the stock lexeme `inform`, but are `informaca` and `informaco` under `hs_pt`; `ação` (`aca`) and `ações` (`aco`) differ in both configs. The trigram tier covers the gap: `word_similarity('acoes','acao explosiva') = 0.33`, `('coracoes','coracao valente') = 0.56`, `('informacoes','informacao') = 0.67`, all above θ = 0.2, so plural variants still match, ranked in the S tier below exact and prefix hits. Side effect seen in the golden set: query `ações` also ranks `Corações de Ferro` (substring `acoes`) above `Ação Explosiva`, so that line only requires the exact title `Ações Secretas`. Spanish stems `accion` and `acciones` to the same lexeme, so `acciones` finds `Acción mutante` through FTS. Non-Latin Non-Latin titles: SDD 06 writes Wikidata romanisations and aliases into `item_alt_titles` (`kind = transliteration`, `lang` with `-Latn`); the properties are SDD 06's mapping [NOT VERIFIED]. Native-script queries match only via trigram; CJK trigram behaviour is [NOT VERIFIED].

### 7.10 Sizing and maintenance

Estimate [NOT VERIFIED]: translations 220k rows (tsv GIN ~90 MB, trigram ~20 MB), alternative titles 150k (~15 MB), editions 200k (~20 MB), entity names 400k (~35 MB), suggest 650k rows (heap ~90 MB, SP-GiST ~40 MB, trigram ~50 MB): ≈ 360 MB. `ANALYZE` touched tables after each ingestion batch (SDD 06); `autovacuum_vacuum_scale_factor = 0.05` on `item_translations`, `search_suggest`, `search_dirty`; `gin_pending_list_limit = 4MB` on trigram indexes; console command `search:reindex` runs `REINDEX INDEX CONCURRENTLY` outside migrations monthly or when an index exceeds 1.5× its post-rebuild size (SDD 12 tracks size).

## 8. Edge cases

- Stop words only ("the"): FTS empty, exact tier still puts "The Thing"-style exact titles first.
- Only exclusions (`-zelda`): zero results, no did-you-mean.
- `%`, `_`, `\`: never reach `LIKE`; `^@` and `<%` take values.
- 100 identical titles ("Hamlet"): popularity then year order them.
- Merge (SDD 01 §7.5): both ids queued dirty; cache stale ≤ 120 s.
- Edition hit without a work translation in the UI locale: work in `en`, `matchedEdition.title` in the edition language.
- Empty page within the 200 cap: empty `results`, 200.
- Timeout: search 503 with `Retry-After: 2`; suggest returns `{data: []}` 200.

## 9. Acceptance criteria

- [ ] `psql "$DATABASE_URL" -c "select to_tsvector('hs_pt','Ação') = to_tsvector('hs_pt','acao')"` → `t`.
- [ ] `psql ... -c "select hs_norm('Pokémon: Edição Ouro!')"` → `pokemon edicao ouro`.
- [ ] `EXPLAIN` of `search_tsv @@ websearch_to_tsquery('hs_en','zelda')`, `hs_norm('zleda') <% title_norm` and `label_norm ^@ 'zel'` shows `Bitmap Index Scan` / `Index Scan` on `item_translations_tsv_gin`, `item_translations_title_trgm`, `search_suggest_prefix`.
- [ ] `go test -tags=integration ./internal/search/... -run 'TestGolden|TestSearchSQLConfined'` passes.
- [ ] `curl -s 'localhost:8080/search/suggest?q=zel' | jq '.data|length'` ≤ 8; first title contains "Zelda".
- [ ] `/api/v1/search?q=zelda&page=11&limit=20` with a key → 400 `invalid-query`.
- [ ] Anonymous `/search?q=zelda` repeated with `If-None-Match` → 304.
- [ ] `SEARCH_CACHE=off k6 run qa/k6/search.js` → `p(95)<300` (search, 20 rps) and `p(95)<50` (suggest, 50 rps) pass.
- [ ] `select pg_size_pretty(sum(pg_relation_size(indexrelid))) from pg_index where indexrelid::regclass::text ~ '(tsv|trgm|suggest)'` ≤ 600 MB on the full catalogue.
- [ ] An event at `now() - 31 days` is gone after `search:purge`.

## 10. Test strategy

- Unit: parser (control chars, lengths, ISBN, exclusion-only), cache key, redaction, audience filter.
- Integration (Postgres 17, `-tags=integration`): migrations up/down, generated and trigger-maintained columns (site vs api vectors), tier order with crafted fixtures, suggest sync with two concurrent runners.
- Relevance: `internal/search/testdata/golden.yaml`, 60 queries (20 per locale) over a 3k-item fixture from the migrated catalogue, ≤ 3 expected slugs each, all required in the top 3; any miss fails CI. Seeds (fixture `internal/search/testdata/fixture.sql`: 82 curated items plus generated filler to 3,000 items): `zleda`, `acao`, `ações`, `o senhor dos aneis`, `senor de los anillos`, `"the last of us" -part`, `gojira`, `dom casmurro`, `machado de assis` (entity top 1), `harry potter` (franchise top 1), one ISBN-13. Relevance bugs add a golden line first.
- Contract: Schemathesis on `searchItems`, `searchEntities` (SDD 05).
- E2E: Playwright + axe, port of `legacy/qa/e2e/specs/search.spec.ts` plus grouped page and scroll.
- Load: `qa/k6/search.js`, 200 queries including legacy `["the","star","love","war","zelda"]` (`legacy/qa/k6/api.js` line 58).
- Security: `TestSecurity_*` for injection and oversized `q` (SDD 13).

## 11. Open issues

1. `CASE` over `regconfig` in a generated column may be rejected. Decided (07-1): trigger-maintained column from the start (§5).
2. `unaccent` before `portuguese_stem` may weaken conflation. Decided and closed (07-2, §7.9): keep the order; the pt-BR golden queries pass.
3. Non-transactional migrations (`CREATE INDEX CONCURRENTLY`) are [NOT IN VUKEM DOCS]. VF-017; workaround: pre-deploy script (§5). Owner: vukem.
4. `inertia.Scroll` meta for non-`orm` offset pages is [NOT VERIFIED]. Owner: vukem. Closes when SDD 09 ships a working page.
5. Route prefix. Decided (07-5): README convention `/api/v1`; SDD 01 §6.2 aligned.
6. `item_alt_titles`. Decided (07-6): owned by SDD 01 §5.2, filled by SDD 06 FR-06-11.
7. API search over third-party text. Decided (07-7): API ranks on allowlisted-source text only (`search_tsv_api`), site on all text.
8. Suggest 50 ms p95 for 2-character prefixes over 650k rows is [NOT VERIFIED]. Owner: architect. Closes with the k6 run; fallback is a 3-character prefix minimum.
9. Railway Postgres version (≥ 15) and index sizes are [NOT VERIFIED]. Owner: architect (SDD 11). Closes with the provisioned version and the §9 size query.
10. Transposition typos sit exactly at θ: `word_similarity('zleda','zelda') = 0.200` while `('zleda','cien anos de soledad') = 0.333`, so in the golden set `zleda` returns `Cien años de soledad` among the top 3 and `Zelda II` falls out; the golden line requires only the two most popular Zelda titles. Candidate fixes (blend `similarity()`, `fuzzystrmatch` `levenshtein`) need an extension decision. Owner: architect.
