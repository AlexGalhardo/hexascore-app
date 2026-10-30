# REV-P3.A.6.1

verdict: APROVADO

final branch: agent/implementer/P3.A.6.1 (968098a)

checked: v1.Deprecate returns ErrSunsetTooSoon when sunset.Sub(since) < 183*24h and otherwise calls vukem api.Deprecate. Exactly 183 days is accepted. The test-only handler returns Deprecation @unix, Sunset as an HTTP-date, and Link rel=deprecation. No production v1 route calls Deprecate. The changelog entry and the owner mail stay manual runbook steps. No new dependency. P3.A.6.2 was not started.

acceptance on agent/implementer/P3.A.6.1: go test ./tests -count=1 -run TestAPIDeprecationHeaders -v -> PASS (0.00s), exit 0.
