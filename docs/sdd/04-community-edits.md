# SDD 04: Community edits

Status: reviewed (2026-10-28 consistency pass). Decisions: D-01, D-03, D-08, D-15, D-17, requirement 5. ADRs: 0010. Related SDDs: 01, 03, 05, 06, 07, 09, 12, 13.

## 1. Objective

Let any signed-in user correct or enrich the catalogue (items, editions, entities per D-01) through admin-reviewed suggestions, never direct edits (research 9.4: TMDB's direct edits need many moderators, IMDb's 1–7 day review frustrates). Goals: reliable data for score and API, a decision within 24 h, append-only history with one-click revert, and contributor accuracy tracked from day one so trusted moderators need no data migration later (D-15).

## 2. Scope

In scope (v1): field edits on items, editions and entities (title/translations, synopsis with source, release date, genres, creators and other relations, external IDs, cover via upload or URL); create requests; merge/duplicate reports; lifecycle, quota, anti-abuse, admin queue, application, `catalog_audit`, revert, status page, "Recent changes" tab, notifications, contributor stats, metrics, i18n en/pt-BR/es (D-17); admin locks.

Out of scope (v1): trusted moderators (modelled, flag off, D-15); suggestion objects, queues and contributor data in the paid API `/api/v1` (applied `source=community` field values do leave through the API under A-07 and the API terms, open issue 04-1); editing from mobile (D-12; `/app/v1` may submit later, A-01); reviews, comments, badges, leaderboards; automatic merge revert.

## 3. Functional requirements

- FR-04-1: A user with a verified email and an account ≥ 7 days old can draft a suggestion on any unlocked record. Verify: integration test per eligibility branch.
- FR-04-2: A suggestion holds `changes` (before/after per field), note ≤ 1,000 chars, locale and 0–3 source URLs; factual fields (7.2) require ≥ 1 URL. Verify: validator unit tests.
- FR-04-3: Submit answers 429 when the author already submitted 3 suggestions (every non-blocked user at launch, D-31; 7.3) in the current UTC day; withdrawn ones count. Verify: integration test with a fake clock across 00:00 UTC.
- FR-04-4: Only transitions in 7.1 are allowed; others answer 409 `invalid-transition`. Verify: table test over all 81 state pairs.
- FR-04-5: Each transition writes one `suggestion_events` row in the same transaction. Verify: integration test.
- FR-04-6: Submit answers 409 `suggestion-duplicate` if the author has an open suggestion on the same field, or anyone has one with an identical `after`. Verify: integration test.
- FR-04-7: Submit or approve on a locked field/record answers 423 `field-locked`. Verify: integration test.
- FR-04-8: The admin queue filters (status, kind, target type, domain, field, locale, author, assignee, age bucket, competing, stale), shows a diff, supports claim/release, approve, reject and bulk actions (≤ 50). Verify: Playwright.
- FR-04-9: Approval writes the catalogue, provenance `source=community`, one audit row per field and the `catalog.changed` job in one transaction. Verify: integration test with injected failure.
- FR-04-10: `catalog_audit` rejects UPDATE/DELETE in the database. Verify: integration test expects SQLSTATE `P0001`.
- FR-04-11: Revert restores `before`, appends a new audit row and marks the suggestion `reverted`, with the conflict rule in 7.5. Verify: integration test per branch.
- FR-04-12: The author (and admins) see a status page with state, timeline, reason and diff; others get 404. Verify: integration + Playwright.
- FR-04-13: Every item, edition and entity page has a public, SSR "Recent changes" tab, 20 entries per page. Verify: Playwright + axe.
- FR-04-14: Decisions, reverts and expiries create an in-app notification and an email in the author's locale (email opt-out). Verify: integration test with `MAIL_MAILER=array`.
- FR-04-15: Profiles show submitted/approved/rejected counts and approval rate after ≥ 1 decision; admins export `catalog_audit` as CSV/NDJSON. Verify: integration tests (export of 100k rows streams).
- FR-04-16: The moderation dashboard shows queue size, oldest open age, p50/p95 time-to-decision and approval rate (7/30 days). Verify: integration test on the query.
- FR-04-17: Moderation has two steps (owner Q-A15, 2026-10-31). A moderator only reviews: it records a recommendation (`approve` or `reject`, optional note ≤ 1,000 chars) on an open suggestion, stored per suggestion and reviewer with the time and the `version` reviewed, and shown in the queue and the detail view. Only an admin approves (applies to the catalogue), rejects for good, bulk-approves or bulk-rejects; merge suggestions apply through approve, so merge is admin-only too. Verify: `TestSecurity_A01_ModeratorCannotApply` and the recommendation integration tests.

