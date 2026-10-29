# ADOPT-vukem-v0.33.0
STATUS: concluida (parcial por escopo: itens grandes diferidos)
Branch agent/senior/ADOPT-vukem-v0.33.0 (commit 4a2d8ec + feedback commit). go.mod vukem v0.33.0; contrib/stripe v0.4.1 and drivers/postgres v0.5.1 are still the latest tags.
Done: VF-016/048 (13 NewKey sites use WithExtension("code"); CodedProblem deleted; APIEdge now sets type=base/code from the code member and also covers /app/v1); VF-029 (zipCSV uses export.EncodeCSV).
No-op: VF-045 (no app/covers in repo), VF-036/VF-043 (no markers in code), VF-032 (app es catalogs differ from scaffold ones: they hold app mail keys, kept), VF-003/025/030/031/041/018/019/008 (workaround not present in app).
Kept: exports_controller cooldown hand-written problem (web route returns JSON problem even for session guard; WithExtension would change that).
Deferred (queue/senior): 3-ADOPT-VF-024, 3-ADOPT-VF-040, 3-ADOPT-VF-046, 3-ADOPT-VF-035.
Acceptance: go build ./... ok; golangci-lint run --build-tags integration ./... -> 0 issues; go test -tags integration ./... -> all ok (tests 284s, catalog 150s); bun run lint/typecheck ok, bun run test 58 passed; go mod verify -> all modules verified.
Risk: problem type for web (non-/api,/app) JSON routes ratings/reports is now about:blank (tests pass; legacy contract is API-only).

REV (especialista): CORRIGIDO_E_APROVADO — reports/REV-ADOPT-vukem-v0.33.0.md (/app/v1 CORS removed).
PRONTO_PARA_MERGE agent/especialista/ADOPT-vukem-v0.33.0-fix
