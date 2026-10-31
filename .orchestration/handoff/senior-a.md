# senior-a handoff
Delivered this round (all in review): FIX-community-problem-codes, P4.A.3.2 (5b75cc0), P4.A.3.3 (8b841c0, stacked on 3.2).
Next rows: P4.A.3.1d (SuggestEdit release_date and relation fields; reuse resources/js/components/community/fields.ts), then add and claim P4.B.1.1 and following code-only P4.B microtasks (check developer-manual-todos for owner gates).
Not done: browser/QA-stack check and e2e runs for 3.2 and 3.3 (specs written: community-create, community-status; need E2E_COMMUNITY_OPEN=true and FLAG_COMMUNITY_SUBMISSIONS=true). Worktrees wt/senior-* remain until merge; worktree node_modules via bun install --frozen-lockfile.
Pitfalls: bash heredocs with big mixed content sometimes fail to parse in the tool, use the Write tool; golangci-lint fmt/run refuse parallel runs and can exceed 120s (run in background); python text mode writes CRLF.