## 4. Non-functional requirements

- NFR-04-1: Submit p95 < 300 ms; queue page p95 < 500 ms with 10,000 open suggestions; apply p95 < 200 ms (excluding cover processing); Recent changes p95 < 150 ms server time. Verify: k6 on staging.
- NFR-04-2: SLA: 95 % of submissions decided within 24 h; ageing badge green < 12 h, amber 12–24 h, red > 24 h; alert when the oldest open suggestion passes 20 h (SDD 12).
- NFR-04-3: Audit retained forever, written in the same transaction as the change, covered by daily backups (D-18).
- NFR-04-4: CSRF on forms, policies fail closed (vukem `https://vukem.dev/docs/security/authorization` "Authorization"), diffs rendered as text, user URLs fetched only via `httpclient.Safe` (vukem `https://vukem.dev/docs/basics/csrf` A01). Verify: `TestSecurity_*` (section 10), ZAP.
- NFR-04-5: Public surfaces show nickname only; account deletion (SDD 03) nulls actor/author ids and shows "deleted user"; values stay.
- NFR-04-6: WCAG 2.2 AA; axe zero serious/critical; diffs use text markers, not colour alone.
- NFR-04-7: Strings under `community.*` via vukem i18n (`https://vukem.dev/docs/digging-deeper/localization` "i18n"); CI fails on a missing key in any locale.
- NFR-04-8: No new service or store (D-08): Postgres plus the database queue.

## 5. Data model

User references use the type of `users.id`. In this app that column is `bigint` (`t.ID()` in `20261029022752_create_users_table`, same as `ratings.user_id`). Row ids of the tables below stay as written.

