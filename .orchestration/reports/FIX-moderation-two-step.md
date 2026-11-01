# FIX-moderation-two-step moderators review, admins apply (owner Q-A15)

status: concluida
branch: agent/senior/FIX-moderation-two-step @ 05945d4 (stacked on agent/senior/FIX-community-api-source @ 931210f → P4.B.2.3)
files: `docs/sdd/04-community-edits.md` (FR-04-17, §5 `suggestion_reviews`, §6 routes and abilities, §7.1), `app/admin/abilities.go`, `app/admin/moderation_handlers.go`, `app/admin/catalog_handlers.go` (`CanApply`), `app/admin/views/moderation{,_detail}.html`, `app/moderation/review.go` (new), `app/moderation/detail.go`, `app/moderation/queue.go`, `database/migrations/20261031190000_create_suggestion_reviews.go`, `routes/admin.go`, `tests/testdata/route_inventory.txt`, `tests/admin_moderation_gate_test.go`, `tests/moderation_review_test.go`
design: new admin-only ability `applySuggestion` gates approve, reject and bulk approve/reject (merge suggestions apply through approve; `/admin/items/{id}/merge` was already behind the panel admin gate). `moderate` keeps queue, detail, claim, release and the new `POST /admin/moderation/{id}/recommend` (`version`, `recommendation` = approve|reject, `note` ≤ 1000). `moderation.Queue.Approve/Reject` re-check `actor.Role == admin` (`ErrAdminOnly` → 403) so no other caller skips the rule. A recommendation is one row per (suggestion, reviewer), upserted, storing reviewer, time and the version seen; it is not a transition and does not bump `version`. Queue shows "approve N · reject N"; detail lists recommendations; approve/reject controls render only for admins.
acceptance:

- `go test -tags integration ./tests -run 'TestSecurity_A01_ModeratorCannotApply|TestModerationRecommend|TestModerationGate|TestSecurity_Community_A01|Inventory' -v` → all PASS (incl. `TestSecurity_A01_HexaRouteStackInventory`)
- `go test -tags integration ./tests -run 'Moderation|Security_Community|Migrat'` → ok; `./app/moderation/ ./app/community/` → ok
- `bun test scripts/ci/migration-policy.test.ts` → 3 pass

validation: go build ok, `golangci-lint run --build-tags integration` on touched packages 0 issues; browser check not done (pages are server-rendered templates asserted by the integration tests)
changelog: Changed: moderation is two-step: moderators record approve/reject recommendations, only admins apply or reject suggestions (Q-A15)
roadmap: no box (amends P4.B.1.x behaviour; P4.B.1.1 ability list now includes `applySuggestion`)
risks / follow-ups:

- 2FA: the moderator 2FA rule (FIX-moderator-2fa, merged on `main` as 3e0cebb) is not in this stacked chain's base; it arrives when the chain merges into `main`. No change to `admin_guard.go` here.
- `suggestion_reviews.reviewer_id` is `ON DELETE CASCADE` (part of the primary key), so deleting a moderator's account drops their recommendations; audit and events are untouched (NFR-04-5 is about public surfaces).
- Recommendation is allowed on `submitted` and `in_review` without a claim; a stale `version` answers 409 `stale-before`, a decided suggestion 409 `invalid-transition`.
- ROADMAP P4.B rows that say "moderate" for approve should read `applySuggestion` (orchestrator, ROADMAP is not mine to edit).

## Review (especialista, 2026-10-31)
verdict: CORRIGIDO_E_APROVADO
re-run (main + branch): `go test -tags integration ./tests -run 'TestSecurity_A01|TestModerationRecommend|TestModerationGate|TestSecurity_Community|Inventory|Moderation|Migrat|Moderator'` → ok; `./app/moderation/ ./app/community/` → ok; `bun test scripts/ci/migration-policy.test.ts` → 3 pass; golangci (integration tag) moderation, admin, database, routes, tests → 0 issues; markdownlint SDD 04 → 0
checked (orchestrator list): every catalogue-writing path is admin-only — approve (incl. merge suggestions) and reject: route gate `applySuggestion` (adminOnly) plus `requireAdmin` in moderation.Queue; bulk approve/reject: same gate before Queue.Bulk, and Queue.Approve/Reject re-check per row; item merge and affiliate links: panel `AbilityAdmin` (`admin.Gate`, admin role only, 404); revert and locks: no routes yet (abilities adminOnly; P4.B.2.4 must gate its route on `revertCatalog`). All covered by `TestSecurity_A01_ModeratorCannotApply` (approve, reject, bulk approve/reject, merge and affiliate routes, catalogue unchanged, controls hidden) and `TestSecurity_Community_A01_ApproveNeedsModerate`. Recommend: moderate ability, version check, no transition, upsert per reviewer; migration 20261031190000 (no suggestion delete path, so the FK without cascade is safe)
fixes: 1. A01 test now asserts a trusted moderator gets 404 on `/admin/items/{id}/merge` and both affiliate-link routes (fe17eeb)
gate on main after merge 2e89655: go build ok; tests TestSecurity_A01|Moderation|Moderator|Inventory|Migrat ok
PRONTO_PARA_MERGE agent/especialista/FIX-moderation-two-step-fix
