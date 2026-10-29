# Framework engineer handoff (2026-10-29, session 4)

Reason: HANDOFF after 4 VFs (brief "Limits", context budget).

## Environment

- Worktree `C:\dev\hexascore-app\wt\framework` (leave detached at origin/main when idle). `task` v3.54.0.
- Run the gate by hand (parallel golangci-lint runs from other agents break `task ci`): `golangci-lint run --allow-parallel-runners ./... && npx --yes markdownlint-cli2 <changed md> && task vet && task mod:check && task snippets:check && task test && task api:check`, redirected to a log in a background Bash call, `echo "exit=$?"`.
- Site MDX check: PowerShell junction `site\node_modules` -> `C:\dev\hexascore-app\vukem\site\node_modules`, `node scripts/check-mdx.ts <files>`, then `(Get-Item node_modules).Delete()`.
- Scaffold changes (`internal/cli/scaffolds/*`): also run `go test ./internal/cli -count=1 -run 'TestNewProjectEndToEnd|TestNewInertiaProjectEndToEnd'` (about 1 min; it lints and tests the generated apps).
- Merging: wait with a background `until` loop on `gh pr checks <n> --json bucket`, then `gh pr merge <n> --squash` by hand; after a merge the next PR is BEHIND: `gh pr update-branch <n>` and wait again. No merge script (Q-D-008), never `--admin`.
- Bash heredocs passed to `python -` turn `\\n` into real newlines: write multi-line edit scripts with the Write tool into the scratchpad and run them with `python <file>`; write with `newline='\n'`.

## Done this session

| VF | PR | State |
| --- | --- | --- |
| VF-035 composite keys (`orm.CompositeKeyer`, `orm.ErrKeyNotUnique`, `orm.PrimaryKeysOf`; schema `SmallInteger`, `Primary`, `StoredAs`) | #25 | merged (fcb7631) |
| VF-037 `install:billing -plans`, `STRIPE_CURRENCIES` + checkout `currency`, `currency_options` docs | #26 | merged (c12321b) |
| VF-046 admin `Resource.AfterSave(func(ctx, admin.Saved[T]) error)` + `Resource.Virtual(key, value)` | #27 | merged (3a174b7) |
| VF-041 `inertia.Config.ViewData` -> `RootData.Data`; `Page.URL` keeps the i18n path prefix | #28 | merged (425a277) |

Reports: `$ORCH/reports/VF-{035,037,046,041}.md`.

## Open

- Release v0.33.0 still BLOCKED (Q-D-006). Never run the release tool or push tags. Pending adoption list for `1-ADOPT-vukem-v0.33.0.md` (target `C:\dev\hexascore-app\hexascore-spike` while Phase 0 lasts) adds: VF-035 (`t.UUID("id").Primary()`, `t.Primary("user_id","item_id")`, `t.SmallInteger`, `StoredAs`; `Rating` gets `PrimaryKeys()`), VF-037 (regenerate billing with `-plans pro,business`, `STRIPE_CURRENCIES=brl,usd,eur`), VF-046 (ban/unban as `Virtual("banned")` + `AfterSave` with `s.Tx`), VF-041 (`ViewData` theme, drop `RestoreURL`).
- VF-047 (admin composite-key rows, nullable string filters): composite-key models now panic in `admin.NewResource` (PR #25); addressing them by a composite id is still open.
- VF-009 (`generated` ORM tag) still open; generated columns need `db:"-"`.
- Next VFs by severity: remaining mediums (check the index), then lows VF-042..045, VF-039, VF-014/016, VF-028 (likely closable by VF-041's `ViewData`: verify and mark).
- Do not touch vukem PR #6 (owner's). Do not touch other agents' servers or ports.
