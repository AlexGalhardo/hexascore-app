# Handoff: especialista (2026-10-31, session 14)

Reason: 11 reviews plus 2 rejections this session (limit about 12). STATUS: HANDOFF.

## Done this session (local `main`, no push, no tag)

| Row | Verdict | Merge |
| --- | --- | --- |
| REV-FIX-observability-railway-native | CORRIGIDO_E_APROVADO (perfsprint in foryou_test, SDD 12 names `alert=trending_counter_dropped`) | 3cc01ff |
| REV-P4.B.4.1 | APROVADO | b491a88 |
| REV-P4.B.4.2 | CORRIGIDO_E_APROVADO (metrics, Grafana files and metrics test stripped per D-35; job logs `alert=community_asserts`; ROADMAP ticked with a note: the A9 in-app check is new work) | 7eff904 |
| REV-FIX-community-create-kinds-ui | APROVADO | 6fe871f |
| REV-FIX-community-create-tab-label | APROVADO | 07807ac |
| REV-FIX-community-mail-optout-ui | CORRIGIDO_E_APROVADO (P4.B.3.1 migration restored; SQLite column in new `20261101140000_add_community_mail_opt_out_sqlite`) | ac21bc3 |
| REV-ADOPT-vukem-0.35.0 | CORRIGIDO_E_APROVADO (predeploy lock kept around drop + migrate, 13796ad; goldens unchanged; covers ETag test ran on RustFS) | 8a73a3c |
| REV-FIX-api-host-only-api | CORRIGIDO_E_APROVADO (golines) | cf25938 |
| REV-P5.B.1.1 / 1.2 / 2.1 | APROVADO / APROVADO / CORRIGIDO_E_APROVADO (mobile catalogues recopied; main's keys.test was red since the mail-optout merge) | 63ed8f4 / ddd03d7 / 44f826d |
| REV-P6.A.3.2, REV-P6.A.3.4 | REJEITADO (rows `rework`, lists in the reports) | none |

Integration tests (Docker back): `app/moderation`, `app/jobs`, `app/community` and the named `./tests` patterns for the first six rows passed on main 2199f1f. Later rows ran their own targeted integration tests.

## Open, in order

1. REV-P6.A.1.3, 1.4, 2.1, 2.2, 2.3, 3.1, 3.3, FIX-ingest-trim-titles: not merged. A fresh-context static review of the whole stack gave these to verify by hand before merging (they are in branches below 3.2):
   - 1.3 `run/continuation.go`: the Step uses the in-memory cursor read before the lock; `SELECT ... FOR UPDATE` reads only status; the UPDATE to `running` has no status guard (a redelivery can reprocess a chunk, move the cursor back, or revive a cancelled run). Probable small fix: read status + cursor under FOR UPDATE, exit unless queued/running, guard the UPDATE.
   - 2.1 `wikidata/select.go`: on resume, if the `after` QID is no longer in the re-run selection every remaining target is skipped and the run ends succeeded.
   - 1.4 `httpx/breaker.go`: any success (also an in-flight one) closes an open breaker without the probe; concurrent finish can allow a second probe. `client.go`: 60 s client timeout covers the whole retry loop while Retry-After pauses up to 300 s.
   - 2.2/2.3: infrastructure errors (circuit open, host not allowed, ctx cancel) become per-target failures and count toward parking.
   - 2.2 `wikidata/entities.go`: process-wide unbounded `refs` cache filled before commit.
   - 1.4: only one `TestSecurity_A01_IngestSSRF` (2 cases); no tests for redirect to a blocked host, http scheme, body caps.
   Worktree `wt/especialista-P6.A` (detached, stack tip 3.4 with main merged, built) may still exist; remove it if stale.
2. P6.A.3.5a (senior-fw, doing) is stacked on the rejected P6.A.3.4: tell the orchestrator.
3. REV-P5.B.2.2, REV-P5.B.2.3, then impl-mobile rows (catalog-ui, item-ui, library-ui, states-ui, detail-ui, settings-wire).
4. impl-web rows: shell, search-palette, catalog.
5. REV-FIX-seed-flag: do not merge separately (ships in hotfix v0.12.4; the orchestrator back-merges origin/main).
6. Full `scripts/ci/local.sh` on main was not run this session (Docker came back mid-session; run it alone or in halves).

## Low findings, no rows

- FIX-api-host-only-api put its changelog line under Added (it is a fix); hotfix v0.12.4 also carries it, expect a CHANGELOG/api_edge.go overlap on the back-merge.
- P4.B.4.2 merge resolution lives in merge commit a5e9095 (the follow-up empty commit was refused by the header length hook).
- Mobile `itemMe` query key reuses element 0 `"me"`; a prefix invalidation of `["me"]` also hits item ratings.

## Pitfalls

- Python on Windows rewrites LF to CRLF in text mode: open with `newline=''`. Helpers used: scratchpad `addcl.py` (changelog group append) and `row.py` (TASKS row status/owner/note).
- Some TASKS rows have no trailing pipe; split on `|` and handle it.
- After merging any change to `resources/lang/*.json` keys under bundled mobile namespaces (`account`, `catalog`, `item`, ...), run `bun run --cwd apps/mobile catalogues` or `keys.test.ts` fails on main.
- `.git/index.lock` can be held briefly by the orchestrator: retry, never delete it.
- golangci-lint hits the default timeout under host load: pass `--timeout 15m --allow-parallel-runners`.
