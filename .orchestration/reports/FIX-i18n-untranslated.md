# FIX-i18n-untranslated Translate untranslated pt-BR and es labels
status: concluida
branch: agent/estagiario/FIX-i18n-untranslated @ 6065fc7 (worktree `C:\dev\hexascore-app\wt\estagiario-FIX-i18n-untranslated`, base main 246ced8)
files: resources/lang/pt-BR.json, resources/lang/es.json
acceptance:
- `bun -e` flatten of `resources/lang/*.json` (before/after) → equal-to-en leaves: pt-BR 39 → 36, es 28 → 27 (`.orchestration/reports/QA-i18n-parity.md` source)
- `bunx vitest run packages/i18n/src/parity.test.ts` → Test Files 1 passed, Tests 5 passed, exit 0
- JSON.parse of en, pt-BR, es → ok
keys changed:
- pt-BR `billing.plans.free`: "Free" → "Grátis" (matches "Criar conta grátis" copy)
- pt-BR `categories.game.gameplay`: "Gameplay" → "Jogabilidade" (matches es)
- pt-BR `categories.game.storytelling`: "Storytelling" → "Narrativa" (matches es)
- es `billing.plans.free`: "Free" → "Gratis" (matches es copy "Empieza gratis")
decided not to change (row said "pt-BR 39, es 28"; the report's own "likely real miss" list is much shorter, I followed the list):
- Pro, Business (plan/tier names, used as product names in copy)
- Nickname (auth/account labels): "nickname" is a deliberate product term in both locales' body copy ("Escolha seu nickname"); owner to confirm
- Menu, Spam: same word in pt-BR/es
- brands, abbreviations, Status, API, Total, SDKs, Plan, `{count} item`/`series`: kept per report
- policy.* / terms.* untouched (DOC-i18n-legal, owner review)
validation: lint n/a (JSON only), typecheck n/a, tests packages/i18n parity ok, browser n/a (copy only; no UI run)
changelog: Fixed: pt-BR and es labels for the Free plan and pt-BR game categories were left in English
roadmap: no box
risks / follow-ups: owner to confirm "Nickname" and tier names Pro/Business are intentional in pt-BR/es. Repo has `bun.lock` untouched (`bun install --frozen-lockfile` in worktree).

## Review (especialista, 2026-10-31)

verdict: REJEITADO
re-run: legacy catalogues `legacy/apps/web/messages/pt-BR/categories.json` lines 4 and 9 (`"gameplay": "Gameplay"`, `"storytelling": "Storytelling"`), `legacy/apps/web/messages/{pt-BR,es}/billing.json` line 3 (`"free": "Free"`, beside `"pro": "Pro"`, `"business": "Business"`)
checked: the four changed values against the legacy app (parity target of v1.0.0), consistency with the kept plan names
rework:
1. The three pt-BR values and the es `billing.plans.free` value are the legacy wording, not misses: the plan names are product names (Free/Pro/Business) and pt-BR game categories keep the English gaming terms. Revert the four values; the branch then has no change and the row can close without a merge.
2. For the QA-i18n-parity list: treat a value equal to `en` that is also equal to the legacy locale value as intentional, so the next parity run does not flag these again.
3. Changing plan or category names is an owner copy decision; only the orchestrator files it in QUESTIONS.md if wanted.
