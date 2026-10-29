# ADR 0007: Ingestion rewritten in Go on vukem queue and scheduler, single process

Status: accepted (2026-10-28). Decisions: brainstorming D-08, D-11; owner requirement 7 (continuous ingestion, quality asserts, safe migrations).

## Context

Legacy ingestion is Python 3.14 (`legacy/webscrappings/`): connectors for Wikidata (SPARQL + entities), Wikipedia extracts (en/pt/es), Open Library, Steam, GOG, TMDB and IGDB; `normalize.py` turns `source_records` into items, translations and external scores; `covers.py` produces WebP covers and thumbnails; `trailers.py` validates YouTube videos; GitHub Actions workflows exist but are disabled, so runs are manual from a developer machine against production. 77 pytest tests with recorded fixtures document the behaviour.

vukem provides typed jobs with sync, memory and database drivers, a cron scheduler, an HTTP client with SSRF protection, storage, uploads and image handling via `contrib/imaging` (`https://vukem.dev/docs/digging-deeper/scheduler`, `https://vukem.dev/docs/architecture/lifecycle`). The launch topology is one Go process (D-08).

## Decision

Port ingestion to Go inside the application: one connector package per source behind a common interface (fetch → `source_records` with canonical hash → normalise → upsert), jobs on vukem's queue with the database driver, schedules on vukem's scheduler (daily incremental, weekly full for a source, covers and trailers as follow-up jobs), all jobs idempotent, resumable by cursor, throttled per host, and bounded in CPU so HTTP traffic keeps priority (worker concurrency 1–2, nice scheduling windows at night in the target region). Status, last run, counts and failures are visible in the admin panel. The Python package stays in `legacy/` as the executable specification: parity is proven by running both on a fixed sample and diffing the normalised output.

Data quality asserts (owner requirement 7) run as a job after each ingestion batch and fail loudly: counts per type do not drop more than 1 % without an explicit flag; every item has an English title; ISBN-13 checksums valid; cover files exist for items marked as having covers; translations carry licence and source URL; no TMDB-derived field older than 6 months (ADR 0006).

## Alternatives considered

- **Keep Python as a Railway cron service.** Fastest path, but a second runtime, a second container (cost under the US$30 ceiling), and two codebases for the same normalisation rules. Rejected by the owner.
- **Keep running Python by hand.** Zero infra cost, but data goes stale and production access from laptops is a security smell. Rejected.
- **Separate Go worker service.** Cleaner isolation; costs a second service. Deferred: the first thing to do when the budget ceiling rises (D-08).

## Consequences

- Positive: one runtime, jobs observable in the admin panel and metrics, ingestion runs continuously without a laptop.
- Negative: a rewrite of about 2,000 lines of Python plus fixtures; image processing in Go (`contrib/imaging`, which the vukem review flagged as needing work) must match the legacy WebP outputs or the covers pipeline is redesigned.
- Obligations: SDD 06 specifies connectors, schedules, backfills, rate limits per source, licence checks and the asserts; SDD 11 covers migrations and backups; `vukem-feedback.md` receives any queue/scheduler/imaging gap.
