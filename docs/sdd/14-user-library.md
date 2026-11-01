# SDD 14: User library and social basics

Status: reviewed (2026-10-28 consistency pass). Decisions: D-06, D-12, D-13, A-01, A-07. ADRs: 0005, 0009, 0010, 0016. Related SDDs: 01, 02, 03, 04, 05, 08, 09, 10, 13.

## 1. Objective

Give members a fast personal library (statuses, time-to-beat, ordered lists), a privacy-respecting public profile, and the trust tools around them (reports, contact, notifications, LGPD export). Lists also feed SDD 08 discovery; the model leaves room for follows and feed (D-06) without tables for them.

## 2. Scope

In (v1): `lists`, `list_items`, `user_item_editions`, `reports`, `contact_messages`, `user_notifications` (owned here, written by SDD 02/04 through one API), `data_exports`; pages Library, Lists, Profile, Notifications, Contact; `/app/v1` routes for D-12 (lists, profile, notifications, export); admin queues.

Out: follows, feed, follower counts (D-06); badges; avatar upload; push; reviews; status and time-to-beat writes (SDD 02).

## 3. Functional requirements

- FR-14-1: A list has one item type fixed at creation; another type → 422 `list-type-mismatch`. Integration.
- FR-14-2: Limits: 10 lists per (user, type), 500 items, title 1–80, description ≤ 500 (legacy `LIST_LIMITS`, `legacy/apps/web/lib/social/constants.ts`), note ≤ 280 (new). Breach → 409 `list-limit-reached`/`list-item-limit-reached`, 422 for lengths. Integration.
- FR-14-3: Re-adding an item is a no-op 200 (legacy `addListItem` `ON CONFLICT DO NOTHING`). Integration.
- FR-14-4: Only the owner reads a private/hidden list or writes any list; others get 404. `TestSecurity_A01_ListOwnerOnly`.
- FR-14-5: Reorder replaces the full order atomically; stale `version` → 409. Property test.
- FR-14-6: A list page is indexable only if public, not hidden, owner not private, ≥ 3 displayable items. Playwright meta check.
- FR-14-7: `/library` groups items by status, then type (`legacy/docs/frontend.md` "Padrão de listagens"), with counts. Playwright + axe.
- FR-14-8: A private profile shows only nickname and initial, 200 + `noindex`, no counts. Port of `legacy/qa/e2e/specs/social.spec.ts` test 1.
- FR-14-9: One open report per (reporter, target); repeats → 200, no row; self → 403 `report-self`; resolution notifies the reporter. Integration.
- FR-14-10: Contact messages are stored, mailed to `CONTACT_INBOX`, listed in admin with `handled`. Port of `contact.spec.ts`.
- FR-14-11: Header shows unread count ("99+" cap); mark-read by id or all. Playwright.
- FR-14-12: Export: once per 7 days, background ZIP, mailed link valid 24 h. Integration, `MAIL_MAILER=array`.

## 4. Non-functional requirements

