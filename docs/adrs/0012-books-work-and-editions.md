# ADR 0012: Books modelled as work + editions

Status: accepted (2026-10-28). Decisions: brainstorming D-10, D-01; research `docs/web-research.md` §6, §7, §9.7.

## Context

Legacy stores one `items` row per book keyed by Wikidata QID with ISBN-13 and Open Library IDs in `external_ids` jsonb; covers come from Open Library; cover coverage for books is 34 % in production. Goodreads relies on volunteer librarians to combine editions by hand; Skoob shows 0–1 editions per work and has duplicate problems. Brazilian readers care about the Portuguese edition (publisher, translator, cover) while ratings and statuses are about the work.

## Decision

A book item is the **work** (identity: Wikidata QID when present, else a HexaScore ID). **Editions** are rows linked to the work: ISBN-13 (validated checksum, unique), ISBN-10 when known, language, publisher (organisation entity), publication year, page count, format (hardcover, paperback, ebook, audiobook), cover storage key, source and licence. Ratings, statuses, time-to-read, lists and the public score attach to the work; the item page lists editions with the user's locale edition first (pt-BR editions highlighted for Brazilian users) and lets a user pick the edition they own or read (stored on the status row, optional). Search indexes work titles and edition titles. Ingestion creates editions from Open Library and Wikidata edition statements; community suggestions can add or correct editions (SDD 04).

Games, films and series stay single-level at launch: platforms (games) and release dates per region (films) are relations on the item, not editions. A later ADR may generalise if a real need appears (D-10 option 3 rejected for now).

## Alternatives considered

- **Work only, ISBNs in a jsonb list (legacy).** Simple; cannot show publisher, translator or cover per edition; ambiguous covers. Rejected.
- **Work + editions for all media types.** Uniform but heavier; no user-facing need for films/series/games at launch. Deferred.
- **Edition as the item (no work level).** Fragments ratings across editions; the Goodreads/Skoob failure mode. Rejected.

## Consequences

- Positive: correct identity for ratings; Brazilian editions first-class; dedup by ISBN-13 at ingestion.
- Negative: more ingestion logic (matching editions to works), more UI (edition picker), a migration of legacy `external_ids.isbn13` into edition rows.
- Obligations: SDD 01 defines the tables and invariants (at least one edition per book when any ISBN is known; one primary edition per locale); SDD 06 defines edition sourcing and dedup; SDD 02 keeps scores on the work.
