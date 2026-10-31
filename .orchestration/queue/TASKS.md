# Task board

Protocol: `PROMPT-WORKERS.md` section 3. Claim by writing your name in `owner` and `doing` in `status`, wait 3 s, re-read. The `files` column is a lock. Rows sorted by `prio`. Seeded 2026-10-31 by the architecture audit.

## Open

| prio | id | role | status | owner | branch | files | note |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 0 | P2.H.3.2 | senior | todo | | | `tests/pow_hardening_test.go`, `app/http/middleware/challenge.go`, `app/challenge/` | ROADMAP; CI on `main` is red because of this test |
| 1 | P2.H.1.1 | senior | todo | | | `app/catalog/listing_view.go`, `app/catalog/item_page.go`, `app/catalog/covers.go`, `resources/js/components/CoverThumb.tsx`, `resources/js/components/catalog/`, `resources/js/components/library/ItemLink.tsx`, `database/seeders/demo_catalog.go` | ROADMAP; first visual task, the others depend on it |
| 1 | P2.H.2.1 | senior | todo | | | `resources/js/components/Skeleton.tsx`, `resources/js/pages/Profile/`, `resources/js/pages/Library/`, `resources/js/pages/Lists/`, `resources/js/pages/Account/`, `resources/js/pages/Developer/`, `resources/js/pages/Billing/`, `app/controllers/profile/`, `app/controllers/library_controller.go`, `app/controllers/lists_pages.go`, `app/controllers/notifications_controller.go` | ROADMAP |
| 1 | P2.H.3.1 | senior | todo | | | `app/controllers/publicapi/items.go`, `openapi/`, `config/billing.go`, `app/api/quota_middleware.go`, `sdks/`, `resources/docs/rate-limits.md` | ROADMAP; burst number waits for Q-A2, enum part can start |
| 1 | REBASE-P4.A.2.4 | senior | todo | | | `app/community/`, `app/http/limits/`, `app/jobs/`, `database/migrations/`, `tests/community*` | cherry-pick `86a031e` (community covers and source URL guard) from `agent/implementer/P4.A.2.4` onto a fresh `agent/senior/P4.A.2.4` branch; never merge the old branch (pre-squash base); acceptance of ROADMAP P4.A.2.4; then REV |
| 1 | ADOPT-VF-001 | senior | todo | | | `app/accounts/`, `app/controllers/account/`, `app/controllers/profile/`, `app/mail/`, `app/requests/`, `database/legacyimport/`, `database/migrations/`, `resources/js/pages/Auth/`, `resources/js/pages/Profile/`, `resources/lang/`, `resources/views/mail/`, `routes/auth.go`, `tests/passkeys*` | owner said `adotar VF-001` (2026-10-31); vukem v0.34.0 ships the passkey API; cherry-pick `1921bbb`, `2efc7ae`, `b438941` from `agent/implementer/P1.C.7.1` onto `agent/senior/P1.C.7.1`, resolve against `main`, acceptance of ROADMAP P1.C.7.1, P1.C.7.2 and P2.F.2.2b; then REV |
| 1 | FIX-legacy-import-no-login | senior | todo | | | `app/console/`, `cmd/activity/`, `cmd/dbcounts/`, `database/legacyimport/` | cherry-pick `661114b`, `6f99082` from `agent/implementer/import-allow-no-login` (opt-in import of legacy users without login); decide whether `263a3a7` (`cmd/activity`, `cmd/dbcounts` wip tools) stays, moves to `scripts/`, or is dropped; then REV |
| 2 | FIX-legacy-cover-retry | senior | todo | | | `scripts/legacy/` | cherry-pick `656f426`, `0a35a00` from `agent/implementer/P2.F-data-migration` (retry of legacy cover uploads); then REV |
| 2 | FIX-export-purge-zip | senior | todo | | | `app/library/export*`, `tests/export*` | cherry-pick `4750ae7` from `agent/especialista/P1.F.4.1-followup-ui-purge-fix` (delete an export zip built after the account purge); then REV |
| 2 | FW-triage-branches | senior | todo | | | `$FW` branches only | triage vukem `agent/implementer/railway-observability` (1 commit, no PR) and `backup/specialist-feedback-a`: open a PR or delete with a one-line report |
| 1 | DOC-board-mirror | estagiario | todo | | | `docs/orchestration/BOARD.md` | refresh BOARD.md from this table every day; format in the file |
| 1 | QA-gate-2026-11-01 | estagiario | todo | | | `.orchestration/reports/QA-gate-2026-11-01.md` | run the four gate commands of ROADMAP "Resuming work" step 2 on `main`, plus `bun run e2e` against the QA container; report failures with the decisive lines |
| 2 | P2.H.1.2 | senior | todo | | | `resources/js/pages/Catalog/Show.tsx`, `resources/js/components/item/`, `resources/js/components/rating/` | ROADMAP; after P2.H.1.1 |
| 2 | P2.H.1.3 | senior | todo | | | `resources/js/pages/Home/Index.tsx`, `app/controllers/home.go` | ROADMAP; after P2.H.1.1 |
| 2 | P2.H.1.4 | senior | todo | | | `resources/js/pages/Library/Index.tsx`, `resources/js/pages/Lists/`, `resources/js/pages/User/`, `resources/js/components/library/ShelfCard.tsx` | ROADMAP; after P2.H.1.1 and P2.H.2.1 (same pages) |
| 2 | P2.H.1.5 | senior | todo | | | `resources/js/pages/Lists/Edit.tsx`, `resources/js/components/search/SuggestBox.tsx`, `app/controllers/lists_controller.go` | ROADMAP; after P2.H.1.4 |
| 2 | P2.H.2.2 | senior | todo | | | `resources/js/app.tsx`, `resources/js/components/Field.tsx`, `resources/js/pages/Auth/Login.tsx`, `resources/js/pages/Auth/Signup.tsx` | ROADMAP; `challenge.go` is locked by P2.H.3.2 until it merges |
| 2 | P2.H.2.3 | senior | todo | | | `app/controllers/ratings_controller.go`, `app/controllers/library_controller.go`, `routes/admin.go`, `resources/js/components/Flash.tsx` | ROADMAP |
| 2 | P2.H.3.3 | senior | todo | | | `app/admin/`, `resources/views/admin/` | ROADMAP |
| 2 | P2.H.3.4 | senior | todo | | | `packages/tokens/`, `resources/js/**/*.tsx` (class names only) | ROADMAP; schedule when no other UI row is `doing` |
| 2 | P2.H.3.5 | senior | todo | | | `app/http/middleware/render_cache.go`, `routes/web.go`, `qa/k6/page-ttfb.js` | ROADMAP |
| 2 | DOC-i18n-legal | estagiario | todo | | | `.orchestration/reports/DOC-i18n-legal.md` | Q-D-007: legal copy (`policy.*`, `terms.*`) still mentions passkeys (now back), follows, credits, PIX; propose the exact wording per locale in the report; no edit to the catalogues until the owner approves |
| 2 | VF-062-adopt | senior | todo | | | `app/controllers/publicapi/`, `app/appapi/`, `go.mod` | adopt the vukem `main` fix for VF-062 (typed POST status) once the owner releases vukem v0.35.0; until then keep the workaround |
| 3 | P4.A.3.1 | senior | todo | | | `resources/js/pages/Community/`, `app/controllers/community/`, `routes/community.go` | ROADMAP reserve; community submission UI (code-only part of P4.A.3) |
| 3 | P2.H.3.6 | senior | todo | | | `CHANGELOG.md`, `docs/PROGRESS.md`, `ROADMAP.md` | release `v0.16.0`; the owner pushes the tag; after every P2.H row |
| 3 | QA-mobile-proto-review | estagiario | todo | | | `.orchestration/reports/QA-mobile-proto-review.md` | read `apps/mobile` and SDD 10; list what the prototype has versus P5.A.1–P5.A.2 acceptance; no code |
| 3 | AUDIT-social-sdd | especialista | todo | | | `.orchestration/reports/AUDIT-social-sdd.md` | prepare the SDD 14 §8 outline for Phase 7 (follows, feed, reviews, comments, reactions, privacy matrix) so P7.A.1.1 starts from a reviewed draft; no decision taken, Q-A1 stays with the owner |

## Blocked (owner gate; see `developer-manual-todos.md` and `docs/orchestration/QUESTIONS.md`)

| id | why |
| --- | --- |
| CI-release-environment (`agent/implementer/manual-access`, commit `e5cfbfd`) | needs the GitHub `Production` environment with a reviewer (P2.B.2.2, owner) before `release.yml` can require it |
| P2.A.*, P2.B.2.*, P2.C.*, P2.D.5.2, P2.D.6.1, P2.E.2.3, P2.E.3.2, P2.E.4.1, P2.F.1.*, P2.F.5.*, P2.F.6.1, P2.G.* | Railway, DNS, Stripe live, real data, cutover, spend, secrets |
| P7.* | `v1.0.0` (D-06 amended 2026-10-31, Q-A1 answered) |

## Done

(moved here by the orchestrator when `merged`; trimmed weekly, history stays in `reports/`)
