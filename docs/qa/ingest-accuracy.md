# Ingestion accuracy sampling (CW-12)

Weekly procedure that measures how often the catalogue worker's stored fields match their sources. It implements RFC-001 §7 ("Sampling") and the CW-12 row of RFC-001 §11. The first report waits for CW-04 (`refresh-existing`) on the local QA stack: **every number below is pending a run** until then.

## Scope

- Measures fields written by the ingest pipeline (`internal/ingest/**`), not the hand-written catalogue.
- Runs on a local scratch database only, with the export of the week's sample. Production data is read only with the owner's yes.
- Ingestion stays dark in production: `FLAG_INGEST_*` flags stay off. A sample taken from the local scratch database does not turn anything on.
- No external API is called by the procedure as code. The checker opens the public source pages in a browser, as a person would. Nothing here calls TMDB, IGDB, Wikidata or the LLM from a test or a script.

## Target and cadence

- Sample at least 50 items per week (RFC-001 asks for 50 out of 200 exported per category).
- Target: at least 97 % field accuracy at each milestone gate (RFC-001 §7).
- One report per ISO week, created on the first working day of the next week.

## Step 1: export the week's pool

The worker exports 200 random items per category to `qa/reports/ingest-sample-<YYYY-MM-DD>.ndjson` (RFC-001 §7). Each line is one item with the fields listed in step 3 and its provenance (source URL and fetch time). The export is produced by the job named in CW-02/CW-04; this file defines only how to draw from it.

## Step 2: draw the sample

1. Record the ISO week (for example `2026-W45`) and the export file name in the report.
2. Draw without replacement with a fixed seed equal to the ISO week number, so anyone can redraw the same sample from the export.
3. Take at least 12 items from each category present in the export, so that 50 items are checked in total when four categories exist. A category with fewer than 12 items in the export is taken whole, and the shortfall is written in the report.
4. Do not replace an item once drawn. An item that cannot be checked (source page gone, offline) is kept in the sample and marked `unverifiable`, not replaced.

## Step 3: fields checked and the source for each

For each item, check the fields below against the source named. A field the item does not have is marked `n/a`, not `ok`.

| Field | Source to compare | How to check | Counts as ok when |
| --- | --- | --- | --- |
| Title | Wikidata entity (label in `en`) and the source page used by the connector | open both, compare the label | the stored title is the same work (not a sequel, remake or other edition) |
| Release year | Wikidata property for publication date, or the connector's source page | compare the year | the year matches; a difference of one year counts as a miss |
| Synopsis (`en`) | the Wikipedia article linked by the item (extract) | compare the first paragraph's meaning | the stored text is the same work's summary, not another work's |
| Cover present | the cover URL stored with the item | open the stored cover | it shows the cover of the same work |
| Credits (count) | the source page that the connector used | count the listed credits | the stored count is within 1 of the source |
| Trailer (verified) | the trailer link stored | open the link | the video title matches the item |
| External IDs | the source page for each stored ID (TMDB, IGDB, Open Library, Steam, GOG as present) | open the page for the ID | the page is for the same work |

Only public pages are read. The checker never logs in to a source and never copies a source's data into the report beyond the ids and a short verdict.

## Step 4: record each check

One line per field in the report table below. A miss is any field whose verdict is `miss`.

| Column | Meaning |
| --- | --- |
| `sample_id` | `<item id>` from the export |
| `field` | one of the seven fields in step 3 |
| `stored` | the value in the local scratch database (short) |
| `source` | the value on the source page (short) |
| `source_url` | the URL read, with the date read |
| `verdict` | `ok`, `miss`, `unverifiable` or `n/a` |
| `error_class` | for a miss: `wrong_work`, `wrong_value`, `stale`, `missing`, `mapping` (the source is right but the normaliser changed it) |
| `action` | the follow-up: `field_reviews` row id, or the ticket if the cause is code |

Misses are also written to `field_reviews` (the table RFC-001 §7 extends) when it exists on a local scratch database at run time. If it does not exist, the miss stays in the report and the report says so.

## Step 5: compute accuracy

- Field accuracy = `ok` fields / (`ok` + `miss`) fields. `unverifiable` and `n/a` are excluded from the denominator and counted separately.
- Report the accuracy overall, per category and per field.
- Count the items with at least one miss separately (item error rate).

## Report template

File: `docs/qa/ingest-accuracy-<YYYY-Www>.md` (new per week; not committed until the numbers exist).

```markdown
# Ingestion accuracy, <YYYY-Www>

Status: pending a run | measured
Export: qa/reports/ingest-sample-<date>.ndjson (<n> items per category)
Sample: <n> items (<per category>), seed <ISO week number>
Checker: <name>, date read <YYYY-MM-DD>

| Metric | Target | Result |
| --- | --- | --- |
| Field accuracy (all fields) | >= 97 % | pending a run |
| Items with at least one miss | recorded | pending a run |
| Unverifiable items | recorded | pending a run |

Per field: <table>
Per category: <table>
Misses: <table from step 4>
Follow-ups: <field_reviews ids or tickets>
```

## Open points

- Four categories (games, movies, series, books) are assumed; the export defines the real categories and the report uses those.
- The 97 % target is the RFC-001 milestone gate; no threshold is set for a single week.
