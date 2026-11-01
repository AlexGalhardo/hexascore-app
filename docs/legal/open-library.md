# Open Library licence check (A-05)

Status: decided under the A-05 rule (SDD 06 §7.1) after the owner cleared the legal reads (Q-A16, 2026-11-01). The owner confirms this file before Open Library content reaches the paid API.

Both pages were accessed on 2026-11-01, once each. Quotes are verbatim and short; the rest is summarised.

## Data: Developer Center / Licensing

Source: <https://openlibrary.org/developers/licensing> (page last edited 2021-06-27).

- The Internet Archive "does not assert any new copyright or other proprietary rights" over the database (Internet Archive, Open Library licensing page).
- It also warns: "There may be existing rights issues on some contributions" (same page).
- Summary: the page names no licence (it does not say CC0 or ODbL) and sets no condition: no attribution requirement, no non-commercial clause, no redistribution limit. It describes the goal as a database that can be openly used for many purposes.

## Covers: Covers API

Source: <https://openlibrary.org/dev/docs/api/covers>.

- "Please, do not crawl our cover API" (Open Library, Covers API, "Guidelines for Cover Use").
- The API is "intended for displaying covers on public facing websites and not for bulk download" (same page, "Bulk Access").
- "A courtesy link back to Open Library is appreciated" (same page, guideline 3).
- Summary: for public pages the guidelines ask for an `<img src>` that points to `covers.openlibrary.org`; bulk copies come from the archive.org cover items instead. Lookups by ISBN and other non-CoverID or non-OLID keys are limited to 100 requests per IP every 5 minutes (403 above that). The page grants no licence for the images and says nothing about redistribution.

## A-05 decision

| Scope | Rule branch (SDD 06 §7.1) | Decision |
| --- | --- | --- |
| Bibliographic data (titles, identifiers, descriptions, dates, page counts, editions) | no rights asserted, treated as the public-domain branch | ingest; API-eligible; licence `OPEN-LIBRARY-NO-RIGHTS-ASSERTED` with the URL of the licensing page; attribution not required |
| Covers | no explicit grant | not API-eligible; site display only, by the guidelines above |

Consequences:

- `CATALOG_API_SOURCES` keeps `openlibrary` (it is already in `config.DefaultCatalogAPISources`). `.env.example` now lists the variable empty, which means that default.
- The `data_sources` row is already right: `internal/ingest/normalize/provenance.go` seeds `OPEN-LIBRARY-NO-RIGHTS-ASSERTED` and cites the licensing page. No change.
- Follow-up after the owner confirms this file: `app/catalog/audience.go` still lists `openlibrary` in `limitedSources`, so the API serves only its identifiers and titles. Lifting that for descriptive fields needs a separate cover rule, because `AllowsContent` also gates covers, and Open Library covers must stay out of the API.
- Follow-up for the covers pipeline: the cover guidelines ask public pages to point at `covers.openlibrary.org` and not to bulk-copy through the API. `internal/ingest/covers` copies Open Library covers into the bucket. The owner decides whether to keep the copies, which is low volume and paced under the rate limit, or switch to hot-linking with a courtesy link.
- A takedown for a contribution with existing rights follows the takedown process of SDD 01; the page's own caveat above is why that path matters here.
