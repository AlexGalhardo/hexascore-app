# Task FIX-highlight-score-scale (priority 1, rodada 0, model: Sonnet, senior pool)

Objective: highlight cards (Home sections and the new Entity/Show groups, any `catalog.Highlights`/listing summary) show the raw `items.score` while the item page divides it by 10 (`stars := math.Round(v)/10` in `catalog.LoadItemPage`), so the same item shows two different scores. Make every site and API read path use one helper for the 0-10 one-decimal scale (SDD 02 §6) and the 5-rating read gate; test first with a fixture whose stored score is 87 -> 8.7 everywhere (Home, Catalog/Index, Entity/Show, /api/v1 listItems/getItem).
Acceptance: the new test; `go test -tags integration ./...`; `golangci-lint run --build-tags integration ./...` 0 issues; `bun run test`.
App work: branch `agent/<your-role>/FIX-highlight-score-scale` from `origin/main`. `TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable`. Report + `1-REV-FIX-highlight-score-scale.md`. Do not wait for CI.
