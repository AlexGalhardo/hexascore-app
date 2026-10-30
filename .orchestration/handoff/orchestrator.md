# orchestrator handoff (2026-10-30, session paused by the owner)

`.orchestration/STOP` exists: workers finish their current step and stop. Delete it when resuming.

## How this session ran

- Owner rules for the session (also in memory and briefs): workers stay alive in standby and get tasks by message (`briefs/common-app.md` "Standby rule"); relaunch a worker only when its session hits the context limit (it ends with `STATUS: IDLE "context limit"` after updating its handoff). Parallel workers and pushes to `main` were approved for this goal; ask again in a new session (global CLAUDE.md rule).
- Owner gates still apply: Railway plan/buckets/DNS/secrets, Stripe live values (Q-H31), Google Play, cutover. Recorded decisions are in `docs/orchestration/QUESTIONS.md` (FR-003 answer, FD-011/FD-014/FD-044/FQ-004 answers, owner requests for VF-001, VF-052, VF-055, VF-056, VF-057, VF-059, processed_events decision, drivers/s3 decision).
- The orchestrator (main session) merges vukem PRs with `gh pr merge <n> --squash` (owner authorized; workers must not). Releases of vukem only on owner request.
- Railway: CLI linked in `hexascore-vukem` to project `hexascore` / `production` / `app`. Reading variables works; `Bash(railway variables:*)` is now in the owner's allow list. A harmless test variable `ORCHESTRATOR_ACCESS_CHECK=2026-10-30` exists on `app` (remove it when convenient). Store variables are correct: `RATE_LIMIT_STORE=cache`, `CACHE_STORE=database`, `SESSION_DRIVER=database`.

## State at pause

- App: `main` released as `v0.8.0` (P1.H.5.1, GitHub release published, local container smoke test green). Merged after the tag: nothing but orchestration files.
- Gate commands (all green at the last merge): `go build ./...`; `golangci-lint run --allow-parallel-runners --build-tags integration ./...`; `TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable go test -tags integration ./...`; `bun run lint && bun run typecheck && bun run test`; `bun run openapi:lint`; markdownlint with an explicit file list. Tests booting with `APP_ENV=production` need `RATE_LIMIT_STORE=cache` and `CACHE_STORE=database`.
- Merge helper: `/tmp/union.py` (union-resolves `.md` conflicts); code conflicts are resolved by hand. VF-NEW ids are renumbered at merge (next free: VF-062).
- Not merged: `agent/senior/FIX-e2e-signup-load` (approved, but conflicts with P2.D.2.3 in `app/http/middleware/challenge.go`); requeued as `queue/senior/0-FIX-e2e-signup-load-r2.md`.

## Workers at pause

- senior: active `4-P1.I.5.1` (Lighthouse); delivered P1.I.3.1 (review queued). Queue: 0-FIX-e2e-signup-load-r2, 5-P1.I.4.1, 6-P2.B.1.4, 6-P2.B.1.5, 7-P2.C.1.2, 7-P2.C.2.2, 8-P2.D.5.3.
- especialista: active `4-P2.F.4.1` (cutover-verify + finalise). Queue: 1-REV-P1.I.3.1, 5-P2.G.1.1, 6-P2.E.1.2, 7-P2.C.3.3; blocked 9-APP-PASSKEYS-b (needs a vukem release with VF-001), 9-P2.F.2.2b (needs P1.C.7.1 passkeys table).
- framework: idle. Active (PRs open): DOCS-SITE-REFRESH (item 2, the full `last-updated.json` refresh, waits until #88-#91 merge), FW-CI-WINDOWS-FLAKES (#80, #81), VF-052 (#66/#67 to recreate on main now that #65 merged), VF-044 (#82), VF-057 (#87), VF-059 (#63).
- estagiario: idle since P2.A.1.1.

## vukem PRs open (merge in this order, `gh pr update-branch` when BEHIND, rerun a failed job once before treating it as real)

#63, #71, #77, #80, #81, #82, #87, #88, then #89, #90, #91 (docs; after them ask the framework for the full last-updated refresh), dependabot #83-#86, then the recreated #66/#67 replacements. Merged this session: #36-#43, #45-#47, #51-#59, #61, #64, #65, #68-#70, #72-#76, #78, #79. Issues #19 and #5 closed.

## Next

1. Resume: delete STOP, `git pull`, `docker compose up -d --wait`, run the gate, relaunch the four workers from their handoffs.
2. Merge the vukem PRs above; ask the owner whether to release vukem (v0.34.0: passkeys, React Email, jsonrpc, HTTP/2+3, Cloudflare, notify, Betterleaks, sitemap, httpclient, impersonation, OpenAPI options); after the release queue `ADOPT-vukem-v0.34.0` (incl. `notify`, `PageOf`, `HashRecoveryCode`, passkeys P1.C.7.x, Cloudflare `TRUSTED_PROXIES` notes, regenerate OpenAPI goldens if #52 changes them).
3. Finish P1.I (I.3.1 review, I.5.1, I.4.1, I.6.1 rows, I.7.1 release v0.9.0), then the code-only P2 items; owner gates stay listed in QUESTIONS.md.
