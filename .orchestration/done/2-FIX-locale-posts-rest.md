# Task FIX-locale-posts-rest (priority 2, rodada 0, model: Sonnet)
Objective: finish the FIX-locale-redirects pattern: every remaining client-side Inertia post/put/patch/delete under resources/js (items rating/status/time-to-beat/edition, lists, contact, notifications, library, reports, billing, data export) must call useLocalized() (resources/js/lib/useLocalized.ts) so redirects keep the /pt-br or /es prefix (SDD 09 §7.1).
Base: origin/main after agent/especialista/FIX-locale-redirects-fix is merged (it adds useLocalized).
Find: `rg -n "\.(post|put|patch|delete)\(\"/" resources/js` -> every hit converted (except JSON fetch() calls to /api, /app/v1, /challenge).
Acceptance: rg above returns only intentional hits listed in the report; bun lint/typecheck/test green; one Go test in tests/locale_redirects_test.go (integration tag) asserting a list create under /es redirects under /es; playwright lists + contact specs on an isolated stack (COMPOSE_PROJECT_NAME=hx-senior) green.
