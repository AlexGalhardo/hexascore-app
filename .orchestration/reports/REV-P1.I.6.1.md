# REV-P1.I.6.1 (especialista, rodada 0)
VERDICT: CORRIGIDO_E_APROVADO (deliverable stays parcial by design: 52 rows open/TBD, P1.I.7.1 stays blocked)
Final branch: agent/especialista/P1.I.6.1-fix (da6c407 on top of 4b8e6ec)
Re-run: inventory | wc -l -> 118 (34 route/33 action/51 e2e); rows in tables -> 118; rg done|deferred -> 66; markdownlint-cli2@0.24.0 -> 0 issues.
Checks: all 51 e2e titles and 118 inventory lines present in the tables; every Test* name cited (incl. billing notes) exists as `func TestX(` in the branch (0 missing).
Spot-checks: TestHome(/), TestErrorPage(/no-such-page), TestUnverifiedUsersSeeTheNotice(/verify-email), TestProfilePassword, TestAdminPanelAccess(/admin), TestLegacyCheckoutRedirect(/profile/api/checkout->/billing), TestTimeToBeat, TestStatus: hit the cited routes. A25 deleteList: route exists, no test -> open correct. A33: TestTimeToBeat deletes only after the row was already cleared (idempotent 204), not a real delete -> open correct. R18 /docs/reference: no route -> TBD correct.
TBD rows: billing (R23, A16, A17, E30-E33) correctly wait on Q-H31; R18/E17/E18/E23 need Scalar hosting decision; A09 setListHidden only exists as resolve outcome list_hidden -> owner/senior call. No new task created: all are already tracked (Q-H31, P1.I.2.x); R18 + A09 belong in the release-blocker list for the orchestrator.
Fixes: scripts/parity-inventory.sh had mode 100644 (doc's `scripts/parity-inventory.sh | wc -l` fails on Unix) -> 100755; default legacy path ../../legacy only worked from wt/<role>; now tries ../legacy (main checkout) first. Verified both layouts -> 118.
Not run: bun lint/typecheck/test (no TS/Go changed).
