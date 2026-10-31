# Handoff: especialista (2026-10-31, session 2)

Reason: about 7 reviews done this session, so the context is heavy. STATUS: HANDOFF.

## Done this session

| Row | Verdict | Merge on local `main` |
| --- | --- | --- |
| handoff item 1 (gate `app/library` + `tests/...` on c9f03cc) | 5 packages ok | `reports/P2.H.1.1.md` line fixed |
| REV-P2.H.3.1 | APROVADO | 1323892 (ROADMAP box left open: Python SDK, see FIX-sdk-python-regen) |
| REV-P2.H.3.3 | APROVADO | c7f0abe |
| REV-FIX-gate-timeouts | APROVADO | 8f9e98c (ROADMAP gate line now `-parallel 4 -timeout 30m`) |
| REV-P2.H.2.3 | CORRIGIDO_E_APROVADO (lint fix b80929c) | 01e8dcf |
| REV-FIX-account-purge-lists | APROVADO | 7317f44 |
| (docs) ROADMAP ticks P2.H.2.3, P2.H.3.3 + CHANGELOG lines for the five above | n/a | eda3a96 |
| REV-FIX-legacy-import-no-login | CORRIGIDO_E_APROVADO (test e077e54 on `agent/especialista/FIX-legacy-import-no-login-fix`) | fe00949 |
| REV-DOC-link-check | APROVADO (report only) | none |
| REV-FIX-covers-base-state | REJEITADO r1 (JSON cover URLs turn relative when `COVERS_BASE_URL` is set; Q-A6) | none; source row `rework`, REV row `blocked` |

Gate on `main` @ 1323892: build, golangci-lint, bun lint/typecheck, go test integration all green. Gate @ eda3a96: all green (29 packages). Gate @ fe00949: build, lint and `database/legacyimport` green.

## Next rows (prio order)

1. REV-P2.H.1.3 (home): run the acceptance `curl -s http://127.0.0.1:3000/ | grep -c '<img'` ≥ 12 and Lighthouse on a QA container (`E2E_IMAGE_TAG=P2.H.1.3`); port 3000 was held by `hexascore-app:main` (probably QA-e2e-main by the intern). Never `--profile app down`.
2. REV-FIX-legacy-cover-retry, REV-FIX-mobile-catalogues, REV-FIX-i18n-untranslated (P2.H.2.3 is merged now: rebase-merge conflicts in `resources/lang/*.json` are expected, resolve by keeping both key sets), REV-P2.H.2.2.
3. Report-only rows: REV-QA-*, REV-DOC-* (sampled fact check + markdownlint).
4. P2.H.1.2 still needs `ratings.spec` and `item.spec` in all three locales (orchestrator note).

## Pitfalls

- `bunx markdownlint-cli2@0.24.0 <file>` also lints the repo globs in the config; filter its output by the file name. It prints old errors from `qa/e2e/test-results/`.
- `python`/`python3` hangs on this host (Store stub); edit with Edit, not a Python script.
- Worktrees left: `wt/especialista-*` for P2.H.3.1, P2.H.3.3, P2.H.2.3, FIX-account-purge-lists, FIX-legacy-import-no-login: remove them with `git worktree remove` after the orchestrator sets `merged`.
- Source branches that I merged keep their names; delete them after `merged`.
