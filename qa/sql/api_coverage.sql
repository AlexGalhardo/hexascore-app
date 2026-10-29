-- Per-type coverage of the fields the public API may return (SDD 05 section 6.5, open issue 05-11).
-- Percentages of public items that carry each field. Run on the first staging copy of the production catalogue
-- (P2.F) and paste the result, with the date, into resources/docs/coverage.md and the pricing statement:
--   psql "$DATABASE_URL" -f qa/sql/api_coverage.sql
-- Read-only; takes no locks beyond a snapshot.
WITH public_items AS (
  SELECT id, type, release_date, runtime_minutes, season_count, episode_count, page_count, external_ids
  FROM items
  WHERE visibility = 'public'
),
roles AS (
  SELECT item_id, role FROM item_entities GROUP BY item_id, role
)
SELECT
  i.type,
  count(*) AS items,
  round(100.0 * count(i.release_date) / count(*), 1) AS release_date_pct,
  round(100.0 * count(i.runtime_minutes) / count(*), 1) AS runtime_pct,
  round(100.0 * count(i.page_count) / count(*), 1) AS pages_pct,
  round(100.0 * count(i.episode_count) / count(*), 1) AS episodes_pct,
  round(100.0 * count(*) FILTER (WHERE EXISTS (
    SELECT 1 FROM item_translations t WHERE t.item_id = i.id AND t.synopsis IS NOT NULL)) / count(*), 1) AS synopsis_pct,
  round(100.0 * count(*) FILTER (WHERE EXISTS (
    SELECT 1 FROM roles r WHERE r.item_id = i.id AND r.role = 'genre')) / count(*), 1) AS genres_pct,
  round(100.0 * count(*) FILTER (WHERE EXISTS (
    SELECT 1 FROM roles r WHERE r.item_id = i.id AND r.role IN ('platform', 'director', 'author'))) / count(*), 1) AS credits_pct,
  round(100.0 * count(*) FILTER (WHERE i.external_ids ? 'isbn13' OR i.external_ids ? 'openlibrary') / count(*), 1) AS book_ids_pct
FROM public_items i
GROUP BY i.type
ORDER BY i.type;
