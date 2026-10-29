# REV-P1.G.3.1
VERDICT: CORRIGIDO_E_APROVADO (rodada 0; reviewed df04924, already merged in b00deb5)
Re-run on origin/main 0d82900 (Postgres 5433):
- bunx vitest run resources/js/components/search -> 2 files, 5 passed
- go test -tags integration ./tests/ -run 'TestInertiaPages/Search' -> PASS Search/grouped, Search/typed; ok 4.46s
- golangci-lint run --build-tags integration ./... -> 0 issues; bun lint/typecheck/test -> 43 passed
Bug fixed: Prev/Next partial reloads used only:[query,results,total]; searchId is new per request and the server records clicks only against that search's result_refs (search.RecordClick), so every click on page 2+ beaconed the stale page-1 searchId and was dropped. Added "searchId" to both `only` lists (+ why comment).
Fix branch: agent/especialista/P1.G.3.1-fix (1f0537e, from origin/main); lint/typecheck/test green (43 passed).
Non-blocking: highlight() slices original text by lowercase index (misaligns on chars whose lowercase changes length, e.g. Turkish I); suggest clicks not reported (documented). didYouMean link not locale-prefixed (pre-existing).
Out-of-scope bug found: duplicate top-level "contact" in en/es/pt-BR.json (from P1.F.3.2) hid contact.body/topics/hint -> fixed in agent/especialista/P1.F.3.2-fix, see reports/P1.F.3.2-fix.md.
PRONTO_PARA_MERGE agent/especialista/P1.G.3.1-fix
