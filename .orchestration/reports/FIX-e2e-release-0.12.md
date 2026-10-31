# FIX-e2e-release-0.12 (FIX-e2e-i18n-switcher, FIX-e2e-keyboard-mobile-header, FIX-e2e-ttb-reload, FIX-e2e-lint-innertext)
status: concluida
branch: agent/senior/FIX-e2e-release-0.12 @ 67da7ed
files: qa/e2e/specs/i18n.spec.ts, keyboard.spec.ts, time-to-beat.spec.ts, lists.spec.ts, ssr-nojs.spec.ts
acceptance:
- full `bun run e2e` on stack sb (port 3140, branch tree) -> 684 passed, 12 failed (12.2m); the 12 are csp.spec.ts:24 and passkeys.spec.ts:11 in all 6 projects = senior-a rows FIX-csp-inline-style and FIX-passkeys-register-e2e. No other failure.
- `bun run lint` -> ok (innerText warnings gone, no rule disabled)
root causes (all stale specs, app not regressed):
- i18n-switcher: language list is inside the closed header `<details>` menu, links not in the a11y tree; spec opens "More" first. SDD 09 does not require an always-visible switcher.
- keyboard-mobile-header: `.first()` matched the desktop nav link (display:none below md); spec uses `:visible`. Mobile section row is a real `<nav>` of links, keyboard path holds.
- ttb-reload: uses `reload()` helper. lint-innertext: textContent / text-node walk.
validation: lint ok, typecheck ok, e2e as above
changelog: Fixed: e2e specs aligned with the redesigned header; Biome innerText warnings removed
roadmap: no box
risks / follow-ups: no app code touched (no inline style added). Remaining red: senior-a's csp and passkeys rows.

## Review (especialista, 2026-10-31)

verdict: APROVADO
re-run: `bunx biome check qa/e2e` → 39 files clean; `bun run typecheck` → ok; full e2e in the Release gate (reports/RELEASE-v0.12.0-gate.md)
checked: specs only (no app code), no rule disabled, leak check keeps the same regex on text nodes, `:visible` keeps the keyboard path on the mobile header
merged: 7500d45 on local `main`
PRONTO_PARA_MERGE agent/senior/FIX-e2e-release-0.12
