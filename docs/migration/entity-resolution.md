# Legacy entity resolution

The `entities` stage of `hexascore:import-legacy` (P2.F.2.5, SDD 01 §5.4) turns the legacy `items.meta` strings (`genres`, `platforms`, `creators`, `countries`) into `entities` and `item_entities`, and each legacy book `external_ids.isbn13` into one edition. It is the minimal mapping; the Go normaliser (P6.A) rebuilds entities from `source_records` later.

## Rules

- Resolved by QID: a string equal to the label (`payload.refs`) of a claim of the role's property in the item's `wikidata` source record (`P136` genre, `P400` platform, `P495` country, creators `P57` movie, `P170` series, `P178` game, `P50` book), exactly as legacy `normalize.py` `ref_labels` built it. The entity with that QID is reused whatever its kind; the link has `source = wikidata`.
- Native: any other string. One entity per kind and exact name (`wikidata_qid` null, listed by DQ-09); the link has `source = legacy` (data source with licence `unknown`, so the API never serves it).
- Roleless creators: every legacy creator; role by item type: `creator` (movie, series, person), `author` (book, person), `developer` (game, organization).
- Slugs: `catalog.UniqueSlug` on the name; a name without Latin letters or digits uses the lowercase QID, else `<kind>-<8 hex of sha256(name)>`.
- Editions: digits of `isbn13`, kept only when `isbn13_is_valid`; `language = und` (legacy stores no edition language, DQ-10), publisher null, `source = legacy`, title = the work's `en` title. A missing, invalid or already-used ISBN yields no edition and never fails the stage (01-1). `items.external_ids.isbn13` is left in place.

## Counts

| Run | Links resolved by QID | Native links | Native entities | Roleless creator links | Editions |
| --- | --- | --- | --- | --- | --- |
| Fixture (`TestEntitiesStage`) | 2 | 5 | 5 | 3 | 1 |
| Production dump | owner-gated, fill after the rehearsal | | | | |

Queries for the production rehearsal, run after the stage:

```sql
SELECT count(*) FILTER (WHERE source = 'wikidata') AS by_qid, count(*) FILTER (WHERE source = 'legacy') AS native_links,
  count(*) FILTER (WHERE role IN ('creator', 'author', 'developer')) AS roleless_creators
FROM item_entities;
SELECT count(*) AS native_entities FROM entities WHERE wikidata_qid IS NULL;
SELECT count(*) AS editions FROM editions WHERE source = 'legacy';
```
