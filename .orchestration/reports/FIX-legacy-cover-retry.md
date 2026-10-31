# FIX-legacy-cover-retry
status: concluida
branch: agent/senior/FIX-legacy-cover-retry @ 1f17eca (two commits: 50bdc27 retry transient copies and --concurrency, 1f17eca retry non-rewindable bodies)
files: scripts/legacy/blob-export.ts
acceptance:
- cherry-pick of 0a35a00 then 656f426 (chronological order) applied clean on main
- `bun run lint` → 0 errors (4 pre-existing warnings); `bun run typecheck` → tsc --noEmit exit 0
validation: no Go touched; not run against the real legacy blobs (owner gate)
changelog: Fixed: legacy cover copy retries DNS, rate-limit and non-rewindable-body errors and accepts --concurrency (HTTP 404 stays one failure)
roadmap: no box
risks / follow-ups: 50bdc27 has no Refs footer (original commit); behaviour only verifiable on a real run.

## Review (especialista, 2026-10-31)

verdict: CORRIGIDO_E_APROVADO
re-run: `bunx vitest run scripts/legacy` → 4 passed; biome clean; typecheck ok (gate)
checked: retry bound (5 attempts, linear 400 ms backoff), 404/4xx permanent except 429, `--concurrency` validated as a positive integer, manifest only appended after a successful write, scope (one script file)
fixes:
1. vitest case: one EAI_AGAIN blip is retried and copied, a 404 is fetched once and counted failed (new retry logic had no test)
follow-up: the transient regex matches on error text (`unexpected error`, `not seekable`); fine for a one-shot migration script
gate on main after merge: covered by the gate line of the session
PRONTO_PARA_MERGE agent/especialista/FIX-legacy-cover-retry-fix
