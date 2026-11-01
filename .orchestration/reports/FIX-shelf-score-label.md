# FIX-shelf-score-label "Your score" on other users' ShelfCards

status: concluida
branch: agent/senior/FIX-shelf-score-label @ 62e82586 (from local `main` @ 2943b636; worktree `wt/impl-web-search-palette`)
files: resources/js/components/library/ShelfCard.tsx, ShelfCard.test.tsx, resources/js/pages/Library/Index.tsx (one `own` prop), resources/lang/{en,pt-BR,es}.json (additive `library.ratedScore`)
acceptance:

- Red first: `bunx vitest run resources/js/components/library/ShelfCard.test.tsx` → 2 failed | 2 passed (other user's stars printed `library.yourScore`, and null stars printed "Not rated by you").
- Fix: ShelfCard takes `own?: boolean`, default `false`. With `own`, labels are "Your score {stars}" / "Not rated by you". Without it, the label is "Score {stars}" (`library.ratedScore`: Score / Nota / Nota) and an unrated item prints nothing. Library/Index passes `own`. User/List passes no stars, so it needs no change. User/Profile.tsx is not edited (P7.A.1.3b lock) and gets the safe default.
- After: same command → 4 passed.

validation: `bun run lint` ok, `bun run typecheck` ok, `bun run test` 150/150 tests pass. The 2 failed files are `scripts/ci/sdk-{job,release}.test.ts`, which import `bun:test`; that failure is pre-existing on `main` and unrelated. Browser on QA :3140, built from this branch: `/u/demoaaab?tab=ratings` reads "Score 7.2" (en) and "Nota 7,2" (pt-BR), and `/library` reads "Not rated by you" (own). Shots: `reports/shots/impl-web/shelf-score-{profile-en,profile-pt-br,library-en}-1440-{light,dark}.png`.
changelog: Fixed: a public profile no longer labels another user's rating as "Your score".
roadmap: no box
risks / follow-ups:

- On the viewer's own public profile (`/u/<me>`), the cards now say "Score" instead of "Your score". That is correct but less personal. When P7.A.1.3b releases `User/Profile.tsx`, pass `own={isOwner}` there (a one-line follow-up for the profile restyle).
- `library.score` ("Your score: {stars}") looks unused. It was left alone.
- `IMPL-web-social-v1` ActivityCard already passes no `stars` and prints `social.feed.score` itself, so the two branches merge without conflict. Expect a trivial additive merge in `resources/lang/*.json`.
- Shared `node_modules` in the main checkout had been relinked into the deleted `wt/especialista-web-batch` store, which broke vitest in every worktree. It was repaired with `bun install --frozen-lockfile` in the main checkout. package.json and bun.lock are unchanged.

## Review (especialista, 2026-11-01)

verdict: APROVADO
checked: `own` defaults to false, so every caller except `Library/Index` (the viewer's own library) gets the neutral `library.ratedScore`; ActivityCard (feed) no longer says "Your score"; unrated items of other users show nothing; key added in en, pt-BR, es. `bunx vitest run resources/js/components/library resources/js/components/social resources/js/pages/Library` 5 files / 16 tests, tsc, biome.

PRONTO_PARA_MERGE agent/senior/FIX-shelf-score-label
