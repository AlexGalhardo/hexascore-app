# FIX-docs-next-cursor
status: review
branch: agent/senior/FIX-docs-next-cursor @ 94a4787
red/green: test changed first and failed (docs: new test in scripts/build-docs.test.ts failed against old text); then fix; targeted tests green
validation: problems: `go test -tags integration ./tests/ -run TestProblem` ok, golangci-lint 0 issues. docs: `bun scripts/build-docs.ts`, `bun test scripts/build-docs.test.ts` 4 pass (markdownlint not installed in the worktree, not run)
changelog: Fixed: /problems/{code} answers 301; docs and SDK READMEs use meta.next_cursor; SDK READMEs list listChanges and webhook operations
notes: docs row also touches sdks/ts/README.md, sdks/python/README.md, scripts/build-docs.test.ts, docs.json

## Review (especialista, 2026-11-01)

verdict: APROVADO
re-run: batch tip 6eacf461: `bun test scripts/build-docs.test.ts scripts/ci/sdk-job.test.ts scripts/ci/sdk-release.test.ts` 16 pass; sdks/ts `bun install --frozen-lockfile`, `bun run typecheck`, `bun test` 5 pass; `bun run sdk:generate` no diff; `go test -run TestProblemCatalog ./tests/` ok; markdownlint ok. Not run: Python SDK pytest/uv build (no uv on this host), workflow dispatch (owner)
checked: SDD 05 6.5 (trusted publishing, no stored tokens, publish defaults off), actions pinned by sha, secrets none, i18n parity (3 locales), scope, tests; changelog line shared with FIX-problems-redirect-301
fixes: none
gate on main after merge: targeted only (section 8.2); full gate once at the end of the batch
PRONTO_PARA_MERGE agent/especialista/P3.B-batch
