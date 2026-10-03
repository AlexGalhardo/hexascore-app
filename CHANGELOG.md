# Changelog

All notable changes to this project are documented here.
Format: [Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/). Versioning: [SemVer](https://semver.org/).

## [Unreleased]

### Added

- `PLAN.md` with decisions and milestones; reference docs in `docs/` (architecture, data model, API, frontend, design system, payments, ingestion, security, testing, process).
- `AGENTS.md` and `CLAUDE.md` agent guides.
- Catalog ingestion in `webscrappings/` (Python 3.14 + uv): Wikidata, Wikipedia, Open Library, Steam and GOG connectors into `source_records`, normalizer to `items`/`item_translations`/`external_scores`, bootstrap with gzipped SQL snapshot export, covers (Vercel Blob optional), revalidation call, pytest suite with recorded fixtures.
- GitHub Actions: `ingest-bootstrap.yml` (manual), `ingest-incremental.yml` (nightly), `ingest-weekly.yml` (scores and synopses), `python-ci.yml` (ruff + pytest on PRs touching `webscrappings/**`).

### Removed

- Files inherited from the previous project (RespondeAê): old plan, docs, setup scripts, tests, REST client samples, graphify output, `vercel.json`.
