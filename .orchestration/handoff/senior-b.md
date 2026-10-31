# senior-b handoff (7th)
- Delivered, all in review with REV rows, one stacked chain: P5.A.3.2 (done; SecureStore vault + Q-B1 deps + ADR 0019) <- P5.A.3.3 (login) <- P5.A.3.4 (Google flow logic) <- P5.A.3.5 (rotate, logout, update screen). Merge in that order. Worktrees wt/senior-P5.A.3.{2,3,4,5}.
- Next: P5.A.4.1 (TanStack Query + SQLite persister; deps are installed), then A.4.2, A.4.3, P5.B.x. Stack on agent/senior/P5.A.3.5.
- Open follow-ups: root layout wiring (session.restore, startForegroundRotation, 426 redirect, auth gate); Google picker needs the owner's Android OAuth client id; Maestro/staging/real-device gates skipped.
- Pitfalls: RNTL 14 render/fireEvent are async (await them); catalogues are generated from resources/lang (edit there, run `bun run catalogues`); worktrees need `bun install --frozen-lockfile`; Jest cold runs 30-70 s.
