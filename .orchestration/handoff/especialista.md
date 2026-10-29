# Handoff especialista (2026-10-29, session 8)
STATUS: HANDOFF after 4 deliveries. Worktree C:\dev\hexascore-app\wt\especialista on agent/especialista/APP-PASSKEYS, clean.
Queue left: 1-REV-P1.H.4.1b, 2-P2.E.1.1, 3-P2.C.1.1, 9-APP-PASSKEYS-b (blocked on vukem VF-001 release; renamed to 9 on purpose).
Done this session:
- P2.A.3.2 -> self PRONTO (agent/especialista/P2.A.3.2): Node stops after HTTP drain, 4 ready checks 2 s, heartbeat, queue order test.
- REV-P2.A.1.1 -> CORRIGIDO_E_APROVADO (agent/especialista/P2.A.1.1-fix): script mode +x, curl -f, [NOT VERIFIED].
- REV-ADOPT-vukem-v0.33.0 -> CORRIGIDO_E_APROVADO (agent/especialista/ADOPT-vukem-v0.33.0-fix): /app/v1 had CORS * (SDD 13 forbids).
- APP-PASSKEYS -> parcial, docs self PRONTO (agent/especialista/APP-PASSKEYS): ADR 0017, SDD 03 §7.8, ROADMAP P1.C.7.
Pitfalls (new):
- Running the image locally needs dummy STRIPE_KEY/STRIPE_WEBHOOK_SECRET/STRIPE_PRICE_PRO/STRIPE_PRICE_BUSINESS + SESSION_DRIVER=database, DATABASE_URL host.docker.internal:5433/hexascore.
- Slow-request shutdown proof: `curl --limit-rate 1K --data-binary @10KB POST /login` then `docker stop -t 40`.
- Red check: stash only the production file, not the test (else the new assertion is not exercised).
- Claim order is by numeric prefix, then name: re-check the queue right before claiming (a REV arrived mid-session).
- Merge overlap: P2.A.3.2 and ADOPT both touch app/library/export.go (different hunks); CHANGELOG [Unreleased] touched by all four branches.
