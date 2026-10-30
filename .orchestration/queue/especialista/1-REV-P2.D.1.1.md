# Review P2.D.1.1

Re-run:

- `go test ./tests -count=1 -run 'TestSecurity_A02_HexaHeadersWeb|TestSecurity_A02_HexaHeadersAPI|TestSecurity_A02_HexaCSPHasStrictDynamic'` → PASS
- `POST /csp-reports` is on the router, in `SessionBypass` and `CSRFBypass`, body limit 16 KiB, 30/min
- `/api/v1` CSP is the `default-src 'none'` policy, not the web nonce policy

Staging curls stay deferred. Report: `.orchestration/reports/P2.D.1.1.md` (PRONTO_PARA_MERGE).
