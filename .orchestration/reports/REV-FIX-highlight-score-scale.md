REV FIX-highlight-score-scale: CORRIGIDO_E_APROVADO. Final branch: agent/especialista/FIX-highlight-score-scale-fix (e7c69f6, on top of agent/senior/FIX-highlight-score-scale).
Reviewed OK: starsScore = gated round(score::numeric)/10, same as scoreFields (item page, /api/v1); used by Highlights and Summaries (Catalog/Index, Entity/Show, lists); min_score filter already 0-10.
Fixed: score.outOf label still said "out of 100" / "de 100" (en, pt-BR, es) on ItemCard and SectionHighlights aria labels -> "out of 10"; CHANGELOG [Unreleased] Fixed entry (branch had none).
Re-run: go test -tags integration ./app/catalog -run TestScoreScale -> PASS TestScoreScaleConsistent; go test -tags integration ./... -> all ok; golangci-lint -> 0 issues; bun lint/typecheck clean, test 57 passed.
Merge note: CHANGELOG [Unreleased] Fixed also added by agent/especialista/P1.G.2.4-fix -> trivial conflict.