```sql
create type suggestion_kind   as enum ('edit','create','merge');
create type suggestion_target as enum ('item','edition','entity');
create type suggestion_status as enum ('draft','submitted','in_review','approved',
  'applied','rejected','reverted','withdrawn','expired');
create type reject_reason as enum ('incorrect','no_source','bad_source','duplicate',
  'policy','spam','stale','target_gone','other');
create type actor_role as enum ('author','admin','moderator','system');

create table suggestions (
  id            uuid primary key,                      -- UUIDv7
  author_id     bigint references users(id) on delete set null,
  kind          suggestion_kind   not null,
  target_type   suggestion_target not null,
  target_id     uuid,                                  -- null only for 'create'
  target_domain text not null check (target_domain in ('game','movie','series','book')),
  locale        text not null check (locale in ('en','pt-BR','es')),
  changes       jsonb not null,
  note          text check (char_length(note) <= 1000),
  source_urls   text[] not null default '{}' check (cardinality(source_urls) <= 3),
  status        suggestion_status not null default 'draft',
  assigned_to   bigint references users(id) on delete set null,
  claimed_at    timestamptz,
  decided_by    bigint references users(id) on delete set null,
  decided_at    timestamptz,
  reject_reason reject_reason,
  decision_note text check (char_length(decision_note) <= 1000),
  submitted_at  timestamptz,
  submitted_day date generated always as ((submitted_at at time zone 'UTC')::date) stored,
  depends_on    uuid references suggestions(id),
  meta          jsonb not null default '{}',           -- source_status, flags
  version       integer not null default 1,            -- optimistic lock
  created_at    timestamptz not null default now(),
  updated_at    timestamptz not null default now(),
  check ((kind = 'create') = (target_id is null))
);
create index suggestions_queue_idx  on suggestions (submitted_at) where status in ('submitted','in_review');
create index suggestions_quota_idx  on suggestions (author_id, submitted_day);
create index suggestions_target_idx on suggestions (target_type, target_id, status);
create index suggestions_changes_gin on suggestions using gin (changes);

create table suggestion_events (
  id            bigint generated always as identity primary key,
  suggestion_id uuid not null references suggestions(id),
  from_status   suggestion_status,
  to_status     suggestion_status not null,
  actor_id      bigint references users(id) on delete set null,
  actor_role    actor_role not null,
  reason        text,
  meta          jsonb not null default '{}',           -- bulk batch id, request id
  created_at    timestamptz not null default now()
);
create index suggestion_events_sid_idx on suggestion_events (suggestion_id, id);

create table suggestion_reviews (                     -- FR-04-17: a moderator's recommendation
  suggestion_id      uuid not null references suggestions(id),
  reviewer_id        bigint not null references users(id) on delete cascade,
  recommendation     text not null check (recommendation in ('approve','reject')),
  note               text check (char_length(note) <= 1000),
  suggestion_version integer not null,                 -- the version the reviewer saw
  created_at         timestamptz not null default now(),
  updated_at         timestamptz not null default now(),
  primary key (suggestion_id, reviewer_id)              -- a new recommendation replaces the reviewer's last one
);

create type audit_source as enum ('community','admin','revert','merge','create');

create table catalog_audit (
  id               bigint generated always as identity primary key,
  occurred_at      timestamptz not null default now(),
  actor_id         bigint references users(id) on delete set null,
  actor_role       actor_role not null,
  source           audit_source not null,
  suggestion_id    uuid references suggestions(id),
  target_type      suggestion_target not null,
  target_id        uuid not null,
  field            text not null,                      -- 'translations.pt-BR.title'
  before           jsonb,                              -- null = absent
  after            jsonb,
  prev_provenance  jsonb,                              -- restored on revert
  reason           text,
  reverts_audit_id bigint references catalog_audit(id),
  request_id       text
);
create index catalog_audit_target_idx on catalog_audit (target_type, target_id, occurred_at desc);
create index catalog_audit_field_idx  on catalog_audit (target_type, target_id, field, id desc);
create index catalog_audit_actor_idx  on catalog_audit (actor_id, occurred_at desc);
create unique index catalog_audit_one_revert on catalog_audit (reverts_audit_id)
  where reverts_audit_id is not null;
-- trigger: BEFORE UPDATE OR DELETE FOR EACH STATEMENT → RAISE EXCEPTION 'catalog_audit is append-only' (P0001);
-- statement level so the write is refused even when zero rows match (the acceptance updates id 1 on an empty table).
-- app role has only INSERT/SELECT on it (SDD 11).

create type contributor_tier as enum ('new','standard','low','blocked');

create table contributor_stats (
  user_id          bigint primary key references users(id) on delete cascade,
  submitted_total  integer not null default 0,
  approved_total   integer not null default 0,          -- reached 'applied'
  rejected_total   integer not null default 0,          -- excludes duplicate/stale/target_gone
  reverted_total   integer not null default 0,          -- revert reason 'incorrect' only
  withdrawn_total  integer not null default 0,
  accuracy         numeric(4,3) generated always as ((approved_total + 1)::numeric
                     / (approved_total + rejected_total + reverted_total + 2)) stored,
  tier             contributor_tier not null default 'new',
  cooldown_until   timestamptz,
  blocked_until    timestamptz,
  trusted_eligible boolean not null default false,      -- D-15, read later
  updated_at       timestamptz not null default now()
);

create table catalog_locks (
  id uuid primary key, target_type suggestion_target not null, target_id uuid not null,
  field text not null default '*', reason text not null,
  locked_by bigint references users(id) on delete set null,
  expires_at timestamptz, created_at timestamptz not null default now(),
  unique (target_type, target_id, field)
);

-- user_notifications: owned by SDD 14 §5 (text user id, dedupe_key); this domain writes it only through
-- notifications.Send (SDD 14 §6.5, §7.5). Open issue 14-1.
```

Invariants: `submitted_at` set once; `applied` implies ≥ 1 audit row; `contributor_stats` rebuildable from suggestions and events (`community:stats:rebuild`). Retention: suggestions, events, audit forever; notification retention per SDD 14 (NFR-14-6).

Legacy: no suggestion or catalogue-audit tables exist. `reports` (`legacy/apps/web/lib/db/schema/social.ts`: targets `list`/`user`, statuses `open/resolved/dismissed`) is the only moderation primitive and stays separate (SDD 13); its "one open report per reporter and target" partial unique index is the precedent for FR-04-6. Legacy admin edits and `mergeItems` (`legacy/apps/web/actions/admin.ts`, `legacy/apps/web/lib/admin.ts`) left no history, so `catalog_audit` starts empty at cutover.

## 6. Contracts

