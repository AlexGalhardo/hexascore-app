# AUDIT-social-sdd: SDD 14 outline for Phase 7 (follows, feed, reviews, comments, reactions)

status: concluida (outline only; no decision taken, no SDD edited)
author: especialista, 2026-10-31
inputs: ROADMAP P7.A–P7.C (lines 3780–3871), SDD 14 (whole), SDD 08 §7.1, SDD 13 §7.0.1/§7.1/§11, SDD 11 §7.6, `docs/brainstorming.md` D-06/D-29/A-07, legacy `apps/web/lib/db/schema/social.ts`, `docs/web-research.md` headings
purpose: P7.A.1.1 (Opus) starts from this draft; every `OPTION` below is for the P7.A.1.1 author or the owner, not decided here

## 1. Defects in the current Phase 7 plan (fix in P7.A.1.1 or earlier)

1. Section clash: ROADMAP P7.A.1.1, P7.A.1.2, P7.A.1.3, P7.B.1.1, P7.B.1.2, P7.C.1.1 cite "SDD 14 §8 Follows and feed", but SDD 14 §8 is "Edge cases" and is cited elsewhere (§8 purge rules). Propose: new sections §12 "Social layer" (or a new SDD 15 "Social"); keep §8. The ROADMAP refs must be rewritten either way.
2. Wrong research refs: P7.A.1.1 cites "web-research §4 (Letterboxd, Goodreads, Skoob feeds)" and P7.B.1.1 cites "web-research §4.2 (Letterboxd review model)". §4 is Metacritic; Goodreads is §6, Skoob §7; Letterboxd is not in `docs/web-research.md` at all. Needs an estagiario research row (Letterboxd feed/review/like model, with sources) before P7.A.1.1, or drop the ref.
3. Wrong OWASP categories (SDD 13 §7.0.1 uses Top 10:2025): P7.B.1.1 names "SDD 13 A03 (sanitised Markdown)" and `TestSecurity_A03_ReviewMarkdown`; A03 is supply chain, injection/XSS is A05 → `TestSecurity_A05_ReviewMarkdown`. P7.A.1.2 and P7.B.1.2 name "SDD 13 A04 (rate limit)"; A04 is crypto, limits are A06 (`app/http/limits`, `TestSecurity_A06_HexaLimits`).
4. Follow model contradiction: P7.A.1.2 requires `TestSecurity_A01_PrivateProfileFollowRequiresApproval` (follow requests), but SDD 14 §5 "Deferred shape" and legacy `follows` have no state column (follower, followee, created_at). Approval needs a `status`/`pending` state, a requests inbox and a notification kind.
5. Feed source contradiction: SDD 14 §5 says the feed reads `ratings.updated_at`, `item_status.updated_at`, `list_items.added_at` with "no new columns", while P7.A.1.1 plans an `activity_events` table and ADR 0019 "built from the write path". Pick one in ADR 0019 (options in §3.2) and delete the other sentence.
6. LGPD basis: the Phase 7 principle "consent for the feed (opt-in at onboarding, off by default for existing users)" is not in D-29 (D-29 covers personalisation and aggregates only), and SDD 13 open issue 4 already records a consent vs legitimate-interest tension. A new D-xx is needed (owner, D-21 legal review). Not decided here.
7. Dependency order: P7.A.1.3 reuses "activity primitives shared with trending, P6.C.1" and P7.C needs P6.C, but P7.A depends only on P2.G. Either add `P6.C.1.1` to P7.A.1.3 "Depends on" or state that P7 owns its own event table. Also the ref "SDD 08 §6" should be §7.1 (Primitives).
8. Markdown subset for reviews: `go.mod` has no Markdown renderer and vukem docs name none (`[NOT IN VUKEM DOCS]` to confirm on <https://vukem.dev/docs>). Options: stdlib-only subset parser (bold, italic, links, line breaks, spoiler block) escaping through `html/template`; or a new dependency (owner yes + ADR). The JS `scripts/build-docs.ts` renderer is build-time only and not for user input.
9. Imported data: SDD 11 §7.6 excludes legacy `follows` from the import but keeps it in the final dump. Phase 7 must say whether those rows are restored from the cutover archive (and under which consent rule, defect 6) or dropped.

## 2. Proposed section map (SDD 14 §12 "Social layer", or SDD 15)

- 12.1 Objective and scope: follows, feed, reviews, comments, reactions; out: DMs, groups, badges, push, social data in `/api/v1` (A-07), mobile social (SDD 10 excludes namespace `social`: OPTION keep out of v1.13 or add `/app/v1` routes).
- 12.2 Functional requirements FR-14-13..: follow/unfollow idempotent; private profile → follow request; block (OPTION: in scope or later); feed 20/page cursor; review one per (user, item), ≤ 5,000 chars, spoiler flag; comments one level, on reviews and lists; reactions `like`, `helpful`, unique (user, target); author can edit/delete; moderation hides everywhere.
- 12.3 NFR: feed p95 ≤ 300 ms at 30 rps with 10k events (P7.A.1.3); rate limits follows 60/h, comments 60/h, reviews (OPTION), reactions (OPTION) as named rules in `app/http/limits`; retention activity 180 days; i18n en/pt-BR/es; axe 0 serious.
- 12.4 Data model (DDL to write): `follows` (+ `state` if defect 4 goes the approval way, `PK(follower_id, followee_id)`, `CHECK follower_id <> followee_id`, index on followee); `activity_events` or none (defect 5); `reviews` (FK to rating or to (user, item); `hidden`; `helpful_count`); `comments` (`target_type in ('review','list')`, `parent_id` nullable, depth 1 by CHECK/trigger); `reactions` (`UNIQUE(user_id, target_type, target_id, kind)`); counters on the parent rows updated in the same transaction (P7.B.1.2). Changes to existing tables: `reports.target_type` + `review`, `comment`; `reports.outcome` + `review_hidden`, `comment_hidden`; `user_notifications.kind` + `new_follower`, `follow_request`, `comment`, `reaction` (dedupe keys and aggregation, e.g. one reaction notice per target per day).
- 12.5 Privacy matrix (rows: viewer anonymous / member / follower / owner / admin; columns: public profile, private profile, hidden list, banned author, deleted author, hidden review/comment) for: profile counts, follower lists, feed entries, reviews, comments, reaction counts, "friends rated" (P7.C.1.1). Rule to state: counts never include private or banned users (FR-14-8 already hides counts on private profiles).
- 12.6 Contracts: web routes (`PUT|DELETE /u/{nickname}/follow`, `/feed`, review and comment CRUD, reactions), policies (`FollowPolicy`, `ReviewPolicy`, `CommentPolicy` with `authz.DenyAsNotFound`, as §6.1), Inertia props, `Defer` group `personal` for cacheable public HTML (P7.C.1.1), problem codes, events (`q.DispatchTx`).
- 12.7 Behaviour: feed query (pull from followed ids with cursor, bounded by follow count; fan-out threshold per ADR 0019), Markdown subset and sanitising (defect 8), helpfulness sort, edit window, notification aggregation.
- 12.8 Edge cases: unfollow mid-scroll, profile turns private (existing followers keep or lose access: OPTION), ban/unban, item merge (reviews and reactions move like `list_items`, target wins), item hidden/takedown, user purge (reviews: anonymise or delete, OPTION aligned with open issue 14-7), self-reaction, review on a deleted rating.
- 12.9 LGPD: basis (defect 6), export (§7.4 ZIP gains `follows`, `reviews`, `comments`, `reactions`; NFR-14-3 budget re-check), purge rules (SDD 03 §7.5), retention.
- 12.10 Acceptance and tests: `TestSecurity_A01_PrivateProfileFollowRequiresApproval`, `TestSecurity_A01_FeedPrivacy`, `TestSecurity_A05_ReviewMarkdown`, `TestSecurity_A06_HexaLimits` extended, `TestReportsReviewComment`, k6 `qa/k6/feed.js`, Playwright `social.spec.ts` (restore the dropped follow/feed tests), `reviews.spec.ts`.
- 12.11 Open issues: defects 4–9 that stay open, each with owner and closing condition; add rows to `docs/sdd/open-issues.md` (P7.A.1.1 acceptance requires it).

## 3. Options for ADR 0019 (no choice made)

1. Pull model: feed = UNION over `ratings`, `item_status`, `list_items`, `reviews` of followed users, keyset cursor on `(occurred_at, id)`; no new table; cost grows with followed count; no retention table to purge.
2. `activity_events` written in the same transaction as each write (`q.DispatchTx` or direct insert), pull on read, 180-day purge; one index `(actor_id, occurred_at DESC)`; matches P7.A.1.1 and "no fan-out until 10k followers".
3. Fan-out on write to per-follower inboxes: rejected in the ROADMAP until 10k followers; listed for completeness.

## 4. Follow-up rows proposed (orchestrator decides)

- `DOC-research-letterboxd` (estagiario, prio 3): Letterboxd feed, review, like and follow model with sources, as `docs/web-research.md` §11; unblocks defect 2.
- `DOC-roadmap-p7-refs` (orchestrator, prio 3): fix defects 1, 2 (refs), 3 and 7 in ROADMAP P7 text; docs only.
- Owner question (QUESTIONS.md, via orchestrator): defect 6 LGPD basis for the feed and defect 9 legacy follows restore.

changelog: none (report only)
roadmap: no box
