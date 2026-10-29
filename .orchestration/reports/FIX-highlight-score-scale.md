STATUS: concluida
Branch: agent/senior/FIX-highlight-score-scale. Files: app/catalog/{listing,highlights,listing_view}.go, app/catalog/score_scale_test.go (new), tests/home_highlights_test.go (expected 87.3 -> 8.7).
Fix: new const starsScore (gated, round(score::numeric)/10, same as scoreFields) used by Highlights and Summaries. Summaries feeds Catalog/Index, Entity/Show and lists; item page and /api/v1 already use scoreFields.
Acceptance: TestScoreScaleConsistent (87 -> 8.7 on highlights, summaries, item page) fails before, passes after; go test -tags integration ./... all ok after updating home test; golangci-lint run --build-tags integration: 0 issues; bun run test: 57 passed.
Risks: /api/v1 list/get not asserted directly in the new test (share scoreFields with item page).
