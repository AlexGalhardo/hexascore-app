# REV-ADOPT-VF-035 (rodada 0)
STATUS: APROVADO
Branch: agent/senior/ADOPT-VF-035 (no changes); merge into current origin/main conflicts only in CHANGELOG.md [Unreleased] (trivial, keep both lines).
Decision check: leaving the released migrations (items in v0.1.0, scoring in v0.2.0) untouched is correct: migrations are forward-only (SDD 11 §7.2, runbook) and rewriting released DDL would change checksums vukem records; the builder also lacks the partial/GIN indexes and CHECKs used there. A pre-launch rewrite is an owner/architect call, not needed.
Verified: `Rating.PrimaryKeys()` = (user_id, item_id), RatingRead inherits; RatingFlag keeps single key because vukem admin.NewResource refuses composite keys (vukem admin/resource.go:158, already tracked in VF-047 "composite keys open"). Ratings writes stay raw SQL.
Re-run: `go test ./app/models/` ok; `golangci-lint run --build-tags integration ./...` 0 issues; `go test -tags integration ./tests/scoring/` ok (63.8s).
Full `go test -tags integration ./...` exit 0, no FAIL.
