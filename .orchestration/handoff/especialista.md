# Handoff: especialista (2026-11-01, session 20)

Reason: review budget used (about 20 rows) and the gate result is final. STATUS: HANDOFF.

## Done this session (local `main`, no push, no tag)

| Rows | Verdict | Merge |
| --- | --- | --- |
| FIX-P7.B.1.3-followups + P7.C.1.1 (one stack; old `agent/senior/P7.B.1.3` not merged) | CORRIGIDO_E_APROVADO: route inventory (exports, per-type activity) was red on main; plural-safe `friendsScore.summary` | e32b2b64 |
| P6.B batch on `agent/senior/P6.B.3.3` @ 987a13f2: P6.B.2.1, 2.2, 2.3, FIX-editions-qid-conflict, FIX-covers-a05-parity, P6.B.4.1-4.5, P6.B.3.3, FIX-ingest-qids-param, P6.A.2.4, P6.B.1.2, 1.3, FIX-ingest-retry-after-timeout | CORRIGIDO_E_APROVADO: migration renamed `20261101191000`; admin ingest routes in the inventory; rollback keeps items with reviews, activity events or affiliate links; `tmdb.expire` not rollbackable; pressure samples trimmed on Record; Twitch secret in the form body. Ticked P6.A.2.4, P6.B.1.2, 1.3, 2.1, 2.2, 2.3, 3.3, 4.1-4.5 (status lines on 2.1, 4.3, 4.5 per D-35) | eb47e274 |
| CLEAN-dangling-vf-ids | CORRIGIDO_E_APROVADO: VF-063, VF-064 numbered; removing the VF-018 semaphore also fixed `TestAdminIngestRollback` (it deadlocked nested chain dispatch on the sync queue) | 4d23b04b |
| CONFIG-email-addresses | APROVADO (vukem-feedback merge) | 37b8770b |
| FIX-ci-local-need-build (mine) | `local.sh` exits 2 when openapi or lighthouse is selected without build; `scripts/ci/local.test.ts` | 9fbb12cf |
| Gate fixes (mine) | vitest excludes `local.test.ts`; rollback test uses `search.IsDirty`; `questions-to-dev.md` (repo copy) blank lines; mobile catalogues recopied | 66cbe79e, b33727f1 |
| SEC-anti-scraping | CORRIGIDO_E_APROVADO: the signed-in page budget needs a logged-in session (`auth_web` peek), not any cookie; VF-065 | 014825f2 |

TASKS.md rows set to `ready` on disk, not committed (the file carries other agents' edits). New prio 3 rows: FIX-ingest-rollback-api-changes, FIX-ingest-alerts-continue, FIX-ingest-qids-pushdown, FIX-covers-orphans-on-rollback, QA-dq-serialise-probe, QA-admin-ingest-e2e, P6-edition-covers, FIX-covers-jpeg-fallback, QA-e2e-stale-item-docs.

## Gate on main (final, jobs run alone or in small groups)

PASS: test-go, test-ts (66cbe79e), test-go-integration (364 s), migrations, vuln (789319b6), deps, setup-windows, build, mobile (b33727f1). lint PASS on every step except commitlint over `v0.12.3..HEAD` (eight older commit messages; history is not rewritten, the release PR carries one squashed message, logged in `questions-to-dev.md`).

Host-only, not code: openapi goldens, Redocly and app-client codegen PASS, Schemathesis gets 20 "Network Error / Connection refused" even alone (Docker Desktop cannot reach the container with `--network host`); lighthouse fails with the Kaspersky signature (perf 0.62, best-practices 0.75, script 386 KB, CSS 231 KB). Both need CI or the Linux container recipe in reports/FIX-lighthouse-budgets-main.md.

SEC-anti-scraping merged after the gate: its targeted Go, integration and vitest suites passed on the branch with main merged in.

## Open, in order

1. REV-FIX-seed-flag: do not merge (hotfix path).
2. Mobile review rows still `todo`: REV-FIX-mobile-deeplink-locale, REV-FIX-mobile-item-add-to-list, REV-IMPL-mobile-v1-lists-ui, REV-IMPL-mobile-entity-paging, REV-FIX-mobile-search-trending-flash, REV-IMPL-mobile-search-trending.
3. Four VF-NEW entries keep provisional ids (EXPORT-WRITER, PROBLEM-TYPE, HOST-ONLY, REQUEST-LINE-FIELDS, MAIL-REPLYTO); next free number is VF-066.
4. Then the orchestrator prepares the v0.13.0 release PR.

## Pitfalls

- The main checkout's `.env.example` carries someone's uncommitted lines; a merge that touches it refuses. Save `git diff .env.example`, `git checkout -- .env.example`, merge, `git apply` the patch (no `git stash`).
- golangci `run` does not check formatting: also run `golangci-lint fmt --diff` on touched packages.
- `TestSecurity_A01_HexaRouteStackInventory` breaks on every new route; run it on each review.
- New resource strings: rerun `bun run --cwd apps/mobile catalogues` or the mobile `keys.test.ts` fails.
- Unregistered leftover folder `wt/especialista-gate` exists; do not reuse that path.
- Do not trap your own IP on a QA stack (`/data/export` bans it for 24 h).
