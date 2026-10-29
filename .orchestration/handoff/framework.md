# Framework engineer handoff (2026-10-29)

Reason: HANDOFF after 7 VFs plus the CI fix (context budget, brief "Limits").

## Environment

- Worktree `C:\dev\hexascore-app\wt\framework` (detached at origin/main when idle). `task` v3.54.0 installed with `go install github.com/go-task/task/v3/cmd/task@v3.54.0` (vukem pins none; docs say `@latest`, latest stable used).
- Site MDX check without installing: junction `site\node_modules` -> `C:\dev\hexascore-app\vukem\site\node_modules`, run `node scripts/check-mdx.ts <files>`, delete the junction after (PowerShell `(Get-Item ...).Delete()`).
- Ruleset: PR + required checks + strict up-to-date + linear history. After each merge the other PRs go BEHIND: `gh pr update-branch <n>`, wait for CI, squash-merge. A serial merge loop (state BEHIND -> update, CLEAN -> merge, failing check -> stop) works well in the background.
- Python edits on Windows: open files with `newline=''` (text mode writes CRLF); heredoc backslashes may be collapsed by the tool, prefer the Edit tool for strings with `\n`.

## Done

| Item | PR | State |
| --- | --- | --- |
| FW-CI-MONGO (mongo:8 -> 9.0.4 digest + health check; kernel >= 6.19 incompatibility) | #8 | merged |
| VF-002 strict SSR (`SSRConfig.Strict`, `inertia.RequireSSR`) | #7 | merged |
| VF-027 `serve -with-ssr`, `build -ssr`/`-ldflags`, `WithVersion` | #9 | merged |
| VF-024 i18n path prefixes (`PathPrefixes`, `SourcePath`, `StripPathPrefix`, `LocalizePath`) | #10 | merged |
| VF-005 `CursorPaginate` over `OrderBy` (composite keyset) | #11 | merged |
| VF-015 `Router.Host` | #12 | merged |
| VF-017 `Migration.NoTransaction` + VF-026 migration lock | #13 | merged (b5a2253) |
| VF-001 | - | wontfix (post-1.0, FQ-001) |

Reports: `$ORCH/reports/{FW-CI-MONGO,VF-002,VF-005,VF-015,VF-017,VF-024,VF-026,VF-027}.md`. vukem `main` head: b5a2253.

## Open

- Docker compose services postgres/mysql/mariadb (project `vukem`) were started from the worktree for local integration runs; stop with `docker compose -f .config/compose.yaml stop` if unwanted.
- vukem-feedback.md: wrap links as `<https://...>` (app markdownlint MD034).
- Release v0.33.0 BLOCKED (owner question Q-D-006, do not retry the tool): the permission classifier denied `go run ./internal/tools/release v0.33.0` ("Create Public Surface"). Request FR-001 in `docs/orchestration/QUESTIONS.md`. Nothing committed or tagged. After the release: adoption task `$ORCH/queue/senior/1-ADOPT-vukem-v0.33.0.md` targeting `C:\dev\hexascore-app\hexascore-spike`, VF statuses `fixed in v0.33.0`.
- Next VFs by brief order (medium first, roadmap refs): VF-018 (queue priority/concurrency), then VF-019, VF-025, VF-008, VF-003, VF-030, VF-031, VF-035, VF-037, VF-038, VF-040; lows after.
- Do not touch vukem PR #6 (owner's).
