# ADR 0010: Community edits as moderated suggestions with audit trail

Status: accepted (2026-10-28). Decisions: owner requirement 5; brainstorming D-15.

## Context

Requirement 5: any user may suggest edits to any catalogue item, at most 3 per day, with moderation, states, audit trail, anti-abuse beyond the limit, and reversion. Competitors (`docs/web-research.md` §9.4): IMDb staff-reviews contributions in 1–7 days with a public status tracker and tracks contributor accuracy; TMDB lets users edit directly and moderators remove unsupported content, requiring an independent source; Goodreads restricts edits to volunteer librarians; Metacritic and Rotten Tomatoes allow none.

## Decision

Edits are **suggestions**, never direct writes. A suggestion is a field-level change set with before/after values, a required source URL for factual fields, and a note; it moves through a state machine (draft → submitted → in_review → approved → applied | rejected; applied → reverted; submitted → withdrawn; stale → expired). Admins moderate from the vukem admin panel at launch (D-15) with a target decision time under 24 hours and an ageing indicator. Applying a suggestion writes the catalogue transactionally, sets provenance to `community`, emits cache invalidation, and appends to an append-only audit log; reverting restores the previous value through the same path. Quota: 3 submissions per user per UTC day; further anti-abuse: minimum account age and verified email, contributor accuracy score adjusting limits, duplicate detection per field, cooldown after rejections, locked fields/items. Contributor accuracy is recorded from day one so a trusted-moderator role can be enabled later without migration. Authors see a public status page per suggestion; items show a "recent changes" tab.

## Alternatives considered

- **Direct edits with post-hoc moderation (TMDB).** Faster for contributors, but vandalism reaches users and the paid API; needs many moderators. Rejected for launch.
- **Auto-approve low-risk fields with a valid source.** Reasonable later; at launch there is no accuracy history to gate on. Deferred.
- **Librarian-only edits (Goodreads).** Excludes most users and recreates the backlog problem. Rejected.

## Consequences

- Positive: safe by construction, fully auditable, reversible, and visibly faster than IMDb if the SLA holds.
- Negative: moderation load falls on the owner at launch; the SLA depends on admin availability (mitigation: queue ageing alerts and bulk actions).
- Obligations: SDD 04 specifies tables, states, quotas, anti-abuse rules, admin UI, metrics and tests.
