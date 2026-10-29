# REV-P1.G.4.1
VERDICT: CORRIGIDO_E_APROVADO (rodada 0; reviewed b837a88, already merged in 3eea98e)
Re-run on origin/main 1a0058b:
- clean checkout: go test -tags integration ./tests -run 'Robots|Docs|Legal|ErrorPage|Shell' -> FAIL [setup failed] `resources\resources.go:18:38: pattern all:public: no matching files found` (BUG 1)
- after fix: same -> ok; go test ./... -> ok; golangci-lint --build-tags integration -> 0 issues; bun lint/typecheck/test -> 46 passed; bun run build:docs -> no diff (docs.json fresh)
BUG 1 (high, breaks CI/fresh clones; Docker unaffected because the assets stage creates public/build): deleting resources/public/robots.txt left resources/public untracked-empty, so //go:embed all:public fails. Fix: resources/public/.gitkeep (dotfiles never served).
BUG 2 (low, hardening): build-docs.ts allowed `/\host` links (`(?!\/)` only blocks `//`); browsers resolve `/\evil` to another host. Regex now `(?![/\])`; test added (red before, green after).
Escaping otherwise sound: all source escaped before tags, fixed tag set, heading ids slugified to [a-z0-9-], code lang sanitised, href attr double-quoted (entities stay inside). Docs.json is repo-authored content.
Fix branch: agent/especialista/P1.G.4.1-fix (c13deac, 02030cb) from origin/main.
PRONTO_PARA_MERGE agent/especialista/P1.G.4.1-fix
