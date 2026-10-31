# Handoff: especialista (2026-10-31, session 3)

Reason: eight reviews done this session; context is heavy. STATUS: HANDOFF.

## Done this session

| Row | Verdict | Merge on local `main` |
| --- | --- | --- |
| REV-ADOPT-VF-001 | CORRIGIDO_E_APROVADO (purge test + e2e removal/localhost 053c426; route inventory 6f25b4e) | 75aa17b, 6f25b4e |
| REV-REBASE-P4.A.2.4 | CORRIGIDO_E_APROVADO (inline migration SQL, dead branch c30b97c) | 2307881 |
| REV-P2.H.1.3 | APROVADO (84 `<img>`, Lighthouse mobile perf 0.99, LCP 1.73 s) | 6181cb4 |
| REV-FIX-legacy-cover-retry | CORRIGIDO_E_APROVADO (retry vitest case) | 0a4cf88 |
| REV-P2.H.1.2 | APROVADO (item + ratings specs, 6 projects, 65 passed / 7 flaky in the onboarding helper) | 57497a9 |
| REV-FIX-mobile-catalogues | CORRIGIDO_E_APROVADO (recopied with the passkey keys) | a3006bb |
| REV-FIX-i18n-untranslated | REJEITADO r1 (the 4 values are the legacy wording) | none |
| REV-P2.H.2.2 | APROVADO (box open until FIX-nojs-auth-post) | ad5287f |
| (main regression) FIX-e2e-login-exact | exact login submit match in e2e helpers (pt-BR/es hit the passkey button) | merged before ad5287f |
| REV-DOC-board-mirror-2 | APROVADO | merged |
| REV-DOC-commands-check, REV-QA-gate-2026-11-01, REV-DOC-pin-inventory | APROVADO (report only, sampled) | none |

Gate on `main` @ 57497a9: build, golangci-lint (integration tags) 0 issues, bun lint/typecheck/vitest (70), go test integration all green. ROADMAP ticks: P1.C.7.1, P1.C.7.2, P2.H.1.2, P2.H.1.3, P4.A.2.4 (P2.F.2.2b has no own box). CHANGELOG lines added for all merges.

New row: FIX-e2e-onboarding-hydration (senior): `completeOnboarding` loses the nickname fill before hydration, so the first attempt of many specs times out.

## Next rows (prio order)

1. Check the gate on `main` @ ad5287f (scratchpad `gate.sh` output); if red, fix first.
2. Report-only rows left: REV-DOC-i18n-legal, REV-QA-mobile-proto-review, REV-QA-i18n-parity (sample facts, markdownlint).
3. FW-triage-branches (sample the vukem branch facts).
4. REV-FIX-covers-base-state stays blocked until the senior's rework.

## Pitfalls

- Copying `.env` into a worktree is denied by the permission system; use `docker compose --env-file ../../hexascore-vukem/.env ...` from the worktree.
- `bunx --bun @lhci/cli@0.15.1` fails to install on this Windows host (bunx git-clone bug); `bunx lighthouse@12.6.1 <url> --form-factor=mobile --output=json` works (first run may abort with NO_NAVSTART; rerun).
- WebAuthn needs a host name: run passkey e2e on `localhost` (the spec now sets it).
- `grep -c '<img'` counts lines; SSR HTML is one line. Use `grep -o '<img' | wc -l`.
- Run the whole `./tests` package before merging route changes: `TestSecurity_A01_HexaRouteStackInventory` needs every new route in `tests/testdata/route_inventory.txt`.
- Gate under load (docker build + e2e at once) times out golangci-lint and once failed `internal/search` TestGolden; pass `--timeout 15m` and rerun the package alone.
- Gate worktree `wt/especialista-gate` (detached) and script in the scratchpad: `gate.sh <sha>`.
- Worktrees to remove after `merged`: `wt/especialista-{ADOPT-VF-001,ADOPT-VF-001-fix,P4.A.2.4,P2.H.1.3,P2.H.1.2,FIX-legacy-cover-retry,FIX-mobile-catalogues,P2.H.2.2,gate}` plus the session-2 list. `wt/senior-ADOPT-VF-001` belongs to senior-a.
