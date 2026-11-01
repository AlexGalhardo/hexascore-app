# Handoff: especialista (2026-11-01, session 21)

Reason: assigned work done. STATUS: HANDOFF.

## Done this session (local `main`, no push, no tag)

| Rows | Verdict | Merge |
| --- | --- | --- |
| REV-FIX-mobile-deeplink-locale, REV-FIX-mobile-item-add-to-list, REV-IMPL-mobile-v1-lists-ui, REV-IMPL-mobile-entity-paging, REV-IMPL-mobile-search-trending, REV-FIX-mobile-search-trending-flash | APROVADO, all six in batch `agent/especialista/mobile-batch-2` (from main 35cd2062): biome ci ok, `tsc` ok, jest `--runInBand --forceExit` 56 suites / 373 tests | 736875d4 |
| Review reports, CHANGELOG lines, P5.B.5.1 status line (session locale done; box stays open on the owner gate) | | 78c611d7 |
| VF numbering: VF-066 EXPORT-WRITER, VF-067 PROBLEM-TYPE, VF-068 HOST-ONLY, VF-069 REQUEST-LINE-FIELDS, VF-070 MAIL-REPLYTO; VF-063, VF-064 and VF-066..070 set to `fixed in v0.36.0 (pending release)` (vukem PR #128, 9cafa133) | | d8ae1372 |

TASKS.md: the six REV rows are set to `ready` on disk but not committed, because the file also carries other agents' edits. No ROADMAP box was ticked: these rows are not microtasks, and P5.B.4.1 and P5.B.5.1 stay open on owner device gates.

## Open

1. REV-FIX-seed-flag: do not merge it (hotfix path).
2. VF-065 (client IP and crawler check) is not covered on vukem `origin/main` 343667b1 and stays open.
3. App code comments still name the provisional ids (`VF-NEW-PROBLEM-TYPE`, `-HOST-ONLY`, `-REQUEST-LINE-FIELDS`, `-EXPORT-WRITER`) in `app/http/middleware/api_edge.go`, `request_log.go`, `app/controllers/exports_controller.go`, `app/jobs/export_ids.go`, `app/recommend/eval/eval.go` and `bootstrap/app.go`. The feedback headings keep those ids in the form "(provisional id ...)", so a grep finds them. ADOPT-vukem-0.36.0 removes those workarounds and the entries.
4. Do not open a release PR. The orchestrator rebuilds the release history after ADOPT-vukem-0.36.0.
5. The prio 3 rows the especialista created are for implementers.

## Pitfalls

- The main checkout's `.env.example` and `tsconfig.json` carry someone's uncommitted lines. A merge that touches either refuses. Save the diff, `git checkout --` the file, merge, then `git apply` the diff (no `git stash`).
- golangci `run` does not check formatting: also run `golangci-lint fmt --diff` on touched packages; use `--timeout 15m -j 2` (host is short on memory).
- `TestSecurity_A01_HexaRouteStackInventory` breaks on every new route; run it on each review.
- New resource strings: rerun `bun run --cwd apps/mobile catalogues` or the mobile `keys.test.ts` fails.
- A batch worktree needs `bun install --frozen-lockfile` before mobile jest; remove it with `git worktree remove --force` (untracked `node_modules`).
- Root markdownlint with explicit file arguments still globs `qa/e2e/test-results`; filter the output for your files.
- Unregistered leftover folder `wt/especialista-gate` exists; do not reuse that path.
- Do not trap your own IP on a QA stack (`/data/export` bans it for 24 h).
