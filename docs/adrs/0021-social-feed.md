# ADR 0021: the activity feed is built from the write path, without fan-out

Status: accepted (2026-11-01). Decisions: D-06 (amended 2026-10-31, Q-A1), D-29; SDD 14 §12; ROADMAP P7.A.1.1.

## Context

Phase 7 adds follows and a feed of followed users' activity: ratings saved, status changes, lists created, items added to lists, reviews published (SDD 14 FR-14-16). The feed must answer p95 ≤ 300 ms at 30 rps with 10k events (NFR-14-7), respect the privacy matrix (SDD 14 §12.5) and a per-user opt-in while the lawful basis is open (Q-A5, SDD 14 open issue 8). HexaScore runs one Postgres and one app service on Railway; there is no Redis or stream store.

The outline in `.orchestration/reports/AUDIT-social-sdd.md` §3 lists three options:

1. Pull over the source tables: the feed is a `UNION` over `ratings`, `item_status`, `list_items` and `reviews` of followed users, with a keyset cursor on `(occurred_at, id)`.
2. An `activity_events` table written in the same transaction as each write, read with a join on `follows` and a keyset cursor on the UUID v7 id; 180-day retention.
3. Fan-out on write into per-follower inbox rows.

## Decision

Option 2. Each write path that produces a feed event already dispatches an event after commit (`scoring.RatingCreated`, `RatingUpdated`, `StatusChanged`, `library.ListCreated`, `ListItemAdded`); a listener in `app/social` inserts one `activity_events` row for it, only when the feature flag `social` is on and the actor has `feed_opt_in`. A crash between the commit and the listener loses one feed entry, never the write itself. The feed reads it with one query (SDD 14 §12.7). Events older than 180 days are deleted daily.

Fan-out (option 3) is not built until a user has 10,000 followers; at that point a follow-up ADR revisits it with measured numbers.

## Consequences

- One indexed table serves `/feed`, the profile activity tab and, later, "rated by people you follow" (P7.C.1.1); the cursor is the event id.
- The source tables keep their shape. Option 1 needs four differently shaped `UNION` branches, each with its own visibility filter, and its cost grows with every new event kind.
- Rating edits within 10 minutes update the last event instead of inserting, so repeated star changes do not flood followers.
- Events duplicate a little data (ids and stars) and need a purge job and an entry in the account export and deletion purge (SDD 14 §12.9).
- Opting out hides past events at read time; nothing is rewritten.
- With the flag off nothing is written, so turning the layer on later starts with an empty feed. That is acceptable while the lawful basis is open.
