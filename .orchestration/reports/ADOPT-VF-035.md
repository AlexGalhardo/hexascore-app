# ADOPT-VF-035 report
STATUS: parcial (model side adopted; migration rewrite deliberately not done)
Branch: agent/senior/ADOPT-VF-035 (pushed, origin/main merged). Files: app/models/{rating.go,scoring.go,rating_test.go}, vukem-feedback.md, CHANGELOG.md.
- Rating declares `PrimaryKeys() []string{"user_id","item_id"}` (orm.CompositeKeyer); RatingRead inherits it via embedding; the `PrimaryKey()="user_id"` hack is gone. Test TestRatingCompositeKey.
- NOT rewritten: items migration (20261029100001, in tag v0.1.0) and scoring migration (20261029120000, in v0.2.0) are already released; task says rewrite only unreleased ones and runbook says forward-only. Also the builder lacks partial/GIN indexes, array/numeric(7,4)-style types and column CHECKs used there, so s.Raw would remain for part. Decision for orchestrator: if pre-launch rewrite is allowed, a follow-up needs a pg_dump -s diff before/after.
- RatingFlag keeps single-column PrimaryKey(): vadmin.NewResource panics on composite models (admin lists it).
- Writes to ratings stay raw SQL (INSERT .. ON CONFLICT, DELETE .. RETURNING overall); orm.Update/Delete not used.
- go build ok; golangci-lint 0 issues; `go test -tags integration ./...` no FAIL (19 ok).
