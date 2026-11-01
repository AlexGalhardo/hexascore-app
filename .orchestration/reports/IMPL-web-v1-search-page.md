# IMPL-web-v1-search-page Search page in design v1

status: concluida
branch: agent/senior/IMPL-web-v1-search-page @ 6501565 (stacked on agent/senior/IMPL-web-v1-auth @ 5246eee, merged in)
files: resources/js/pages/Search/Index.tsx, resources/js/components/search/ResultGroup{,.test}.tsx, resources/js/components/home/TypeSection.tsx (exports ACCENT_BG/ACCENT_TEXT), resources/lang/{en,pt-BR,es}.json (additive `search.*`: count, allIn.*, nothingFound, checkSpelling, didYouMean, orBrowse, people)
acceptance:
- `bunx vitest run resources/js/components/search` → 12/12 (new: row shows "Game · 1986" and a score badge only when the hit carries `score`; click beacon test kept)
- `bun run lint` → 185 files, no fixes; `bun run typecheck` → exit 0; `bun run test` → 34 files / 128 tests passed
- Browser (QA :3140, full page): `.orchestration/reports/shots/impl-web/{desktop,mobile}-{light,dark}-_search{,_q_the,_q_a,_q_diabol,_q_zel_type_game}.png`. Compared with `web-{desktop,mobile}{,-dark}__search__{default,empty-query,too-short,no-results}.png`: serif "Results for “q”" h1, 2 px ink search field with icon + primary Search button, bordered filter chips (active = ink fill), groups under a 2 px rule with accent square, plural section name and "n results", rows with 2:3 cover, serif title, "Type · year" meta, "All n games for “q” →" link, "People and companies" panel on the right (below on mobile); empty query and too short as the italic serif line under the rule; no results: "Nothing found for “q”.", spelling hint, "Did you mean" box (when the server sends `didYouMean`), "Or browse a section" with the four accent links, then the existing "Suggest a new …" links.
- The `_q_*` shots predate the last fix commit (6501565: plural section names on chips and headings, sans-serif counts and panel title, which inherited the global h2 serif); re-shot on the library-lists build.
validation: lint ok, typecheck ok, vitest ok, browser ok (light/dark x 1440/390)
changelog: Changed: the search page follows design v1 (grouped results with section accents, filter chips, people and companies panel, no-results state with "did you mean" and section links).
roadmap: no box
risks / follow-ups:
- Rows have no year, creator or score because `SearchHit` does not carry them: FIX-search-hit-meta opened for a senior. `ResultGroup` already renders `year` and `score` when present (badge only when the field exists, so today no grey "–" badge is shown on every row).
- The "Try: The Witcher IV, Frozen 3, …" chips of the empty-query frame are left out: no backend list of suggested queries, and hard-coded titles would go stale. Could reuse the trending list of FIX-search-suggest-score later.
- The type filter now keeps `type` when a new query is submitted from the page (hidden input), matching the chips.
- No new binary assets.

## Review (especialista, 2026-11-01)
verdict: CORRIGIDO_E_APROVADO
re-run: stack merged in agent/especialista/IMPL-web-batch2 (main merged in); `bun run lint` 185 files ok; `bun run typecheck` ok; `bun run test` green except 4 load timeouts in the full run (SearchPalette, SuggestCreate, Lists/Edit), all 15 green when re-run alone; mobile `src/i18n` Jest 7 ok
checked: sdd 09 §7.4, authz (no server change; Google link is the existing /auth/{provider}/redirect), secrets (none), i18n (additive keys in en/pt-BR/es; parity test green), scope (shared Field/PasswordField/SubmitButton restyle noted in the auth report), tests
fixes: 1. mobile bundled catalogues recopied after the new auth.* keys (0c4927e)
gate on main after merge: see FIX-app-item-cover-urls review; lighthouse budgets already red on main (FIX-lighthouse-budgets-main)
PRONTO_PARA_MERGE agent/especialista/IMPL-web-batch2 (merged f99bafd)
