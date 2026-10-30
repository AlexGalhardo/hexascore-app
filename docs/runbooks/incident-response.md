# Incident response

Operator procedure for SDD 13 §7.9 (FR-13-7, NFR-13-6). The owner is the sole operator and decides severity and whether to notify. The ANPD deadline is per ANPD rule, verify before use (SDD 13 issue 5). Do not treat the 72 h figure below as that legal deadline.

## Severity

| Severity | Examples | Response |
| --- | --- | --- |
| SEV1 | data breach, active exploitation, leaked `APP_KEY` or Stripe key | start within 1 hour; maintenance mode if needed; user notice target 72 h |
| SEV2 | exploitable flaw, no sign of use | fix within 72 h |
| SEV3 | hardening gap, Medium ZAP finding | next release |

## Steps

1. Contain. Revoke the exposed credential, rotate keys (SDD 13 §7.4 and SDD 12 R9), ban the account if one is involved, and run `vukem down` when the app must stop taking traffic.
2. Export Railway logs before they expire. `railway logs` only keeps a short window: save the export (service, environment, UTC range) with the incident notes in the same hour you open the incident. Do not wait for the review.
3. Assess which data and which people are affected. Record the tables, the time range and whether passwords, tokens, e-mail addresses or payment identifiers were included.
4. Notify. When there is relevant risk, notify ANPD per ANPD rule, verify before use, and the affected users with the templates in `docs/runbooks/incident-notice.en.md`, `incident-notice.pt-BR.md` and `incident-notice.es.md`.
5. Fix the cause, confirm the security tests that cover it, then write the review.

## Review

Within 5 working days, copy `docs/incidents/TEMPLATE.md` to `docs/incidents/YYYY-MM-DD-<slug>.md` and fill it in. The review states a UTC timeline, impact, root cause, detection gap, actions with owners, and any SDD change.

## Log export

Railway log retention is shorter than the review window. Step 2 is mandatory for SEV1 and SEV2, and for any SEV3 that depends on a request log. Store the export outside the Railway project (the incident file's directory is not a log archive). Name the file `docs/incidents/YYYY-MM-DD-<slug>-logs.txt` and do not commit secrets that the log still contains: redact tokens and `APP_KEY` before the file is added to git.
