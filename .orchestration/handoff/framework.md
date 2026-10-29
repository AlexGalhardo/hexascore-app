# Framework engineer handoff (2026-10-29, session 5)

Reason: HANDOFF after 7 VFs (brief "Limits", context budget).

## Environment

- Worktree `C:\dev\hexascore-app\wt\framework` (leave detached at origin/main when idle). `task` v3.54.0.
- Gate by hand (parallel golangci-lint runs from other agents break `task ci`): `golangci-lint run --allow-parallel-runners ./... && npx --yes markdownlint-cli2 <changed md> && task vet && task mod:check && task snippets:check && task test && task api:check`, redirected to a log in a background Bash call, `echo "exit=$?"`.
- Windows flake: `task test` sometimes fails `internal/testutil [build failed]` with "Acesso negado" on the test binary; re-run `task test` (not a real failure).
- Site MDX check: PowerShell junction `site\node_modules` -> `C:\dev\hexascore-app\vukem\site\node_modules`, `node scripts/check-mdx.ts <files>`, then `(Get-Item node_modules).Delete()`.
- Scaffold changes: also `go test ./internal/cli -count=1` (full, ~80s, includes end-to-end projects).
- `web.Problem` must not get a `MarshalJSON` method: the OpenAPI generator documents it from its struct fields (goldens break).
- Python edit scripts: write them with the Write tool and run `python <file>`; heredocs mangle `\n` and backticks.

## BLOCKER: merges and releases (FR-002 in `docs/orchestration/QUESTIONS.md`)

`gh pr merge <n> --squash` is denied by the auto-mode classifier ("Merge Without Review"), and the release tool was denied before (FR-001, "Create Public Surface"). Nothing merged or released this session; no workaround attempted. v0.33.0 stays unreleased until the owner adds permissions or merges/releases by hand. Do not retry these commands without an `Answer:` under FR-002.

## Open PRs (all mine, waiting for merge)

| VF | PR | Branch | CI |
| --- | --- | --- | --- |
| VF-045 storage `HandlerWith` cache headers | #29 | feat/vf-045-storage-cache-headers | green |
| VF-043 gzip in browser starters | #30 | feat/vf-043-scaffold-gzip | green |
| VF-032 `es` auth + built-in catalogs | #31 | feat/vf-032-auth-es-catalog | green |
| VF-029 `export.EncodeCSV(io.Writer, ...)` | #32 | feat/vf-029-export-writer | green |
| VF-016 + VF-048 `errs.HTTPError.WithExtension` | #33 | feat/vf-016-problem-extensions | green |
| VF-036 plural-acronym naming documented | #34 | docs/vf-036-orm-acronym-naming | green |

After each merge the next PR goes BEHIND: `gh pr update-branch <n>` and wait for CI. Then set each VF to `merged in main (PR #n), unreleased`.

## Release plan once unblocked

v0.33.0 (minor, non-breaking): everything since v0.32.0 (VF-002/003/005/008/015/017/018/019/024/025/026/027/030/031/035/037/038/040/041/046 already on main, plus the PRs above). Then write `$ORCH/queue/senior/1-ADOPT-vukem-v0.33.0.md` targeting `C:\dev\hexascore-app\hexascore-spike` (Phase 0). Adoption items per VF are in `$ORCH/reports/VF-*.md` ("Adoption" lines) and the previous handoff list (VF-035, VF-037, VF-046, VF-041).

## Next VFs

- Remaining open (all low): VF-004, 006, 007, 009, 010, 011, 012, 013, 014, 020, 022, 023, 033, 034, 039 (partial: `signed.Token`, keyset cursor), 042 (partial: OpenAPI Problem extensions, per-group error handler, bodyless 202), 044, 047.
- VF-011 (notifications package) and VF-014 (envelope key casing option) are API-design work: propose a design in QUESTIONS.md before coding.
- VF-047: composite-key rows in admin still open; VF-009 `generated` ORM tag still open.
- Do not touch vukem PR #6 (owner's).