- NFR-14-1: Public list/profile SSR p95 ≤ 400 ms at 30 rps, errors < 0.5 %; `/app/v1` list reads p95 ≤ 150 ms (k6, staging).
- NFR-14-2: Library for a 5,000-item user: ≤ 6 queries, p95 ≤ 300 ms (query-counting integration test).
- NFR-14-3: Export of a 20,000-rating user ≤ 60 s, ZIP ≤ 50 MiB, worker memory ≤ 128 MiB.
- NFR-14-4: Rate limits: list writes 60/min per user (legacy `actions/lists.ts`), reports 10/h (legacy `actions/reports.ts`), contact 5/h per IP (legacy `lib/contact.ts` `CONTACT_LIMIT`), mark-read 60/min; keyed with `ratelimit.ByUser(guard)` / `ratelimit.ByIP` (<https://vukem.dev/docs/security/rate-limiting> "Keys"). 429 `rate-limited`.
- NFR-14-5: en/pt-BR/es copy; axe 0 serious violations.
- NFR-14-6: Retention: notifications read > 180 days or any > 365 days deleted daily; contact messages 24 months; export files 24 h; resolved reports 24 months.

## 5. Data model

`users.id` is `text` (SDD 03 §5); SDD 04's `user_notifications.user_id bigint` is wrong and is replaced by this DDL (open issue 1).

```sql
CREATE TABLE lists (
  id uuid PRIMARY KEY,                                   -- v7
  user_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  type item_type NOT NULL,                               -- immutable (trigger)
  title text NOT NULL CHECK (char_length(title) BETWEEN 1 AND 80),
  description text CHECK (char_length(description) <= 500),
  is_public boolean NOT NULL DEFAULT true,
  hidden boolean NOT NULL DEFAULT false,                 -- moderation
  item_count int NOT NULL DEFAULT 0 CHECK (item_count BETWEEN 0 AND 500),
  version int NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now());
CREATE INDEX lists_user_type_idx ON lists (user_id, type);
CREATE INDEX lists_public_idx ON lists (updated_at DESC) WHERE is_public AND NOT hidden;

CREATE TABLE list_items (
  list_id uuid NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
  item_id uuid NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  position int NOT NULL CHECK (position > 0),
  note text CHECK (char_length(note) <= 280),
  added_at timestamptz NOT NULL DEFAULT now(),           -- SDD 08 signal
  PRIMARY KEY (list_id, item_id),
  UNIQUE (list_id, position) DEFERRABLE INITIALLY DEFERRED);
CREATE INDEX list_items_item_idx ON list_items (item_id);

CREATE TABLE user_item_editions (                       -- optional book edition
  user_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  item_id uuid NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  edition_id uuid NOT NULL REFERENCES editions(id) ON DELETE CASCADE,
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, item_id));                       -- trigger: edition.work_id = item_id

CREATE TYPE report_status AS ENUM ('open','resolved','dismissed');
CREATE TABLE reports (
  id uuid PRIMARY KEY,
  reporter_id text REFERENCES users(id) ON DELETE SET NULL,
  target_type text NOT NULL CHECK (target_type IN ('list','user','suggestion')),  -- + 'review' later
  target_id text NOT NULL,
  reason text NOT NULL CHECK (reason IN ('spam','offensive','harassment','impersonation','other')),
  details text CHECK (char_length(details) <= 500),
  status report_status NOT NULL DEFAULT 'open',
  outcome text CHECK (outcome IN ('dismissed','list_hidden','user_banned','suggestion_rejected','no_action')),
  admin_note text CHECK (char_length(admin_note) <= 1000),
  resolved_by text REFERENCES users(id) ON DELETE SET NULL,
  resolved_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now());
CREATE INDEX reports_status_idx ON reports (status, created_at);
CREATE UNIQUE INDEX reports_open_unique_idx ON reports (reporter_id, target_type, target_id) WHERE status = 'open';

CREATE TABLE contact_messages (                         -- legacy parity
  id uuid PRIMARY KEY, user_id text REFERENCES users(id) ON DELETE SET NULL,
  name text NOT NULL, email text NOT NULL,
  subject text NOT NULL CHECK (subject IN ('question','feedback','bug')),
  message text NOT NULL CHECK (char_length(message) BETWEEN 32 AND 512),
  handled boolean NOT NULL DEFAULT false, created_at timestamptz NOT NULL DEFAULT now());

CREATE TABLE user_notifications (
  id uuid PRIMARY KEY, user_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  kind text NOT NULL CHECK (kind IN ('suggestion_decided','suggestion_reverted','suggestion_expired',
       'report_resolved','rating_flagged','list_hidden','system')),
  payload jsonb NOT NULL,                                -- ids + i18n params, never HTML
  dedupe_key text NOT NULL,                              -- 'report:<id>'; retries insert once
  read_at timestamptz, created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (user_id, dedupe_key));
CREATE INDEX user_notifications_unread ON user_notifications (user_id, created_at DESC) WHERE read_at IS NULL;
CREATE INDEX user_notifications_user ON user_notifications (user_id, id DESC);

CREATE TABLE data_exports (
  id uuid PRIMARY KEY, user_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  status text NOT NULL CHECK (status IN ('queued','running','ready','failed','expired')),
  path text, size_bytes bigint, requested_at timestamptz NOT NULL DEFAULT now(),
  ready_at timestamptz, expires_at timestamptz);
CREATE INDEX data_exports_user_idx ON data_exports (user_id, requested_at DESC);
```

Invariants: `item_count` = row count (DQ assert nightly); positions are dense `1..n` after every reorder, gaps allowed after removals; `list_items.item.type = lists.type` for new rows (legacy mixed lists keep their rows, `legacy/docs/data-model.md` migration `0005_list_type`).

Migration: lists, list_items (`added_at` := `lists.created_at`, SDD 08 §5), reports (`reason` copied; legacy target ids already text; a `user` target becomes the imported user's numeric id, and reports whose target user or list was not imported are dropped), contact_messages copied 1:1; `list_items.position` renumbered dense `1..n` in legacy order; lists and their rows of users not imported (purged) are skipped; `follows` rows archived in the cutover backup (SDD 11), not imported.

Social tables (`follows`, `activity_events`, the `users.feed_opt_in` column and the new notification kinds) are specified in §12 (Phase 7, after `v1.0.0`).

## 6. Contracts

### 6.1 Authorization

`ListPolicy` generated by `vukem make:policy List`, registered with `authz.For[*models.List]` (<https://vukem.dev/docs/security/authorization> "Configuration"). `View`: allow if owner, else allow only when `is_public && !hidden && !owner.profile_private`, else `authz.DenyAsNotFound()` ("Results: allow, deny or hide"). `Update`, `Delete`, `AddItem`, `Reorder`: owner only, `DenyAsNotFound` for others. Every handler loads the row before `c.Authorize` and index queries filter by `user_id` ("Security"). `ProfilePolicy.ViewDetails` gates profile sections the same way.

### 6.2 Web pages (Inertia)

Rendered with `inertia.Render`; the unread count is a shared prop set by middleware with `inertia.Share(ctx, "unreadNotifications", n)` (<https://vukem.dev/docs/frontend/inertia> "Shared props", "Rendering pages").

| Component | Route | Auth | Props |
| --- | --- | --- | --- |
| `Library/Index` | GET `/library?status=&type=&sort=` | session | `counts{status:{type:n}}`, `groups[{type, items[ItemRef+status+ttb+edition]}]` (`Scroll`, 30/type), `filters` |
| `Lists/Index` | GET `/lists` | session | `lists[{id,type,title,isPublic,hidden,itemCount,coverThumbs[≤4]}]`, `limits` |
| `User/List` | GET `/u/{nickname}/lists/{id}` | public | `list`, `items[{position,note,addedAt,item}]`, `canEdit`, `seo{indexable}` |
| `Lists/Edit` | GET `/lists/{id}/edit` | owner | `list`, `items`, `version` |
| `User/Profile` | GET `/u/{nickname}?tab=ratings\|status\|lists&type=` | public | `profile{nickname,initial,isPrivate,isOwner}`, `stats{ratingsByType,statusByType}`, `ratings` (`Scroll`), `lists` |
| `Account/Notifications` | GET `/me/notifications` | session | `notifications` (`Scroll`, 20) |
| `Contact/Index` | GET/POST `/contact` | public | `topics`, form token |

Web mutations (session + CSRF): `POST /lists`, `PATCH|DELETE /lists/{id}`, `PUT|DELETE /lists/{id}/items/{itemId}`, `PUT /lists/{id}/order`, `PUT|DELETE /items/{id}/edition`, `POST /reports`, `POST /me/notifications/{id}/read`, `POST /me/notifications/read-all`, `POST /profile/export`.

### 6.3 `/app/v1` (A-01; typed handlers via `api.Register`, <https://vukem.dev/docs/rest-apis/typed-handlers>)

| Method, path | Ability | Input → output |
| --- | --- | --- |
| GET `/me/lists` | `catalog:read` | `?type&cursor` → `{data:[ListSummary], meta:{next_cursor}}` |
| GET `/lists/{id}` | `catalog:read` | → `ListDetail{..., version, items[]}` (≤ 500, unpaged) |
| POST `/lists` | `lists:write` | `{type,title,description?,isPublic,firstItemId?}` → 201 |
| PATCH/DELETE `/lists/{id}` | `lists:write` | `{title?,description?,isPublic?}` → 200 / 204 |
| PUT/DELETE `/lists/{id}/items/{itemId}` | `lists:write` | `{note?}` → 200 / 204 |
| PUT `/lists/{id}/order` | `lists:write` | `{version, itemIds[]}` → `{version}` |
| PUT/DELETE `/items/{id}/edition` | `status:write` | `{editionId}` → 200 / 204 |
| GET `/me/library` | `catalog:read` | `?status&type&cursor` → as `Library/Index` groups |
| GET `/users/{nickname}` | `catalog:read` | → public profile (private → nickname only) |
| GET `/me/notifications` | `catalog:read` | `?unread&cursor` → `{data, meta:{next_cursor, unread}}` |
| POST `/me/notifications/read` | `profile:write` | `{ids[≤100]}` or `{all:true}` → 204 |
| POST `/me/export` | `profile:write` | → 202 `{exportId, status}` |
| POST `/reports` | `profile:write` | `{targetType,targetId,reason,details?}` → 201 / 200 |

Writes accept `Idempotency-Key` (SDD 05 §6.4). Problems (kebab-case): `not-found` 404, `list-limit-reached` 409, `list-item-limit-reached` 409, `list-version-conflict` 409, `list-type-mismatch` 422, `order-mismatch` 422, `edition-not-of-work` 422, `report-self` 403, `export-cooldown` 429 (with `Retry-After`), `rate-limited` 429, `bot-check-failed` 422. SDD 05 §6.4 lists these routes with their abilities and points here for details (open issue 14-2).

### 6.4 Admin (`/admin`, <https://vukem.dev/docs/digging-deeper/admin> "Resources")

`vukem make:admin-resource Report` and `ContactMessage`: list/filter by `status`/`handled`, search `details`/`message`. Contact uses plain CRUD (toggle `handled`). Report outcomes are app routes (custom panel actions are [NOT IN VUKEM DOCS], VF-008, as SDD 06): `POST /admin/reports/{id}/resolve {outcome, adminNote?}` → hides the list, bans via SDD 03 §6.3 `POST /admin/users/{id}/ban` (dispatches `user.banned`), or rejects via SDD 04; 409 `report-not-open`. `lists` resource exposes `hidden` only.

### 6.5 Events and internal API

Queue jobs dispatched with `q.DispatchTx` inside the write transaction, so they exist only if it commits (<https://vukem.dev/docs/digging-deeper/queue> "Jobs and transactions"):

- `list.item_added {listId, itemId, userId, addedAt, isPublic}`, `list.item_removed {listId, itemId}`, `list.visibility_changed {listId, visible}`, `list.deleted {listId}`: consumed by SDD 08 (`list_adds`, co-occurrence) and the cache purger.
- `library.export {exportId}` (queue `default`, `queue.Retries(3)`, `queue.Timeout(5*time.Minute)`).

`notifications.Send(ctx, tx, userID, kind, dedupeKey, payload)` (HexaScore package) is the only writer of `user_notifications`; SDD 04 `community.notify` and SDD 02 flag notices call it. Push hook for later: a `notifications.Sent` event, no subscriber in v1.

## 7. Behaviour and algorithms

### 7.1 Lists

- Create: `pg_advisory_xact_lock(hashtext('lists:'||user_id))`, count per type, insert (legacy `legacy/docs/data-model.md`).
- Add: list row `FOR UPDATE` (legacy `lockOwnedList`), check type and count, insert at `max(position)+1`, bump `item_count`, `version`. Remove leaves a gap.
- Reorder (decision: full permutation + dense renumbering, not fractional): lock; stale `version` → 409; `itemIds` ≠ current set → 422 `order-mismatch`; one `UPDATE … FROM unnest($ids) WITH ORDINALITY` (unique deferred); `version+1`. Legacy up/down swap (`actions/lists.ts` `moveInList`) becomes a client permutation.
- Cover: 2×2 grid of the first 4 displayable items' SDD 06 thumbs (no generated image); `og:image` = first item's 600 px cover.
- SDD 08 co-occurrence: only lists `is_public AND NOT hidden` whose owner is not private, banned or anonymised.

### 7.2 Library and profile

Library reads `item_status` (SDD 02) joined to `time_to_beat`, `ratings`, `user_item_editions`, grouped in `ITEM_TYPES` order. Row quick actions: status, stars, time-to-beat (completed games), add-to-list, edition (books), via SDD 02 routes or §6. Sorts: `updated` (default), `title`, `score`. Profile: rating and status counts per type, recent ratings (20/page), public lists. Avatar: nickname initial (legacy `app/[locale]/u/[nickname]/page.tsx`); Google `image_url` not shown (decision: no third-party images on public pages). Unknown nickname → 404.

### 7.3 Reports and contact

Report insert: `ON CONFLICT DO NOTHING` on the partial index. Resolve (one transaction): set status/outcome, apply it, close other open reports on the target alike, `notifications.Send` each reporter (`'report:<id>'`); a hidden list's owner gets `list_hidden`. Contact: ADR 0016 layers (honeypot, 3 s minimum fill, signed form token, PoW on threshold), store, then `mailer.SendLater` with `To = CONTACT_INBOX`, `ReplyTo = sender` (<https://vukem.dev/docs/digging-deeper/mail> "Sending later, from a queue", "The envelope"); missing `CONTACT_INBOX` stores only (legacy).

### 7.4 Data export

1. `POST`: latest non-`failed` `data_exports` row < 7 days old → 429 `export-cooldown`; else insert `queued`, `DispatchTx` `library.export`.
2. The job reads in one repeatable-read transaction (SDD 03 §6.4) and writes a ZIP (Go `archive/zip`) to a private `storage` disk (<https://vukem.dev/docs/digging-deeper/storage> "The local disk"): `profile`, `ratings`, `statuses`, `time_to_beat`, `editions`, `lists`, `suggestions`, `notifications`, `account` (SDD 03 §6.4 rest) as `.json`; `README.md` (per-file schema, 0–10 stars, minutes); `ratings.csv` (`type,title,year,externalIds,overall,c1..c6,ratedAt`) and `lists.csv` (`list,position,type,title,year,note,addedAt`) for Goodreads/Letterboxd-style import elsewhere.
3. CSV cells get the leading-`'` formula guard of `export.CSV` (<https://vukem.dev/docs/digging-deeper/exports> "Quick start"); those helpers write to an `http.ResponseWriter`, a writer-only variant is [NOT IN VUKEM DOCS], so the job applies the rule itself.
4. Mark `ready`, mail the link `storage.TemporaryURL(signer, "/exports", path, 24*time.Hour)`; `storage.Handler(disk, signer)` serves it and answers 403 on expiry (<https://vukem.dev/docs/security/signed-urls> "Temporary file downloads", "Configuration").
5. Hourly scheduler task deletes files past `expires_at` and sets `expired` (`s.Func(...).Hourly().WithoutOverlapping()`, <https://vukem.dev/docs/digging-deeper/scheduler> "Overlapping runs").

### 7.5 Notifications

List by `id DESC` (UUID v7 ≈ time) with cursor pagination (<https://vukem.dev/docs/rest-apis/pagination>); unread count `SELECT count(*) FROM (… LIMIT 100)` on the partial index. Daily 04:30 purge per NFR-14-6.

## 8. Edge cases

- Item merge (SDD 01 §7.5): `list_items` move with `ON CONFLICT DO NOTHING` (target wins), moved rows append at the end, `item_count` recomputed; `user_item_editions` move if the edition belongs to the target work, else dropped.
- Item hidden or taken down (SDD 01 `items.visibility` `hidden`/`takedown`, FR-01-14; open issue 14-3): rows kept; pages and counts skip non-`public` items; setting `public` again brings them back.
- User deletion purge (SDD 03 §7.5): current rule pending the owner's answer to open issue 14-7 (architect default): **public lists are anonymised** (owner shown as "deleted user", description and per-item notes cleared because free text can identify the person; title kept), **private lists are deleted**; `user_item_editions`, notifications and exports deleted; reports by the user keep the row with `reporter_id` NULL; reports **about** the user are closed `no_action`.
- Banned owner: lists stay, `noindex`, excluded from SDD 08.
- Concurrent reorder (web + app): second → 409; client refetches. Concurrent adds at 499: row lock, one 409.
- Reporting a list you cannot see → 404. Export during deletion grace: allowed; purge deletes the file.
- Notification about a deleted item: generic label, no link.

## 9. Acceptance criteria

- `go test -tags=integration ./tests -run 'TestSecurity_A01_ListOwnerOnly|TestListLimits|TestListTypeMismatch'` → pass; non-owner PATCH → 404.
- `go test ./app/library -run TestReorderProperty` → 10,000 random permutations: positions dense, set unchanged, stale version → 409.
- `go test -tags=integration ./tests -run TestExportZip` → ZIP has the 12 files, CSV cells starting `=` are prefixed `'`, second request in 7 days → 429 `export-cooldown`, link after 24 h → 403.
- `go test -tags=integration ./tests -run TestReportResolveNotifies` → one notification per reporter, duplicate open report → 200 and 1 row.
- `curl -s /u/<private>` → 200, body has no counts, `<meta name="robots" content="noindex">`.
- `bunx playwright test specs/social.spec.ts specs/lists.spec.ts specs/contact.spec.ts specs/notifications.spec.ts` → pass, axe 0 serious.
- `vukem openapi:export -prefix /app/v1 | diff - openapi/app-v1.json` → empty after the routes are added.
- `k6 run qa/k6/public-lists.js` → p95 ≤ 400 ms at 30 rps, errors < 0.5 %.

## 10. Test strategy

- Unit: limits, permutation validator, CSV guard, SEO rule. Property: reorder sequences (`testing/quick`).
- Integration (Postgres): policies per route × {owner, other, anonymous, admin}; merge; purge; notification dedupe on job retry; DQ `item_count`.
- Contract: Schemathesis on §6.3. Load: k6 list and profile pages. Security: ZAP on `/u/*`, `/lists/*`, `/contact`; `TestSecurity_A01_*` IDOR on lists, notifications, exports.
- E2E: Playwright ports of `legacy/qa/e2e/specs/social.spec.ts` (privacy test; follow/feed dropped, D-06), `ratings.spec.ts` "status shelves" and "curated list", `admin.spec.ts` "reports a user", `contact.spec.ts`.

## 11. Open issues

1. `user_notifications` DDL. Decided (14-1): this SDD owns it; SDD 04 §5 now cross-references §5.
2. `/app/v1` routes in SDD 05. Decided (14-2): SDD 05 §6.4 now lists `/me/library`, `/users/{nickname}`, `/items/{id}/edition`, `/me/notifications*`, `/reports` and `POST /me/export`.
3. Item visibility. Decided (14-3): SDD 01 `items.visibility` (§8).
4. Export format. Decided (14-4): background ZIP (§7.4); SDD 03 FR-03-16/§6.4 now reference this SDD.
5. Private profile and pages. Decided (14-5): 200 nickname-only with `noindex`; SDD 09 §6.2 updated with `/library`, `/lists`, `/lists/{id}/edit`.
6. Writer-only CSV/JSON helpers for jobs are [NOT IN VUKEM DOCS]. Owner: vukem. Closes with a `vukem-feedback.md` answer.
7. Lists of a deleted user. Owner question (asked with the SDD 13 batch). Current rule until answered: anonymise public lists, delete private ones (§8). Closes with the owner's answer.
8. Lawful basis for the feed (Q-A5 a). Owner question, with legal review (D-21). D-29 covers aggregates and personalisation, not showing a user's activity to followers. Current rule until answered: the whole social layer is off (feature flag `social` off, §12.3), and when it is on, a user's activity reaches followers only after an explicit opt-in (`users.feed_opt_in`, default `false` for every account). Closes with a D-xx in `docs/brainstorming.md`; if the answer is "public profile implies feed", the opt-in column becomes a default-on opt-out in a later migration.
9. Legacy follows (Q-A5 b). Owner question. Current rule: Phase 7 starts with an empty graph; the legacy rows stay in the cutover archive (§5 migration note, SDD 11 §7.6). Closes with the owner's answer; a restore, if chosen, is a one-off command that inserts `accepted` rows only between users who still exist and are not banned.
10. Markdown subset for reviews (P7.B.1.1). Open: `go.mod` has no Markdown renderer and <https://vukem.dev/docs> names none [NOT IN VUKEM DOCS]. Default: a stdlib-only subset (bold, italic, links, line breaks, spoiler block) rendered through `html/template`; a library needs the owner's yes and an ADR. Closes in P7.B.1.1.
11. Letterboxd reference (Q-A9). The research report is partial; §12 does not depend on it. Closes with the owner's answer.

## 12. Social layer

Phase 7 (ROADMAP P7.A–P7.C, D-06 as amended 2026-10-31). This section covers follows and the activity feed (P7.A); reviews, comments and reactions (P7.B) add subsections here in their own microtasks. Outline: `.orchestration/reports/AUDIT-social-sdd.md`.

### 12.1 Scope

In: follow and unfollow, follow requests for private profiles, follower and following counts, the `/feed` page (followed users, newest first), the "recent activity" tab of the public profile. Out: DMs, groups, blocking (later row), push, social data in `/api/v1` (A-07), social routes in `/app/v1` (SDD 10 keeps namespace `social` out of the Android bundle until a mobile row adds it).

### 12.2 Functional requirements

- FR-14-13: `PUT /u/{nickname}/follow` creates a follow; repeating it is a no-op (200 with the current state). Following a public profile gives `accepted`; following a private profile gives `pending` and notifies the followee (`follow_request`). `DELETE /u/{nickname}/follow` removes the row in any state (unfollow or cancel a request). Following yourself is 422 `cannot-follow-self`.
- FR-14-14: The followee accepts or declines a pending request from `/me/follow-requests`; accept sets `accepted` (no notice to the follower in v1); decline deletes the row. Only `accepted` rows count anywhere.
- FR-14-15: The public profile shows follower and following counts and a follow button. A private profile keeps FR-14-8: no counts, only nickname and initial, and the follow button (which sends a request).
- FR-14-16: `/feed` lists activity of users the viewer follows with `accepted` state, 20 per page, newest first, cursor pagination (SDD 09 §7.3). Event kinds: `rating_saved`, `status_changed`, `list_created`, `list_item_added`, `review_published` (P7.B).
- FR-14-17: The public profile gains a "recent activity" tab with the same cards, restricted to the profile owner's events and the privacy matrix (§12.5).
- FR-14-18: A user's events are recorded and shown to others only while `users.feed_opt_in` is true (open issue 8). The opt-in is offered at onboarding and in settings, off by default; turning it off hides past events at once (read-time filter) and stops recording new ones.

### 12.3 Feature flag and non-functional requirements

- Feature flag `social` (<https://vukem.dev/docs/digging-deeper/feature-flags>, env `FLAG_SOCIAL`, code default off, same pattern as SDD 04 `community_submissions`): when off, the social routes answer 404, no activity event is written, and the profile renders without counts or the follow button. It stays off in production until open issue 8 is answered and the privacy policy is updated (D-21).
- NFR-14-7: `/feed` p95 ≤ 300 ms at 30 rps with 10k events (k6 `qa/k6/feed.js`, P7.A.1.3); at most 3 queries per page.
- NFR-14-8: Rate limits (`app/http/limits`): `follows` 60/h per user (`ratelimit.ByUser`); 429 `rate-limited`.
- NFR-14-9: Retention: `activity_events` older than 180 days deleted daily at 04:40 (the same scheduler as NFR-14-6).

### 12.4 Data model

```sql
ALTER TABLE users ADD COLUMN feed_opt_in boolean NOT NULL DEFAULT false;

CREATE TABLE follows (
  follower_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  followee_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  state text NOT NULL CHECK (state IN ('accepted','pending')),
  created_at timestamptz NOT NULL DEFAULT now(),
  accepted_at timestamptz,
  PRIMARY KEY (follower_id, followee_id),
  CHECK (follower_id <> followee_id));
CREATE INDEX follows_followee_idx ON follows (followee_id, state);

CREATE TABLE activity_events (
  id uuid PRIMARY KEY,                                   -- v7, the cursor
  actor_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  kind text NOT NULL CHECK (kind IN ('rating_saved','status_changed','list_created',
       'list_item_added','review_published')),
  item_id uuid REFERENCES items(id) ON DELETE CASCADE,
  list_id uuid REFERENCES lists(id) ON DELETE CASCADE,
  payload jsonb NOT NULL DEFAULT '{}',                   -- stars, status; ids only, never text
  created_at timestamptz NOT NULL DEFAULT now());
CREATE INDEX activity_events_actor_idx ON activity_events (actor_id, id DESC);
CREATE INDEX activity_events_created_idx ON activity_events (created_at);
```

`user_notifications.kind` gains `new_follower` and `follow_request` (dedupe key `follow:<follower_id>`). Rapid repeats collapse: a rating changed twice within 10 minutes updates the actor's last `rating_saved` event for that item instead of inserting a second one (ADR 0020).

### 12.5 Privacy matrix

Who sees what (viewer in rows). "Activity" means feed entries and the profile activity tab; "counts" means follower and following counts.

| Viewer \ target | Public, opted in | Public, not opted in | Private | Banned or anonymised |
| --- | --- | --- | --- | --- |
| Anonymous | counts, activity tab | counts | nickname only (FR-14-8) | nothing (404 profile per SDD 03) |
| Member, not following | counts, activity tab, follow | counts, follow | nickname, request button | nothing |
| Accepted follower | counts, activity tab, feed | counts | counts, activity tab, feed if opted in | nothing |
| Owner | everything | everything | everything | n/a |

Rules that apply to every cell: events on a list that is private or `hidden`, or on an item whose visibility is not `public` (SDD 01), are dropped at read time; counts never include private, banned or anonymised users on the other side of the edge; a profile turning private keeps its `accepted` followers (Instagram and Goodreads model) and new follows become requests.

### 12.6 Contracts

| Route | Auth | Response |
| --- | --- | --- |
| `PUT /u/{nickname}/follow` | session | 303 back; 404 unknown or banned target; 422 `cannot-follow-self`; 429 |
| `DELETE /u/{nickname}/follow` | session | 303 back (also when no row existed); 404 unknown user |
| `GET /me/follow-requests` | session | Inertia `Social/Requests` |
| `POST /me/follow-requests/{nickname}/accept` and `/decline` | session | 303 back; 404 when no pending row |
| `GET /feed?cursor=` | session | Inertia `Feed/Index` with `events`, `nextCursor` |
| `GET /u/{nickname}?tab=activity` | public | the profile page with `activity` deferred in group `personal` |

Following a banned user is 404, like a missing one; accept and decline match only the signed-in followee's own pending rows, so no policy object is needed. Follow writes run in one transaction with `notifications.Send`. Routes live in `routes/follows.go` (`routes/social.go` is the OAuth sign-in file). The public profile gets a `social` prop (`followers`, `following` unless hidden; `state` for a signed-in non-owner).

### 12.7 Feed query

ADR 0020: events are written on the write path (no fan-out); the feed reads `activity_events` of followed actors with a keyset cursor:

```sql
SELECT e.* FROM activity_events e
JOIN follows f ON f.followee_id = e.actor_id AND f.follower_id = $viewer AND f.state = 'accepted'
JOIN users u ON u.id = e.actor_id AND u.feed_opt_in AND u.banned_at IS NULL AND u.anonymized_at IS NULL
WHERE ($cursor IS NULL OR e.id < $cursor)
ORDER BY e.id DESC LIMIT 21;
```

The visibility filters of §12.5 apply to the joined items and lists in the same query.

### 12.8 Edge cases

- Item merge (SDD 01 §7.5): `activity_events.item_id` moves to the target.
- User deletion: `ON DELETE CASCADE` removes follows and events; anonymisation (open issue 7) deletes them explicitly in the purge.
- Ban: the banned user's edges stay but are filtered out of counts and feeds; unban restores them.
- Opt-out after events were shown: read-time filter, no backfill delete; retention removes them within 180 days.

### 12.9 LGPD

Basis: open issue 8 (Q-A5). The data export ZIP (§7.4) gains `follows.csv` (both directions, nicknames) and `activity.csv`. The purge (SDD 03 §7.5) covers both tables.

### 12.10 Tests

`TestFollows`, `TestSecurity_A01_PrivateProfileFollowRequiresApproval`, `TestSecurity_A06_HexaLimits` (rule `follows`), `TestFeedPrivacy`, `TestSocialDisabledByDefault` (routes 404 and no events with the flag off), k6 `qa/k6/feed.js`, Playwright `social.spec.ts`.
