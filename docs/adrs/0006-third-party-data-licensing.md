# ADR 0006: Third-party data posture: Wikidata-only in the paid API, critic scores dropped

Status: accepted (2026-10-28). Decisions: brainstorming D-03, D-09, A-05; owner requirement 7 (respect licences and terms of each source).

## Context

Findings in `docs/web-research.md` (sources accessed 2026-10-28):

- TMDB API Terms of Use (2023-10-20): no commercial use without a separate written agreement; no caching beyond 6 months; no resale; logo attribution required.
- IGDB official FAQ: free for commercial projects with a partnership arranged by email and visible static attribution; local storage encouraged; the binding Twitch Developer Services Agreement could not be read.
- IMDb datasets: personal and non-commercial use only; API via AWS Data Exchange at a listed US$150,000/year.
- Metacritic and Rotten Tomatoes: scores are licensed products (Fabric Origin, proposal forms), display must link back, no sharing.
- Wikidata: CC0. Wikipedia extracts: CC BY-SA with attribution and source URL (already stored by the legacy app). Open Library: licence not yet checked (A-05).

The legacy app serves TMDB- and IGDB-derived fields through a paid API without agreements and shows Metacritic/Rotten Tomatoes values in `external_scores`.

## Decision

1. The **paid public API exposes only fields whose provenance is Wikidata, Wikipedia (with attribution metadata), Open Library (after A-05) or HexaScore itself** (scores, statuses, trending, community edits). TMDB- and IGDB-derived fields are display-only on hexascore.app and the mobile app, never in `/api/v1`.
2. **Every stored field carries provenance**: `source` (enum) and `licence` (enum) at field level for enriched data; the API serialiser filters by an allowlist of sources per audience. Changing the posture later is configuration plus a re-serialisation, not a migration.
3. **TMDB values are refreshed or dropped within 6 months** of fetch; a scheduled job enforces it. **IGDB attribution** is shown in a static place on item pages and the about page; **TMDB logo attribution** and the not-endorsed notice are shown where TMDB fields are used.
4. **Critic scores are removed** at launch (`external_scores` table dropped from the model). They return only under a licence, via a new ADR.
5. **IMDb datasets are not ingested.**
6. No commercial agreements are pursued now (owner choice). The pricing and API docs state the data coverage of the API explicitly.

## Alternatives considered

- **Pursue written agreements with TMDB and IGDB now, with this posture as automatic fallback (architect's recommendation).** Keeps richer API coverage if agreements arrive. The owner chose the cheaper, immediately compliant option; this ADR can be superseded when an agreement exists.
- **Drop TMDB and IGDB entirely.** Loses synopsis, cover and game metadata coverage on the site. Rejected.
- **Keep the legacy behaviour.** Open legal exposure for a paid product. Rejected.

## Consequences

- Positive: the paid API is legally clean at launch; provenance is a DX feature (developers see the licence of each field).
- Negative: API coverage for films, series and games is thinner than the site's; buyers must be told; some fields (e.g. runtime, platforms) may be missing where Wikidata is incomplete.
- Obligations: SDD 01 defines the provenance columns and the serialiser allowlist; SDD 06 defines the 6-month TMDB refresh job and the attribution checks; SDD 05 documents coverage per field; the owner reads the Twitch Developer Services Agreement before relying on IGDB display beyond the FAQ.
