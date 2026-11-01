# IMPL-web-v1-library-lists Library, lists and shelf cards in design v1

status: concluida
branch: agent/senior/IMPL-web-v1-library-lists @ e2525f3 (stacked on agent/senior/IMPL-web-v1-search-page @ 6501565)
files: resources/js/components/library/ShelfCard{,.test}.tsx, resources/js/pages/Library/Index{,.test}.tsx, resources/js/pages/Lists/Index.tsx, resources/js/pages/User/List.tsx, resources/lang/{en,pt-BR,es}.json (additive `library.yourScore`, `notRated`, `shelvesOf`, `titleCount`)
acceptance:
- `bunx vitest run resources/js/pages/Library resources/js/pages/Lists resources/js/components/library` → 10/10 (ShelfCard: "Type · year · status" meta, "Your score 8.5" / "Not rated by you"; library: per-status heading with the title count)
- `bun run lint` → 185 files, no fixes; `bun run typecheck` → exit 0; `bun run test` → 34 files / 128 tests passed (run before e2525f3, which only adds `font-normal` to the select class)
- Browser (QA :3140, logged in as the demo admin, nine `item_status` rows inserted into `hexascore_qa_implweb` for the shots): `.orchestration/reports/shots/impl-web/{desktop,mobile}-{light,dark}-_{library,lists}-in.png`. Compared with `web-{desktop,mobile}{,-dark}__library__default.png`: eyebrow "<name>'s shelves", 64 px serif h1, title count on the right, labelled v1 selects + primary Apply, one section per status under a 2 px rule with serif heading and "n titles", three-column bordered cards (2:3 cover, serif title, "Game · 2009 · Want to play", "Not rated by you"), "See all" link in primary. Lists index and public list page use the same heading, rule and card look; the new-list form uses the v1 Field/SubmitButton.
- The library shots were taken on the build before e2525f3 (selects still bold there).
validation: lint ok, typecheck ok, vitest ok, browser ok (light/dark x 1440/390)
changelog: Changed: the library, your lists and public list pages follow design v1 (status sections, shelf cards with your score).
roadmap: no box
risks / follow-ups:
- No community score badge on the cards: library items, profile shelves and list entries carry only the viewer's `stars`. FIX-library-item-score opened for a senior; `ShelfCard` already takes a `score` prop. The header shows "n titles" only; "· n rated" waits for the same row.
- Section names come from the existing `social.status.shelves.*` strings ("Want to", "In progress", "Completed"); the frame says "Want to play, watch or read". Left as is (copy change for the text-editor agent if wanted).
- The profile page (`User/Profile.tsx`) gets the new ShelfCard but no other restyle; Lists/Edit was not restyled beyond the shared Field/SubmitButton (no frame for it).
- No new binary assets.

## Review (especialista, 2026-11-01)
verdict: CORRIGIDO_E_APROVADO
re-run: stack merged in agent/especialista/IMPL-web-batch2 (main merged in); `bun run lint` 185 files ok; `bun run typecheck` ok; `bun run test` green except 4 load timeouts in the full run (SearchPalette, SuggestCreate, Lists/Edit), all 15 green when re-run alone; mobile `src/i18n` Jest 7 ok
checked: sdd 09 §7.4, authz (no server change; Google link is the existing /auth/{provider}/redirect), secrets (none), i18n (additive keys in en/pt-BR/es; parity test green), scope (shared Field/PasswordField/SubmitButton restyle noted in the auth report), tests
fixes: 1. mobile bundled catalogues recopied after the new auth.* keys (0c4927e)
gate on main after merge: see FIX-app-item-cover-urls review; lighthouse budgets already red on main (FIX-lighthouse-budgets-main)
PRONTO_PARA_MERGE agent/especialista/IMPL-web-batch2 (merged f99bafd)
