# FIX-community-create-kinds-ui entity and edition creates in the suggest form

status: concluida
branch: agent/senior/FIX-community-create-kinds-ui @ b23b0a1 (stacked on agent/senior/P4.B.4.2 @ 7f87764)
files: `app/community/page.go` (`LoadCreatePage` takes `target`, `CreatePage.Target`, fields + `createOnly`, draft per target), `app/controllers/community/suggest.go` (`?target=`, `target` prop, redirect keeps `&target=`), `resources/js/pages/Community/SuggestCreate.tsx` + test, `resources/lang/{en,pt-BR,es}.json` (`community.create.targets.*`, `community.create.choose`, `community.fields.{entity_kind,work_id,language}`, `community.entityKinds.*`), `tests/community_create_kinds_test.go`, `docs/sdd/04-community-edits.md` §6 route row.
sdd: SDD 04 §7.6 create (P4.B.2.3b `checkCreate`), route table.

behaviour: on a create, a second tab row picks Title / Person, company or tag / Book edition (books only). `?target=entity` adds the kind select (person, organization, franchise, platform, genre); `?target=edition` adds the parent book search (`ItemPicker`, domain book) and the language code input. Submit stays disabled until the create-only fields are set; the server is still the source of truth (`checkCreate`). The body sends `targetKind` = target (merge stays `item`). Unknown target, an edition outside books, or a merge with a non-item target → 404.

acceptance:

- `go test -tags integration ./tests -run TestCommunityCreateTargetKinds -count=1` → red before the change, ok after (fields per kind, 404s, entity draft saved with redirect `?kind=create&target=entity` and restored only on the entity tab).
- `go test -tags integration ./tests -run 'TestCommunity|Suggest' -count=1` → ok; `./app/community/ -run 'Page|Create'` → ok.
- `bunx vitest run resources/js/pages/Community` → 17 passed (3 new: target tabs incl. edition only for books, entity create body, edition create picks book + language and blocks submit until set).

validation: `bun run lint` → clean; `bun run typecheck` → clean; i18n parity (vitest lib) → pass; `golangci-lint run --build-tags integration` community/controllers/tests → 0 issues; markdownlint SDD 04 → 0.
changelog: Added: suggest form can propose new people, companies, tags and book editions
roadmap: none (follow-up of P4.B.2.3b)
risks / follow-ups:

- No catalogue write added (drafts only; apply stays admin-only, covered by `TestSecurity_A01_ModeratorCannotApplyEntityCreate`).
- A restored edition draft shows the book id until it is re-picked (the draft stores only `work_id`).
- Entity creates still hang off a domain path (`/suggest/new/game?target=entity`); the domain is the record's `target_domain`, not a property of the entity.

QA (stack built from the stacked FIX-community-mail-optout-ui @ 2eab02c, which contains b23b0a1; :3121):

- `/suggest/new/book?kind=create&target=edition` shows the "What to add" tabs (Title, Person, company or tag, Book edition), the edition fields, the Book search and the language input; Save draft disabled until book and language are set.
- Picked "Irure" from the search, filled title, ISBN 13, `pt-BR`, a source URL, saved → `suggestions` row `create | edition | book | draft` with `work_id` (the picked book's uuid), `language` `pt-BR`, `external_ids.isbn13`, `title`.
- Follow-up wording nit: the item tab reads "Title", which can be confused with the Title field; the specialist may prefer "Game, movie, series or book".

## Review (especialista, 2026-10-31)
verdict: APROVADO
re-run (tab-label tip with main merged, contains kinds-ui): `bunx vitest run resources/js/pages/Community resources/js/lib` → 6 files, 25 passed (i18n parity included); `bun run typecheck`, `bun run lint` clean; `go build ./...`, `go test ./app/community/...` ok; `golangci-lint run --build-tags integration ./app/community/... ./app/controllers/community/... ./tests/...` → 0 issues
checked: SDD 04 §7.6 create and route table, authz (drafts only; apply stays admin-only, server `checkCreate` is the source of truth; unknown target and edition outside books 404), secrets none, sql (no new query; draft lookup bound), i18n (en/pt-BR/es parity), scope, tests
pending (Docker down): `go test -tags integration ./tests -run 'TestCommunityCreateTargetKinds|TestCommunity|Suggest'`
gate on main after merge (6fe871f): go build ok; full gate waits Docker
PRONTO_PARA_MERGE agent/senior/FIX-community-create-kinds-ui
