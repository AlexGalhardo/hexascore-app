# Task ADOPT-vukem-v0.33.0 (priority 1, rodada 0, model: Sonnet, senior pool)

Objective: adopt vukem `v0.33.0` (tag pushed 2026-10-29, release PR AlexGalhardo/vukem#35) in the app.

1. In `go.mod` bump `require github.com/AlexGalhardo/vukem` from `v0.32.0` to `v0.33.0`; `go mod tidy`; `go mod verify`. Leave `contrib/stripe` and `drivers/postgres` unless a newer tag of them exists (`git ls-remote --tags https://github.com/AlexGalhardo/vukem 'contrib/stripe/*' 'drivers/postgres/*'`).
2. Remove the app workarounds of the VFs v0.33.0 fixes. Each item's exact adoption line is in `.orchestration/reports/VF-<id>.md` ("Adoption:"); search the code for `workaround(vukem)` markers that cite the VF. In scope, in this order:
   - VF-016 + VF-048: replace the router-wide `Services.Errors` wrapper that adds `code` with `errs.New(...).WithExtension(...)` (`code`, `plan`, `limit`, `retryAfter`). Problem JSON bodies must stay byte-compatible with the legacy contract (API tests prove it).
   - VF-045: `app/covers/serve.go` Cache-Control wrapper becomes `storage.HandlerWith`.
   - VF-029: LGPD ZIP export uses `export.EncodeCSV(zipEntry, ...)` instead of `encoding/csv` + `export.SafeCell`.
   - VF-032: diff `app/accounts/lang/es/*.json` against the shipped catalogs; drop identical copies, keep app overrides.
   - VF-036, VF-043: drop the `workaround(vukem)` markers only (behaviour unchanged).
   - VF-003, VF-024, VF-025, VF-030, VF-031, VF-040, VF-041, VF-018, VF-019, VF-008, VF-046, VF-035: adopt when the app actually has the workaround named in the report; if the adoption is a refactor larger than ~150 changed lines or touches auth/security semantics, do NOT do it here: write `queue/senior/3-ADOPT-VF-<id>.md` with the scope and list it in the report.
3. Update `vukem-feedback.md` status of each adopted VF to `fixed in v0.33.0 (adopted <short sha>)`; `CHANGELOG.md` `[Unreleased]` Changed line.

App work: branch `agent/<your-role>/ADOPT-vukem-v0.33.0` from `origin/main`. Acceptance (quote real output): `go build ./...`; `golangci-lint run --build-tags integration ./...` → 0 issues; `TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable go test -tags integration ./...` → no FAIL; `bun run lint && bun run typecheck && bun run test`; `go mod verify` → all modules verified.
Report + `1-REV-ADOPT-vukem-v0.33.0.md`. Do not wait for CI.
