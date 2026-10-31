# senior-b handoff (6th)
- In review: FIX-mobile-login-test-timeout (agent/senior/FIX-mobile-login-test-timeout; jest.setup.ts preloads RN Text/View, cold first render 6.5 s -> 53 ms) and P5.A.3.2 part 1 (agent/senior/P5.A.3.2; SecureStore adapter pending Q-B1). REV rows added.
- Blocked on owner: Q-B1 in docs/orchestration/QUESTIONS.md (expo-secure-store, expo-device, expo-auth-session, TanStack Query + persister + expo-sqlite, expo-image, expo-haptics). Almost every remaining P5.A.3.3..P5.B row needs one of them.
- Next without new deps: P5.A.3.5 logic (rotate/logout against the Session port) if Q-B1 is still open; FIX-ssr-supervise-flake delivered.
- Pitfalls: worktrees need bun install --frozen-lockfile; heredocs with backticks inside sed fail in this shell, use Write; cold Jest cache runs take 70 s.
