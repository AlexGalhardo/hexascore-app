# FIX-covers-route GET /covers/{key...} from the bucket
status: concluida
branch: agent/senior/FIX-covers-route @ a455a53
files: routes/web.go (one route line), app/controllers/covers_controller.go, tests/covers_route_test.go
acceptance:
- go test ./tests -run 'TestCoversRoute|TestSecurity_A01_CoverKey' -> ok (serve + immutable header, 404 missing, 404 without bucket, traversal variants)
validation: gofmt ok, go build ok, go vet ok, golangci-lint 0 issues; browser/QA stack not run (no real bucket object; demo SVGs still served by the static middleware, which runs before the router)
design: vukem storage.HandlerWith (https://vukem.dev/docs storage page, disks and serving) with Options.CacheControl = "public, max-age=31536000, immutable"; disk = drivers/s3 when S3_BUCKET set (same config as cutover-verify), else 404. Key validated by catalog.CoverURL, then mapped to the covers/ object path; the handler adds dotfile/backslash/containment checks. Route registered before the /{path...} catch-all.
changelog: Fixed: GET /covers/{key} now streams imported covers from the private bucket with immutable caching.
roadmap: no box
risks / follow-ups: unit tests use storage.Memory, not a live S3/RustFS object; reviewer may check on the RustFS compose stack. ETag depends on the S3 driver exposing ETag().

## Review (especialista, 2026-10-31)

verdict: CORRIGIDO_E_APROVADO
re-run: `go test ./tests -run 'TestCoversRoute|TestSecurity_A01' -count=1` → FAIL on the branch (`TestSecurity_A01_HexaRouteStackInventory`: unclassified route GET /covers/{key...}), ok after the fix
fixes: 1. `tests/testdata/route_inventory.txt` classifies `web GET /covers/{key...}` (d2f326b on agent/especialista/FIX-covers-route-fix)
live check (image 07c4129 + RustFS bucket `esp-covers`, port 3111): `GET /covers/esp/test.jpg` → 200 image/jpeg, `Cache-Control: public, max-age=31536000, immutable`, `CSP: sandbox; default-src 'none'`, nosniff, no Set-Cookie; missing key → 404; `..%2F`, `%2e%2e/`, `..%5C` → 404 (bucket-root `secret.txt` never served); `/covers/demo-book-0.svg` → 200 SVG from the static middleware (demo wins); without S3 (stack :3110) → 404
finding (not blocking): the S3 driver sends no ETag and no Last-Modified, so `If-None-Match` gets 200, not 304. The immutable header makes revalidation moot for content-addressed keys; follow-up for the vukem S3 driver (`storage.HandlerWith` ETag) left to the orchestrator
checked: sdd 11 §7 (private bucket, app-served covers), authz (public route, no session write), traversal (`TestSecurity_A01_CoverKeyTraversal`), secrets (disk from `S3_*` config, error logged without values)
merged: fd08b26 on local `main`
PRONTO_PARA_MERGE agent/especialista/FIX-covers-route-fix
