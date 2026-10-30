# Handoff especialista (2026-10-29, session 14)
STATUS: STOPPED (owner ended session). No task in progress. Worktree on agent/especialista/P2.F.4.1 (pushed, clean). Stack hx-especialista down.
Done: REV-FIX-e2e-signup-load APROVADO (PRONTO_PARA_MERGE agent/senior/FIX-e2e-signup-load); P2.F.4.1 delivered (self, PRONTO_PARA_MERGE, needs sampling).
Next in queue (in order): 5-P2.G.1.1, 6-P2.E.1.2, 7-P2.C.3.3. Skip 9-*. Start: git fetch origin; git switch -c agent/especialista/<ID> origin/main.
Pitfalls (new): e2e on isolated stack needs E2E_APP_CONTAINER=hx-especialista-app-1 (mail helper reads docker logs) besides E2E_BASE_URL=http://127.0.0.1:3200; internal/search TestSearchSQLConfined forbids search_* SQL outside internal/search (use search.Rebuild/DrainSuggest); enum columns need ::text before coalesce(...,''); sqlclosecheck wants defer Close (legacyimport.collect helper).
Older pitfalls: signup form_token; biome format TS; use Write tool not heredoc python with backticks; playwright -c qa/e2e/playwright.config.ts; golangci-lint --allow-parallel-runners; pgtest tests behind //go:build integration.
Tools: scratchpad up.sh/ov.yml (hx-especialista on 127.0.0.1:3200).
