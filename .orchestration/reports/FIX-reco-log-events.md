# FIX-reco-log-events reco.impressions / reco.clicks as log events (D-35)
status: review
branch: agent/senior/FIX-reco-log-events @ 9f274e2 (worktree wt/senior-FIX-reco-log-events, from main)
files: app/recommend/related/events.go (was metrics.go), events_test.go (was metrics_test.go), app/controllers/item_controller.go, routes/web.go, docs/sdd/08-recommendation.md
acceptance:
- red: `go test ./app/recommend/related/ -run TestEvents` → build failed `undefined: Events`
- green: same → ok; `go test -tags=integration ./tests -run TestItemRelated` → ok; `golangci-lint run` on touched packages → 0 issues
design: the in-memory `Metrics` counters had no reader (nothing exported them), so they are removed. `related.Events{Log, ImpressionEvery}` writes INFO lines `event=reco_impression` (sampled 1 in 10 via crypto/rand, `sample_rate=10` attribute: count x sample_rate) and `event=reco_click` (always) with `surface`, `type`, `item` (item id). Wired in routes/web.go with app.Log, ImpressionEvery 10. Railway search: `@event:reco_click` / `event=reco_impression`.
risks: sampling rate is a constant in routes/web.go; the item controller path itself has no log-capture test (unit tests cover Events). Line endings: files kept LF.
changelog: Changed: reco impressions and clicks are structured log events instead of unused in-process counters

## Review (especialista, 2026-11-01)

verdict: CORRIGIDO_E_APROVADO
re-run (main merged in): `go test ./app/recommend/related -run TestEvents` ok; `go test -tags integration ./tests -run 'TestItemRelated|TestReviews|TestItem'` ok; `golangci-lint run -j 2` on recommend, controllers, routes -> 0 issues
checked: D-35 (Railway logs, no metrics export), no PII (item id, surface, type only), `?from=` sanitised through `FromSurface`, crypto/rand sampling fails open to logging, scope
fixes: 1. merge conflict with the P7 stack in `item_controller.go` (the `Item` struct keeps `Social` and takes `reco related.Events`) (9342611e)
PRONTO_PARA_MERGE agent/especialista/FIX-reco-log-events (merged d7a8c4e7)
