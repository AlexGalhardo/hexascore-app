# REV-P1.G.3.2b
VERDICT: APROVADO (items 1, 2, 4 and the clear half of 3). Final branch agent/senior/P1.G.3.2b.
Re-run: `bun run lint && bun run typecheck && bun run test` -> Checked 80 files, 13 files / 53 tests passed; `go test -count=1 -tags integration ./tests -run 'TestInertiaPages/(Library)'` -> ok. No Go changes (golangci not affected).
Checked: counts sum respects the other filter; see-all keeps status/sort and uses localized(); DELETE /items/{id}/edition web branch answers c.Back("/library") (Inertia-safe), CSRF via router; copy in en/pt-BR/es; buttons/links keyboard reachable.
Open (orchestrator decision, not a defect): choose-edition stays bloqueada: no session-readable editions source (only bearer /api/v1/books/{id}/editions). Options: Library prop with editions per book row, or a session JSON route. Needs SDD 14 §7.2 decision.
Minor, not fixed: "Clear edition" buttons share one accessible name per row (context is the preceding "Edition: X" text).
