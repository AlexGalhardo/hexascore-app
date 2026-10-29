# Handoff especialista (2026-10-29, session 4)
STATUS: HANDOFF (heavy context). Queue left: 1-REV-P1.H.2.1b.md (untouched).
Done this session (all CORRIGIDO_E_APROVADO, pushed):
- REV-P1.F.2.1 -> agent/especialista/P1.F.2.1-fix (banned owner profile noindex; merge test for user_item_editions).
- REV-P1.G.3.2 -> agent/especialista/P1.G.3.2-fix (banned owner list noindex; Library/Index locale prefix; Lists/Edit keyboard focus + heading key). New task queue/senior/2-P1.G.3.2b.md.
- REV-P1.H.3.1 (r1 + orchestrator r2) -> agent/especialista/P1.H.3.1-fix (from the unmerged senior branch; merge still waits Q-H31). VF-NEW-P1.H.3.1 filed.
Pitfalls (new):
- Both G.3.2-fix and H.3.1-fix touch CHANGELOG [Unreleased]: expect a trivial merge conflict.
- contrib/stripe: Subscription has custom UnmarshalJSON; embedding it to add fields silently drops them. Use Client.Call + raw JSON (app/billing/stripe_calls.go).
- golines: run `golangci-lint fmt ./tests/... ./app/...` instead of hand-wrapping.
- markdownlint-cli2 on one file may lint CHANGELOG too (config globs).
- No axe-core installed: a11y checks only via vitest/jsdom keyboard tests.
Previous pitfalls (handoff session 3) still apply.
Worktree C:\dev\hexascore-app\wt\especialista on branch agent/especialista/P1.H.3.1-fix, clean.
Skills: karpathy-guidelines/caveman not loaded (context economy).
