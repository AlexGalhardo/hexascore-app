# FIX-community-create-tab-label

Branch `agent/senior/FIX-community-create-tab-label` @ 3e2ebcb, stacked on `agent/senior/FIX-community-create-kinds-ui` @ b23b0a1 (merge that first).

## Change

`community.create.targets.item` (the first target tab on `/suggest/new/{domain}?kind=create`):

| locale | before | after |
| --- | --- | --- |
| en | Title | Game, movie, series or book |
| pt-BR | Título | Jogo, filme, série ou livro |
| es | Título | Juego, película, serie o libro |

Only the catalogue values change; the key and the component are untouched. The mobile catalogues do not bundle the `community` namespace, so no `catalogues` run is needed.

## Tests (written first, red, then green)

- New test in `resources/js/pages/Community/SuggestCreate.test.tsx` asserts the three catalogue values (the component tests mock `t`, so the value is checked on the catalogues).
- `bunx vitest run resources/js/pages/Community/SuggestCreate.test.tsx packages/i18n/src/parity.test.ts`: before the catalogue edit 1 failed / 7 passed; after, 2 files, 13 tests passed.
- `bunx biome check` on the test file: clean. `bunx tsc --noEmit -p .`: no errors.

## Changelog line (for the specialist)

- Changed: the community create form's item tab reads "Game, movie, series or book" (en/pt-BR/es) instead of "Title".

## Review (especialista, 2026-10-31)
verdict: APROVADO
re-run (tab-label tip with main merged, contains kinds-ui): `bunx vitest run resources/js/pages/Community resources/js/lib` → 6 files, 25 passed (i18n parity included); `bun run typecheck`, `bun run lint` clean; `go build ./...`, `go test ./app/community/...` ok; `golangci-lint run --build-tags integration ./app/community/... ./app/controllers/community/... ./tests/...` → 0 issues
checked: SDD 04 §7.6 create and route table, authz (drafts only; apply stays admin-only, server `checkCreate` is the source of truth; unknown target and edition outside books 404), secrets none, sql (no new query; draft lookup bound), i18n (en/pt-BR/es parity), scope, tests
pending (Docker down): `go test -tags integration ./tests -run 'TestCommunityCreateTargetKinds|TestCommunity|Suggest'`
gate on main after merge (07807ac): go build ok; full gate waits Docker
PRONTO_PARA_MERGE agent/senior/FIX-community-create-tab-label
