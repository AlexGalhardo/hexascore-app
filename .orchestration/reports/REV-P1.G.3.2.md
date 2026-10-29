# REV-P1.G.3.2 (rodada 0)
VERDICT: CORRIGIDO_E_APROVADO
Branch: agent/especialista/P1.G.3.2-fix (from origin/main, commit c7d57d5).
Fixes:
- User/List of a currently banned owner was indexable; now X-Robots-Tag noindex + seo.indexable=false (SDD 14 §8). OwnerNickname also returns the ban flag. Test added in TestInertiaPages_UserLibraryLists/User/List.
- Library/Index router.get + next-page link used bare "/library" (dropped the pt-BR/es locale prefix) -> localized().
- Lists/Edit: keyboard focus was lost after move up/down (row node moves); now restored to the same row (opposite button at the ends). New vitest resources/js/pages/Lists/Edit.test.tsx.
- Lists/Edit items heading used lists.addToList ("Add to list") -> new key userLists.items (en/pt-BR/es).
Re-run: go test -count=1 -tags integration ./tests -run 'TestInertiaPages/(User|Library|Lists|Account)' -> 6 subtests PASS, ok 5.105s. bun run lint ok; bun run typecheck ok; bun run test 12 files / 51 tests passed (+1 new). golangci-lint run --build-tags integration ./... -> 0 issues.
axe: not run — axe-core is not a dependency (would be a new dep); keyboard sanity covered by the vitest above (jsdom), not by a real browser. Selects/inputs have visible labels; icon-free buttons have aria-labels.
New task: queue/senior/2-P1.G.3.2b.md (library counts unused, no "see all" per type without filter, no edition picker UI).
PRONTO_PARA_MERGE agent/especialista/P1.G.3.2-fix
