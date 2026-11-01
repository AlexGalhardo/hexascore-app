# Handoff: especialista (2026-11-01, session 25)

Reason: assigned work done. STATUS: HANDOFF.

## Done this session (local `main`, no push, no tag)

| Rows | Verdict | Merge |
| --- | --- | --- |
| FIX-lighthouse-script-size / REV-FIX-lighthouse-script-size | CORRIGIDO_E_APROVADO: added `resources/js/lib/queryClientGuard.test.ts` (a3f07a24), a static guard that fails when a query hook caller is not wrapped in `withQueryClient` | a88d9350 |

- Both rows are `ready`.
- The CHANGELOG line and the review section in `reports/FIX-lighthouse-script-size.md` landed in 362f184a.
- The worktree and the branch are removed.

## Gate jobs on main

| Job | Result |
| --- | --- |
| `build` (a88d9350) | PASS, including SSR `/up` |
| `test-ts` (362f184a) | PASS |
| Lighthouse | See below |

Lighthouse ran in the clean container `lhci-runner:local` against a QA stack of the a88d9350 image. `lhci assert` exited 0. Script sizes:

| URL | Script |
| --- | --- |
| `/` | 134.8 KB |
| `/people/*` | 134.8 KB |
| `/games` | 138.1 KB |
| `/games/*` | 146.5 KB |
| `/books/*` | 146.5 KB |

The budget is 150 KB. The host-Chrome `local.sh` lighthouse job was not run: it is invalid on this host (Kaspersky).

## Open

1. The item-page margin is only 3.5 KB. Another eager import on Catalog or Show will break the budget again.
2. The session-24 follow-ups still stand:
   - IndexNow backlog when the key is set later.
   - No-JS auth with Turnstile on.
   - Migration id dated 2026-11-02.

## Pitfalls

- A merge commit needs a Conventional header. The hook rejects git's default `Merge branch` message, so finish the merge with `git commit -m "chore(...): merge ..."`.
- `lh.sh` and `lhsum.js` are in the session scratchpad. Usage: `lh.sh <port> <out> <repo>`, then `node lhsum.js <out>` prints the medians.
- The previous pitfalls still apply: `MSYS_NO_PATHCONV`, `TEST_DATABASE_URL` per call, local edits to `.env.example`, and the golangci wrapper with `setup-windows`.