Locale prefixes per D-17. User routes: `web` guard + `auth.Verified` (vukem `https://vukem.dev/docs/security/authentication` "Middleware"). Inertia requests get validation errors as props and other errors through the `Error` page; JSON clients get RFC 9457 problem+json (vukem `https://vukem.dev/docs/basics/middleware`, content negotiation). Problem base: `https://hexascore.app/problems/`.

| Method | Path | Auth | Notes |
| --- | --- | --- | --- |
| GET | `/suggest/{target_type}/{id}`, `/suggest/new/{domain}` | verified | Inertia `Community/SuggestEdit` / `SuggestCreate` (target snapshot, editable fields, locks, quota left) |
| POST / PATCH | `/suggestions`, `/suggestions/{id}` | author, `draft` | save draft → 303 |
| POST | `/suggestions/{id}/submit` | author | 303 to status page, or 403/409/422/423/429 |
| POST | `/suggestions/{id}/withdraw` | author | → 303 |
| POST | `/suggestions/uploads/cover` | draft owner | multipart, 1 file → `{key,width,height}` |
| GET | `/suggestions/{id}`, `/me/suggestions` | author/admin, else 404 | `Community/SuggestionStatus`, `Community/MySuggestions` |
| GET | `/{type}/{slug}/changes`, `/entities/{kind}/{slug}/changes` | public | `Catalog/RecentChanges`, SSR (D-07) |
| GET | `/admin/moderation`, `/admin/moderation/{id}`, `/admin/moderation/metrics` | `moderate` | queue, detail with diff and recommendations, FR-04-16 |
| POST | `/admin/moderation/{id}/{claim,release}` | `moderate` | `version` |
| POST | `/admin/moderation/{id}/recommend` | `moderate` | `version`, `recommendation` = approve, reject; `note?` (FR-04-17) |
| POST | `/admin/moderation/{id}/{approve,reject}` | `applySuggestion` | `version`; approve `confirm_stale?`, `confirm_title` on merge; reject `reason`, `note` |
| POST | `/admin/moderation/bulk` | `moderate`; `applySuggestion` for approve and reject | `ids[]` ≤ 50, action → per-id results |
| GET | `/admin/catalog-audit`, `/admin/catalog-audit/export.{csv,ndjson}` | `viewCatalogAudit` | filters; export via vukem `export` (`https://vukem.dev/docs/digging-deeper/exports` "Exports") |
| POST | `/admin/catalog-audit/{id}/revert` | `revertCatalog` | `reason`, `mode` = single, cascade, force |
| POST / DELETE | `/admin/locks`, `/admin/locks/{id}` | `lockCatalog` | target, field, reason, expiry |

