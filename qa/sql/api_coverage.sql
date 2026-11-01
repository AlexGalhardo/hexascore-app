-- Per-field API coverage (SDD 05 section 6.2 "Coverage", open issue 05-11, D-33): for each item type and each
-- field the paid API may return, the share of public items whose value comes from an allowed source, so it
-- reaches customers. The rules mirror the serialiser (app/catalog/audience.go Policy, api_items.go loader,
-- English row); TestAPICoverageSQL compares both on fixtures, so a mismatch is a bug in one of them.
--
-- Run on the staging copy with a read-only role and save the CSV for the site (P3.B.3.2):
--   psql "$STAGING_DATABASE_URL" --csv -f qa/sql/api_coverage.sql > /tmp/api_coverage.csv
--   bun scripts/api-coverage-json.ts /tmp/api_coverage.csv > resources/data/api_coverage.json
-- The allowlist defaults to config.DefaultCatalogAPISources; when production sets CATALOG_API_SOURCES, pass the
-- same value: PGOPTIONS='-c hexascore.api_sources=wikidata,hexascore' psql ...
WITH allowed AS (
  SELECT
    string_to_array(coalesce(nullif(current_setting('hexascore.api_sources', true), ''),
      'wikidata,wikipedia,openlibrary,hexascore,community'), ',') AS api,
    -- catalog.limitedSources: identifiers and titles only until the Open Library licence note (A-05).
    ARRAY['openlibrary'] AS limited
),
public_items AS (
  SELECT
    i.id,
    i.type,
    i.external_ids,
    CASE WHEN te.title IS NOT NULL THEN te.title_source
      WHEN i.original_title IS NOT NULL THEN i.provenance ->> 'original_title' ELSE 'hexascore' END AS title_src,
    CASE WHEN te.synopsis IS NOT NULL THEN te.synopsis_source END AS synopsis_src,
    CASE WHEN i.release_date IS NOT NULL THEN i.provenance ->> 'release_date' END AS release_src,
    CASE WHEN i.runtime_minutes IS NOT NULL THEN i.provenance ->> 'runtime_minutes' END AS runtime_src,
    CASE WHEN i.page_count IS NOT NULL THEN i.provenance ->> 'page_count' END AS pages_src
  FROM items i
  LEFT JOIN item_translations te ON te.item_id = i.id AND te.locale = 'en'
  WHERE i.visibility = 'public'
),
flags AS (
  SELECT
    p.type,
    p.title_src = ANY (a.api) AS title,
    p.synopsis_src = ANY (a.api) AND NOT p.synopsis_src = ANY (a.limited) AS synopsis,
    p.release_src = ANY (a.api) AND NOT p.release_src = ANY (a.limited) AS release_date,
    p.runtime_src = ANY (a.api) AND NOT p.runtime_src = ANY (a.limited) AS runtime,
    p.pages_src = ANY (a.api) AND NOT p.pages_src = ANY (a.limited) AS pages,
    -- The loader passes no cover source until covers have a public URL builder, so the API never shows one.
    false AS cover,
    EXISTS (SELECT 1 FROM jsonb_object_keys(p.external_ids) k
      WHERE k IN ('wikidata', 'isbn13') OR k = ANY (a.api)) AS external_ids,
    EXISTS (SELECT 1 FROM item_entities ie
      WHERE ie.item_id = p.id AND ie.role = 'genre' AND ie.source = ANY (a.api)) AS genres,
    EXISTS (SELECT 1 FROM item_entities ie
      WHERE ie.item_id = p.id AND ie.role <> 'genre' AND ie.source = ANY (a.api)) AS credits
  FROM public_items p
  CROSS JOIN allowed a
)
SELECT
  f.type::text AS type,
  v.field,
  count(*)::int AS items,
  round(100.0 * count(*) FILTER (WHERE v.ok) / count(*), 1) AS pct
FROM flags f
CROSS JOIN LATERAL (VALUES
  ('title', f.title), ('synopsis', f.synopsis), ('releaseDate', f.release_date),
  ('runtimeMinutes', f.runtime), ('pageCount', f.pages), ('cover', f.cover),
  ('externalIds', f.external_ids), ('genres', f.genres), ('credits', f.credits)
) AS v (field, ok)
GROUP BY f.type, v.field
ORDER BY f.type, v.field;
