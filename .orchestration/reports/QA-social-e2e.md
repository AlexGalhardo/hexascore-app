# QA-social-e2e Playwright follows spec

status: entregue (spec written; e2e is not a gate, owner 2026-10-31)
branch: agent/senior/QA-social-e2e @ 7c150c25 (stacked on P7.A.1.3b @ 637de139)
files: `qa/e2e/specs/social.spec.ts` (new `follows` describe), `docker-compose.yml` (`FLAG_SOCIAL: ${FLAG_SOCIAL:-false}` on the app env)
acceptance:

- red: not run (Playwright needs the QA stack; running e2e is not a gate). The spec skips unless `E2E_SOCIAL_OPEN=true`, and the app needs `FLAG_SOCIAL=true`.
- `bunx playwright test -c qa/e2e/playwright.config.ts --list qa/e2e/specs/social.spec.ts` → 12 tests (2 per project, en/pt-BR/es desktop and mobile)
- Covered: follow updates the follower count, unfollow resets it, axe on the profile; the star turns private, so the fan's follow becomes "Requested"; the inbox shows the `follow_request` notice (axe) and Accept; the owner sees 1 follower and the fan sees Unfollow.
- Run: `FLAG_SOCIAL=true` on the QA stack (AGENTS.md parallel-stack recipe), then `E2E_SOCIAL_OPEN=true E2E_BASE_URL=... E2E_APP_CONTAINER=... bun run e2e -- qa/e2e/specs/social.spec.ts`

validation: `bun run typecheck` ok; `bunx biome ci qa/e2e` clean
changelog: none (tests only)
roadmap: closes the Playwright box of P7.A.1.2
risks / follow-ups:

- Never executed against a stack. The first run may need selector fixes, for example if the Accept button shares its name with another notification action.
- The second member uses a fixed `X-Forwarded-For` per worker (`10.250.<worker>.<repeat>`) to stay clear of the per-IP signup and login limits.

## Review (especialista, 2026-11-01)

verdict: CORRIGIDO_E_APROVADO (stack reviewed as one: P7.A.1.3b -> QA-social-e2e -> P7.B.1.1 -> P7.C.1.2 -> P7.B.1.2 -> P7.B.1.3)
re-run (`agent/especialista/P7-social-stack`, main merged in):
- `go test -tags integration ./tests -run 'TestFeed|TestProfileActivityTab|TestActivityRetention|TestPublicProfile|TestExportZip|TestDeletion|TestSecurity_A0|TestFollows|TestSocial|TestOnboarding|TestReviews|TestOGImage|TestComments|TestReactions|TestReport|TestItem|TestList|TestAdmin|TestProfile'` -> ok (62 s), `TestSecurity_A01_HexaRouteStackInventory` green after the merge with main
- `go test ./app/social ./app/seo ./app/http/...` ok; `go test -tags integration ./tests/migration` ok; `bun test scripts/ci/migration-policy.test.ts` 3 pass; `go test ./tests/i18n` ok
- `golangci-lint run --build-tags integration -j 2` on the touched packages -> 0 issues; `golangci-lint fmt --diff` clean; `bun run lint` clean; `bun run typecheck` exit 0; `bunx vitest run resources/js` -> 47 files, 138 passed; markdownlint SDD 14 clean
- `dwebp -version` in `node:26.13.0-alpine3.24` + `libwebp-tools=1.6.0-r0` (the runtime stage) -> 1.6.0 at /usr/bin/dwebp
checked: SDD 14 §12.5 privacy matrix, §12.9 purge/export, §12.10/§12.11; SDD 13 A01 (routes in the inventory), A05 (review Markdown is a node tree, no HTML), A06 (comments 60/h, reactions 300/h); SQL bound everywhere (table names only from the target type constant); flag off -> every new route 404 (`TestReviewsDisabledByDefault`, `TestFeedDisabledByDefault`); scope; tests
fixes:
1. migration version collision: main took `20261101170000` (entities name index); the reviews migration is now `20261101171000_create_reviews_table` (f23695ba). Note for P6.B.4.1: its `20261101190000` collides with `20261101190000_moderate_reviews_and_comments`, rename it at its merge.
2. merge of main: `item_controller.go` keeps `related` (similar/crossMedia) and `Social`; `routes/web.go` keeps the reco metrics wiring plus `items.Social` and the `/og` route; `Catalog/Show.tsx` keeps Reviews and MoreLikeThis; lang `social.feed` keeps `review_published` and `score`; `tests/feed_test.go` keeps both sides.
3. moderation notices: pt-BR/es showed the raw target type ("(review)"); the web inbox now translates `targetType` through new `notifications.target.{review,comment}` keys (15be12b3, 6ee03b5f). Mobile falls back to the generic notice for these kinds.
judged open points: `POST /reports` stays outside the flag (review/comment targets exist only if created while the flag was on, and moderation must keep working after the flag is turned off); no "unhide" outcome is a follow-up (owner/product), noted in ROADMAP; Playwright specs (social, reviews) and `k6 feed.js` not run (not gates, batch run).
roadmap: tick P7.B.1.2, P7.B.1.3, P7.C.1.2; P7.A.1.3 and P7.B.1.1 stay open with status lines (k6, Playwright)
PRONTO_PARA_MERGE agent/especialista/P7-social-stack (merged a9343e81)