Moderation pages (`/admin/moderation*`, `/admin/catalog-audit*`, `/admin/locks`) are server-rendered `html/template` views inside the vukem admin layout (<https://vukem.dev/docs/digging-deeper/admin>, <https://vukem.dev/docs/basics/views>), not Inertia pages, so `/admin` stays one UI (open issue 09-12); forms post with CSRF tokens and answer 303. The ban outcome calls SDD 03 `POST /admin/users/{id}/ban`.

Submission kill switch: feature flag `community_submissions` (<https://vukem.dev/docs/digging-deeper/feature-flags>, names allow only `a-z`, `0-9` and `_`; env `FLAG_COMMUNITY_SUBMISSIONS`), code default off, turned on in production only once moderation is staffed (an admin able to meet the 24 h decision target, D-15). While off, the suggest pages show a "coming soon" notice, `POST /suggestions*` and `/submit` answer 403 `community-submissions-closed`, and the moderation queue stays reachable for admins.

Abilities are gate abilities defined like `admin`/`viewAuditLogs` in vukem `https://vukem.dev/docs/digging-deeper/admin` "Wiring". `moderate` (queue, detail, claim, release, recommend) also goes to role `moderator` only when flag `community_trusted_moderators` is on (default off, `/admin/_flags`, `https://vukem.dev/docs/digging-deeper/admin` "Feature flags"). `applySuggestion` (approve, reject, merge apply), revert, locks and the audit log stay admin-only whatever the flag says (FR-04-17, owner Q-A15). Every role behind these abilities needs 2FA (SDD 13 FR-13-5).

| Status | Type | Extra members |
| --- | --- | --- |
| 403 | `suggestion-not-eligible` (unverified, < 7 days, blocked) | `eligibleAt` |
| 409 | `invalid-transition` | `from`, `to` |
| 409 | `suggestion-duplicate` | `existingId` (own suggestion only) |
| 409 | `stale-before` (value or `version` changed) | `field`, `current` |
| 409 | `revert-conflict` | `laterAuditIds` |
| 422 | `validation` (schema, URL, honeypot) | `errors` |
| 423 | `field-locked` | `field`, `lockReason` |
| 429 | `suggestion-quota` | `limit`, `used`, `resetAt`; `Retry-After` = seconds to 00:00 UTC |
| 429 | `suggestion-cooldown` | `cooldownUntil`; `Retry-After` |

Burst 429s and upload 413/422 come from vukem `ratelimit` and `upload`.

Jobs (vukem `https://vukem.dev/docs/digging-deeper/queue` "Queue", "Events", "Mail"):

- `catalog.changed {target_type, target_id, fields, audit_ids}` via `q.DispatchTx` in the apply/revert transaction; handlers purge cache (SDD 01), reindex search (SDD 07), update sitemap `lastmod` and the changes feed (SDD 05); idempotent by `audit_ids`. Implementation: one transaction inserts the ids into `catalog_changed_done` and stops when none is new, queues the owning item or entity in `search_dirty` (an edition queues its work) and sets its `updated_at`, which the sitemap prints as `<lastmod>`. The changes feed row is already written by the apply transaction, and the page caches are query-keyed with TTLs of 60 s or less, so there is no per-item cache entry to purge.
- `community.notify`: calls SDD 14 `notifications.Send` (dedupe key `suggestion:<id>:<state>`), sends mail via `SendLaterTx` with `EnvelopeContext` in the author's locale.
- `community.check_sources`: GET each URL through `httpclient.Safe` (5 s, 3 redirects, 256 KiB), stores status in `meta`; never blocks.
- `community.cover_fetch`: downloads a URL cover via `httpclient.Safe`, then `upload.Save` with the cover rules.
- Hourly: expire drafts, refresh tiers, release claims idle > 2 h; notification purge is SDD 14's job.

No OpenAPI operation or SDK in v1; later the same handlers mount at `/app/v1/suggestions` with ability `suggestions:write` (A-01).

## 7. Behaviour and algorithms

### 7.1 State machine

| From | To | Trigger | Who |
| --- | --- | --- | --- |
| — | draft | create | author |
| draft | submitted | submit (eligibility, quota, validation) | author |
| draft | withdrawn | discard | author |
| draft | expired | 14 days without edit | system |
| submitted | in_review | claim | admin (moderator if flag on) |
| submitted, in_review | (same) | recommend approve or reject, no transition and no `version` bump (`suggestion_reviews`) | moderator or admin |
| in_review | submitted | release, or claim idle > 2 h | admin, system |
| submitted, in_review | approved | approve with deferred work (cover processing, unapplied `depends_on`) | admin |
| submitted, in_review, approved | applied | apply transaction commits | system |
| approved | in_review | apply failed permanently | system |
| submitted, in_review | rejected | reject with reason | admin |
| submitted, in_review | withdrawn | withdraw | author |
| submitted, in_review | expired | target deleted or merged (`target_gone`) | system |
| applied | reverted | revert | admin |

Terminal: rejected, reverted, withdrawn, expired. "Clone as new draft" (admin) copies a terminal suggestion to its author without using quota. Approve without deferred work records `approved` and `applied` events in one transaction. A `version` mismatch on any moderator action answers 409 `stale-before`.

### 7.2 Suggestion object

```json
{
  "kind": "edit", "target_type": "item", "target_id": "0192…", "locale": "pt-BR",
  "changes": {
    "translations.pt-BR.title": {"before": null, "after": "O Senhor dos Anéis"},
    "release_date": {"before": {"value": "2001-12-19", "precision": "day"},
                     "after":  {"value": "2001-12-18", "precision": "day"}},
    "genres": {"add": ["ent_fantasy"], "remove": ["ent_drama"]},
    "creators": {"add": [{"entity_id": "0193…", "role": "director"}], "remove": []}
  },
  "source_urls": ["https://www.wikidata.org/wiki/Q127367"],
  "note": "Date per Wikidata and the distributor's press release."
}
```

Scalars carry `before`/`after`; relations carry idempotent `add`/`remove`. Fields come from an allowlist per target (SDD 01); an empty effective change is a 422.

| Field | Validation | Source |
| --- | --- | --- |
| `title`, `translations.<l>.title` | 1–300 chars, NFC, no control chars | required for `title` |
| `synopsis.<l>` | 50–2,000 chars plain text, contributor's own words | required |
| `release_date` | ISO date + precision, 1870 to now + 10 y | required |
| relations (genres, creators, organizations, franchise, platforms) | existing entities of the right kind/domain, ≤ 20 ops | required for creators/organizations |
| `external_ids.<p>` | Wikidata `Q\d+`, TMDB/IGDB digits, Open Library `OL\d+[WMA]`, ISBN-13 checksum | required |
| `cover` | upload key or https URL; JPEG/PNG/WebP ≤ 5 MiB, short side ≥ 300 px | optional |
| `create` / `merge` | required fields of the kind (SDD 01) / same type and domain | required / optional |

Source URLs: https, ≤ 2,048 chars, not hexascore.app, not on a shortener denylist. Covers use `upload.Images(5 << 20)` (vukem `https://vukem.dev/docs/digging-deeper/uploads` "Uploads"; AVIF is undetectable there, so refused), stored under `community/covers/`, resized by the SDD 06 cover pipeline at apply. At submit each scalar `before` must equal the current value, else 409 `stale-before`.

### 7.3 Quota and anti-abuse

Submit transaction: lock `contributor_stats` row `FOR UPDATE` (serialises one user's concurrent submits), count rows with `submitted_day` = current UTC date, compare with the tier limit. Counting happens at submission; withdrawn and expired submissions still count (no submit/withdraw cycling). Refused submits consume nothing.

| Tier | Rule (recomputed on each decision and hourly) | Limit/day |
| --- | --- | --- |
| new | 0 applied | 3 |
| standard | ≥ 1 applied, and decided < 10 or accuracy ≥ 0.40 | 3 |
| low | decided ≥ 10 and accuracy < 0.40 | 3 |
| blocked | decided ≥ 20 and accuracy < 0.20, or admin block | 0 for 30 days |

At launch every non-blocked tier gets 3/day (D-31); tiers are kept for stats, the moderation queue and `blocked`, and a future trusted-moderator quota needs a new decision. `decided` = approved + rejected + reverted; accuracy is Laplace-smoothed (section 5). `trusted_eligible` = approved ≥ 50, accuracy ≥ 0.90, no `spam` rejection in 90 days.

- Eligibility: `auth.Verified`, account ≥ 7 days, not banned (SDD 03).
- Cooldown: 3 rejections for `incorrect`/`no_source`/`bad_source`/`policy`/`spam` within 24 h → 48 h; any `spam` → 7 days.
- Duplicates: under `pg_advisory_xact_lock` on the target, query open suggestions with `changes ? field`. Different `after` values from different authors are allowed and flagged "competing"; applying one flags the rest stale.
- Locks: `catalog_locks` (field or `*`) block drafts, submits and approvals; expired rows ignored.
- Honeypot: hidden `website` field; filled → generic 422 + security event `community.honeypot`; 3 in 7 days → `blocked`.
- Burst (`ratelimit.ByUser` behind `ByIP`, vukem `https://vukem.dev/docs/security/rate-limiting` "Rate limiting"): drafts 30/h, submit/withdraw 10/min, cover uploads 10/h; `RATE_LIMIT_STORE=cache` on the database cache.
- Licensed display-only values (D-03, TMDB/IGDB) are suggestible, but the community value is stored as its own provenance, never relabelled third-party data. Connectors (SDD 06) skip `source=community` fields for 180 days after the audit row and open a `system` suggestion when the source value diverges (open issue 04-3); after 180 days a divergence still opens a `system` suggestion instead of overwriting.

### 7.4 Moderation and apply

Queue: oldest first; age, kind, target, fields, author tier/accuracy, competing/stale/unreachable badges, assignee. Diff per field: current, `before`, `after` (word diff for text, chips for relations, covers side by side). Bulk actions use one transaction per suggestion and report per-id outcomes.

Apply (`db.Tx`, vukem `https://vukem.dev/docs/database/orm` "Query builder"):

1. Lock suggestion `FOR UPDATE`; check status and `version`.
2. Lock target rows `FOR UPDATE` (ordered by id, as legacy `mergeItems`).
3. Per field: if current ≠ `before` and no `confirm_stale` → 409; write `after`; set provenance to `community` (the SDD 01 `provenance[field]` / `*_source` shape; the `data_sources` row is inserted on first apply). The suggestion id is the audit row's `suggestion_id`, not part of the provenance value. Edition fields keep the row-level `source`.
4. Insert one audit row per field with the actual previous value and provenance.
5. Update suggestion, events, `contributor_stats`.
6. `DispatchTx` `catalog.changed` and `community.notify`.

Applies on one target serialise on the submit advisory lock (`pg_advisory_xact_lock` on the target) taken before step 1, so flagging rivals stale (`meta.stale`) cannot deadlock against a concurrent apply. Step 5 sets `meta.stale` on the other open suggestions of the target that touch an applied field. A cover change or an unapplied `depends_on` stops at `approved` with no catalogue write.

Admin-panel catalogue edits also write `catalog_audit` (`source=admin`, `actor_role=admin`) in the same transaction as the write. They go through `POST /admin/catalog/{type}/{id}/fields` (admin-only, `applySuggestion`), which takes the per-target advisory lock and the target row lock, reads each field's `before` under that lock, writes through the apply path (provenance `hexascore`, `api_changes`, `catalog.changed`) and skips unchanged fields. The panel resources stay read-only: a resource `BeforeSave`/`AfterSave` hook (`https://vukem.dev/docs/digging-deeper/admin` "Resources") runs around `orm.Update`, which rewrites every column and would clobber a concurrent apply's provenance (open issue 4, VF-010). `admin.AuditLog` keeps logging panel CRUD; only `catalog_audit` feeds Recent changes and revert.

### 7.5 Revert

On audit row A for field f:

- `single`: if current = `A.after`, write `A.before`, restore `A.prev_provenance`, append R (`source=revert`, `reverts_audit_id=A`, `before=A.after`, `after=A.before`). If a later row changed f, 409 `revert-conflict`.
- `cascade`: revert later non-reverted rows on f newest first, then A, in one transaction; linked suggestions become `reverted`.
- `force`: set `A.before` regardless; R's `before` is the current value; later suggestions stay `applied`.
- Reverting a suggestion reverts all its rows together. A revert cannot be reverted (unique index); restoring requires a new suggestion. Only reason `incorrect` counts against the author.

Implementation: revert is admin-only (`revertCatalog`; the service also refuses a non-admin actor), takes the per-target advisory lock of apply, and writes through `catalog.RestoreField` (the value and the exact `prev_provenance`; a null provenance drops the key) or `catalog.RestoreRelation` (relation snapshots carry each row's `Source`, so a removed relation comes back with its own source). Rows with `source` `revert`, `merge` or `create` answer 409 `not-revertible`; a row already reverted answers 409 `already-reverted` (also on the unique-index race). `single` answers 409 `revert-conflict` with `laterAuditIds` (empty when the field changed outside the audit log); `cascade` reverts the later rows on the same field newest first and marks their suggestions `reverted`, without cascading into those suggestions' other fields. Revert reasons are the reject reasons minus `target_gone`. A reverted synopsis takes the licence of its restored source from `data_sources`; its source URL is not restored. A translated title that did not exist before the change is not revertible (the translation row is not deleted).

### 7.6 Create, merge, public views

`create` inserts the record with audit rows `before=null` (`source=create`) and unblocks `depends_on` suggestions. `merge` calls the SDD 01 merge service (legacy parity: same type, row locks, conflicting ratings skipped), writes one row `field='*merge*'` with both ids and moved counts, expires open suggestions on the source, and is not revertible (typed-title confirmation). Implementation: a create needs `external_ids.wikidata` (the item identity), refuses a Wikidata id already in the catalogue with 409 `suggestion-duplicate`, stores the new id in `meta.created_id` (`target_id` stays null on a create), and then applies its approved dependents, each in its own transaction. Item creates only in v1; entity and edition creates answer 409 until the create form carries a kind or parent work. The merge approve form asks for `confirm_title`, the deleted item's title (case-insensitive), and the `*merge*` audit row sits on the surviving item with `before` = both `catalog_merge_audit` snapshots. Recent changes shows field label, old → new and nickname ("HexaScore team" for admins, "deleted user"). Tier and accuracy are visible only to the user and admins.

## 8. Edge cases

- Concurrent 3rd and 4th submit: the row lock admits one; the other gets 429.
- Submit at 23:59:59.900 UTC counts for that day.
- Admin edit while a suggestion is open: approve needs `confirm_stale`; audit stores the real `before`.
- Target deleted or merged: open suggestions `expired` (`target_gone`), no stats impact.
- Author banned or deleted: open suggestions withdrawn by system; audit stays.
- Lock added on open suggestions: approve 423 until unlocked.
- Expired draft: its uploaded cover is deleted.
- Mail down: the job retries; the in-app row is already committed.
- Apply deadlock: rolled back, retried once, then shown to the admin.
- Bulk approve of two competing suggestions: second fails `stale-before`.

## 9. Acceptance criteria

- [ ] `vukem migrate` then `vukem migrate:status` lists the five tables' migrations as Ran (`user_notifications` belongs to SDD 14).
- [ ] `go test -tags=integration ./app/community/... -run TestQuota` → PASS: 3 submits 303, 4th 429 with `type=…/suggestion-quota` and `Retry-After` > 0; after the fake clock reaches 00:00 UTC, 303.
- [ ] `-run TestWithdrawStillCounts` → submit, withdraw, 2 submits, 4th 429.
- [ ] `-run TestTransitions` → every pair outside 7.1 returns 409.
- [ ] `psql "$DATABASE_URL" -c "update catalog_audit set after='null' where id=1"` → `ERROR: catalog_audit is append-only`.
- [ ] `-run TestApplyAtomic` → injected failure leaves item, suggestion, audit unchanged and no job.
- [ ] `-run 'TestRevert|TestEligibility|TestDuplicate|TestFieldLock'` → PASS (409/403/409/423 branches).
- [ ] `go test ./... -run TestSecurity_Community` → PASS.
- [ ] Playwright `community.spec.ts`: suggest → submit → approve → Recent changes → revert; axe clean in en, pt-BR, es.
- [ ] k6 `moderation.js` with 10,000 open suggestions: queue p95 < 500 ms, submit p95 < 300 ms.

## 10. Test strategy

- Unit: field validators, state table, tiers, accuracy, `Retry-After`, diff escaping.
- Integration (Postgres 17): 20 concurrent submits by one user, duplicate lock, apply/revert, append-only trigger, stats rebuild, deletion nulling, `DispatchTx` rollback.
- Security: `TestSecurity_Community_A01_PolicyFailsClosed`, `_A01_SourceURLSSRF`, `_A01_CSRF`, `_A03_DiffEscapesHTML`, `_A04_HoneypotLogged`; ZAP.
- E2E: Playwright + axe, including a two-admin claim conflict.
- Nightly asserts: every `applied` has audit rows; no open suggestion targets a missing record; stats equal rebuild.

## 11. Open issues

1. Contributor licence. Decided (04-1): yes, `source=community` values leave through the paid API under A-07 (perpetual non-exclusive contributor licence in the terms of use; API terms on use and attribution).
2. Quota for trusted contributors. Decided: 3/day for everyone at launch (D-31); a trusted-moderator role gets its own quota by a new decision.
3. Connectors and `source=community` fields. Decided (04-3): skip for 180 days, `system` suggestion on divergence (7.3); adopted in SDD 06.
4. vukem admin `BeforeSave` does not document access to the stored record's previous values [NOT IN VUKEM DOCS]; `vukem-feedback.md` candidate: pass the loaded record or the audit `{before, after}` map, or add an in-transaction `AfterSave`. Owner: vukem.
5. The vukem admin panel documents no custom record actions, bulk actions, assignment or diff views [NOT IN VUKEM DOCS]; the queue is an app page under `/admin/moderation` using the same gate. `vukem-feedback.md` candidate: `Resource.Action`. Owner: vukem.
6. vukem has no persistent in-app notifications [NOT IN VUKEM DOCS] (`https://vukem.dev/docs/digging-deeper/sse` SSE does not replay); VF-011. Decided: SDD 14 owns `user_notifications` (14-1); SDD 09 renders the inbox. Owner: vukem.
7. Calendar-aligned windows are not documented for vukem's fixed-window limiter, so the daily quota uses a DB count. Owner: vukem. Closes if such a limiter ships.
8. Copied synopses are checked by moderators only; automated similarity checks deferred. Owner: architect. Closes after 30 days of production data.
9. Merge is irreversible because legacy merge drops conflicting ratings; a reversible merge (tombstone, redirect, kept rows) belongs to SDD 01. Owner: architect.
