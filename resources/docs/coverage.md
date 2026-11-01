# Data coverage

Every field the API returns comes from a source we are allowed to republish. Some fields are therefore patchy, and some are always omitted. This page states what to expect before you build on it.

## Fields and their sources

| Field | Source | Expected gap |
| --- | --- | --- |
| `id`, `type`, `slug`, `url`, score fields, `trending`, `timeToBeat` | HexaScore (API terms) | none |
| `title`, `releaseDate`, `genres`, `people`, `organisations`, `franchises`, `platforms`, `countries`, runtime, pages, seasons, episodes | Wikidata (CC0) and the community | runtime, platforms and episode counts are patchy |
| `synopsis` | Wikipedia (CC BY-SA 4.0, with `sourceUrl`) | items without an article |
| editions and book covers | Open Library and Wikidata | not applicable to other types |
| covers for films, series and games, and trailers | none | never in the API |
| `externalIds` | Wikidata, ISBN-13, Open Library | TMDB, IGDB and Steam ids are omitted |

Every response lists the fields it left out in `meta.omittedFields`, and each sourced value carries its `provenance` (source and licence).

## Coverage percentages

The table below gives, per type, the share of public items for which the API returns each field. It is measured with `qa/sql/api_coverage.sql` on a copy of the production catalogue, using the same source rules as the API, and shows the date of the measurement.
