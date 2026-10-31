# Handoff: especialista (2026-10-31, session 6)

Reason: six reviews done (five v0.12.0 release blockers plus one vukem PR) and the full Release gate run; context is heavy. STATUS: HANDOFF.

## Freeze

Lifted. v0.12.0 is released: PR #60 squash 219c873, tag pushed. Local `main` carries `origin/main` (merge b994c7b). E2E no longer gates anything; run only the specs a change touches.

## Done this session

| Row | Verdict | Result |
| --- | --- | --- |
| REV-FIX-e2e-release-0.12 | APROVADO | merged 7500d45 |
| REV-FIX-csp-inline-style | APROVADO | merged f3846ee |
| REV-FIX-covers-route | CORRIGIDO_E_APROVADO (route_inventory line, d2f326b); live RustFS check ok, the S3 driver sends no ETag | merged fd08b26 |
| REV-FIX-passkeys-register-e2e | APROVADO (CI origin localhost:3000 consistent) | merged 07c4129 |
| REV-FIX-app-client-regen | APROVADO (codegen diff from the image is clean) | merged 7a5bf01 |
| REV-FW-otlp-pr | CORRIGIDO_E_APROVADO (last-updated.json, 3e2095d2) | vukem PR #104, see below |

Orchestration commits: ec136c3, 83bec2b. Gate: `reports/RELEASE-v0.12.0-gate.md`. Every job passed on 7a5bf01 except these:

- commitlint: 3 old commits; solved by the squash.
- Lighthouse: not run, by owner stop.
- k6: the p95 threshold missed locally at 358 ms against 300 ms (Docker Desktop).

e2e scored 695/696 with one search timeout under load; it is no longer part of the gate (owner).

## Open

1. Done: vukem PR #104 squash-merged as daf6d302 (33/33 green), REV-FW-otlp-pr `ready`, worktree removed.
2. NEXT: the merge queue, in the orchestrator's order:
   1. P4.A.3.1c
   2. FIX-nojs-auth-post
   3. DOC-sdd-flag-names
   4. FIX-scoring-burst-clock
   5. P5.A.1.1, then 1.2, then 3.1 (catalogue recipe in the session-5 handoff, see git history of this file)
   6. the DOC-roadmap-p7-refs patch
   7. QA-e2e-speed, if delivered

   Then run the gate (scratchpad `gate.sh <sha>`), tick ROADMAP, write the CHANGELOG lines under `[Unreleased]`, and set the rows `ready`.
3. Then the vukem rows REV-VF-NEW-AUDIT-6, REV-VF-NEW-FIX-gate-timeouts (d88f52b2; already pre-read: env var `PGTEST_DROP_TIMEOUT`, no API change) and REV-VF-NEW-AUDIT-1 (vukem PRs, same flow as #104), then REV-P2.H.1.4/1.5 and REV-P2.H.3.5. The pre-read notes for P2.H.3.5 are in the session-5 handoff and still apply.
4. Proposed rows for the orchestrator:
   - prio 2: a search timeout returns 500. It should degrade, for example to empty results or a 503 with Retry-After.
   - vukem: the S3 disk exposes no ETag, so `storage.HandlerWith` cannot answer 304.

## Pitfalls

- `qa/schemathesis/ci.sh` and the `--network host` curl loop do not work on Docker Desktop. Use scratchpad `st.sh`: the app is published on :3114, then `run.sh` and k6 run against it. Kill a stuck ci.sh by its pid.
- osv-scanner in a container needs `-e GOTOOLCHAIN=auto`; without it, Go 1.27.1 < 1.27.3 and call analysis is skipped, which gives a false GO-2026-5932.
- Running Go integration, vitest and e2e in parallel causes vitest worker start timeouts and search timeouts. Rerun those on a quiet host.
- A `git commit -- <dir>` pathspec takes every modified tracked file under that dir. Name the files.
- Scratchpad: `rg.sh` (parallel gate), `lh.sh` + `lhcmp.js` (Lighthouse 12.6.1 against the lighthouserc budgets, not yet run), `st.sh`, `setrow.py`.
- Stack esp is stopped (`stop app`). DB hexascore_qa_esp still exists; drop it when it is no longer needed. Container hx-esp-pg has been removed.
