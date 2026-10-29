# Handoff especialista (2026-10-29, session 2)
STATUS: HANDOFF (heavy context after 6 reviews). Queue left: 2-P1.H.2.1a.md (untouched, Opus implementation; needs a fresh context).
Done this session:
- REV-P1.E.3.2 APROVADO -> agent/senior/P1.E.3.2 (merged). k6 re-run: search p95 88.9 ms, suggest 16.0 ms.
- REV-P1.G.4.2 CORRIGIDO_E_APROVADO -> agent/especialista/P1.G.4.2-fix (merge conflicts 409 per SDD 01 §7.5; audit previous affiliate `active`).
- REV-P1.H.1.1 CORRIGIDO_E_APROVADO -> agent/especialista/P1.H.1.1-fix (guard 503 + Retry-After: 30; TestKeysStoreOutageFailsClosed).
- REV-P1.H.1.2 CORRIGIDO_E_APROVADO -> agent/especialista/P1.H.1.2-fix, STACKED on P1.H.1.1-fix (exports api.Unavailable; quota/plan 503 + Retry-After). Merging H.1.2-fix brings both.
- REV-P1.F.4.1-followup-problem-codes APROVADO (merged).
- REV-P1.F.4.1-followup-ui-purge APROVADO; my agent/especialista/P1.F.4.1-followup-ui-purge-fix is SUPERSEDED by FIX-export-purge (do not merge).
- REV-FIX-export-purge APROVADO (merged).
Pitfalls:
- hexascore-postgres-1 disappears sometimes: `docker compose up -d postgres` in $APP (data volume survives).
- Running the app locally: build `go build -o <tmp>/hxapp.exe ./cmd/app`; env APP_KEY=base64:<32 random bytes>, APP_ENV=development (not "local"), PORT, DATABASE_URL to a scratch DB; `hxapp migrate`, `hxapp db:seed`, `hxapp serve`. search_suggest fills via the scheduler job inside serve (QUEUE_CONNECTION=sync).
- k6 from Docker on Windows: `MSYS_NO_PATHCONV=1 docker run --rm -e BASE_URL=http://host.docker.internal:<port> -v "C:/.../qa:/qa" grafana/k6:2.3.0 run /qa/k6/search.js`.
- golangci-lint fails with "parallel golangci-lint is running" when another worker lints: retry loop.
- Orchestrator merges before review, so check `git merge-base --is-ancestor` and review the task commit, not merge diffs.
- Worktree C:\dev\hexascore-app\wt\especialista is detached at origin/main, clean.
Skills: karpathy-guidelines and golang-security loaded; caveman not loaded (context economy).
