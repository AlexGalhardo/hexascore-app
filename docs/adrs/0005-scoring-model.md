# ADR 0005: Half-star input, published Bayesian aggregate, stars 0–10 display

Status: accepted (2026-10-28). Decisions: brainstorming D-00, D-13; product goal 3 (scores more reliable than competitors).

## Context

Legacy: six categories per media type, each 0–100 integer entered with a slider, `overall = round(sum/6)` generated column, item aggregates recomputed transactionally, 5-rating minimum applied on read. Competitor research (`docs/web-research.md` §8–9): every incumbent hides its formula; review bombing is handled reactively and opaquely; Rotten Tomatoes is the only one publishing rules and its headline is binary.

## Decision

1. **Input**: a star widget per category, half-star steps from 0.5 to 10.0, one click (D-00). Stored as integers 5–100 in steps of 5 (`stars × 10`). No row = no rating.
2. **Aggregate**: per category and overall, a Bayesian average toward a per-type prior with a published `m` (minimum weight), a published minimum of 5 ratings to display, and a confidence level derived from count and variance. The formula, its constants and its version are public on a help page and exposed as `formulaVersion` in the API. SDD 02 fixes the exact formula and constants.
3. **Display**: stars 0–10 with one decimal everywhere (site, app, API), plus `ratingCount` and `confidence` (D-13). The internal 0–100 precision is not exposed.
4. **Anti-manipulation as published policy** (SDD 02): no ratings before release; burst limiter when activity on an item is anomalous; account-age and verified-activity weighting; per-user daily limits; shadow-weighting instead of deletion; a visible "unusual activity" notice; a nightly full recompute that self-heals aggregates.
5. **Legacy migration**: existing values are rounded to the nearest multiple of 5 with an audit record.

## Alternatives considered

- **Keep 0–100 sliders.** More granular, slower to use, no competitor familiarity. Rejected by the owner.
- **Plain arithmetic mean (Goodreads, Skoob).** Easy to game with few votes; no confidence signal. Rejected.
- **Undisclosed weighting (IMDb, Metacritic).** Protects against gaming by obscurity but contradicts the product goal of trustworthy, explainable scores. Rejected.
- **0–100 display with confidence band (architect's recommendation).** Keeps the "HexaScore 0–100" identity; the owner preferred stars for familiarity. Confidence survives as a field and a UI indicator.

## Consequences

- Positive: explainable scores; a differentiator no competitor occupies; a storage format compatible with the legacy columns and API DTOs (steps of 5 inside 0–100).
- Negative: a published formula can be targeted; the mitigation is the published anti-abuse rules plus shadow weights whose parameters are not published.
- Obligations: formula changes are versioned and announced like API changes; unit tests pin the formula with worked examples; property tests check monotonicity and bounds.
