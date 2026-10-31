# senior-b handoff (3rd)
- Delivered in review: FIX-covers-base-state r2 (fb02136), FIX-e2e-nav-timeouts (4ba9a2f, also FIX-e2e-onboarding-hydration). Older reviews: see TASKS.md (mobile track P5.A.1.1 -> 1.2 -> 2.1 -> 3.1 stacked; P2.H.* web rows).
- FIX-e2e-nav-timeouts: 3rd consecutive e2e run still owed (needs free :3000; image hexascore-app:FIX-e2e-nav-timeouts is built from the rebased branch). Never `up` over the specialist's stack; the specialist did over mine twice.
- Next: P5.A.3.2 once P5.A.3.1 merges (SecureStore: check apps/mobile/package.json, else QUESTIONS.md; no new dependency without owner yes). Then P2.H.3.5, FIX-nojs-auth-post.
- Pitfalls: worktrees need `bun install --frozen-lockfile`; Windows sed lacks \t; python open() text mode is fine but gofmt after edits; never `docker compose down`; foreground bash waits above 180 s get backgrounded, use Monitor.
