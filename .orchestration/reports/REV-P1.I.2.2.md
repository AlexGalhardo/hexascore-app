# REV-P1.I.2.2 (rodada 0)
STATUS: APROVADO
Branch: agent/senior/P1.I.2.2 @ c9e96ab (no changes by especialista).
Re-run (isolated stack hx-especialista, fresh DB, port 3200):
- playwright catalog/item/search/ratings/time-to-beat -> "120 passed (56.3s)", exit 0 (fresh DB: no 429).
- bun run lint 0, typecheck 0, test "59 passed"; go test ./app/catalog ok; golangci-lint ./app/catalog "0 issues".
Review: ParseListQuery NUL/invalid-UTF-8 guard is the right place (single parser for all listings) with a table test; helpers are test-only; projectSlot avoids the per-item burst limiter across projects; non-ported legacy features justified (no UI yet / D-09).
Risk: signup retry loop could double-submit if the first click was slow but succeeded (toPass reclicks only while URL not reached; second POST would hit "email taken" and stay on page -> test fails loudly, not silently). Acceptable.
PRONTO_PARA_MERGE agent/senior/P1.I.2.2
