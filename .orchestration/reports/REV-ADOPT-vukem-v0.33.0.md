# REV-ADOPT-vukem-v0.33.0 (rodada 0) — CORRIGIDO_E_APROVADO
Final branch: agent/especialista/ADOPT-vukem-v0.33.0-fix (517c663, on top of senior branch).
Finding (security/SDD): APIEdge sent /app/v1 through the public-API CORS handler → `Access-Control-Allow-Origin: *` on /app/v1, contradicting SDD 13 l.96 ("/app/v1 sends no CORS headers"). Red proof on senior code: `api_edge_test.go:46: /app/v1 Access-Control-Allow-Origin = "*", want none`.
Fix: /app/v1 gets only the problem-code writer (no CORS); doc comment + CHANGELOG line updated; tests: /app/v1 no-CORS + coded 404 type, list-limit 409 asserts code+type, item-ratings-limited 429 asserts type (were only status / code substring).
Checked OK: 13 WithExtension sites; export ZIP via export.EncodeCSV keeps SafeCell guard (vukem v0.33.0 CSVWriter.write; tests/export_test.go asserts `'=HYPERLINK("x")`); no frontend reads problem `type`, so web routes falling to about:blank is harmless.
Re-run: go build ./... OK; go mod verify → `all modules verified`; golangci-lint --build-tags integration → `0 issues.`; go test -tags integration ./... → no FAIL/panic lines (exit 0); targeted `TestListLimits|TestRatingItemLimitedProblemCode|TestAPIEdge` → ok.
Merge note: agent/especialista/P2.A.3.2 also edits app/library/export.go (NewExports hunk only) — different hunks, expected clean.
