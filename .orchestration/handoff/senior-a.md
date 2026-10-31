# senior-a handoff

In review: P2.H.3.5 (8e8a4f9), P4.A.3.1b (2923481), P4.A.3.1c (4d346a9), FIX-nojs-auth-post (5388590), DOC-sdd-flag-names (ad46202). Merged: P4.A.3.1 part 1, P3.A.4.1, FIX-sdk-python-regen, P2.H.2.1, FIX-covers-base-state. Community chain: 1b -> 1c, merge in order.
Blocked: FIX-sdk-ts-generator (Q-A8), vukem push (Q-A7), Q-A11.
Next: FIX-community-problem-codes, P4.A.3.1d and P4.A.3.2 once P4.A.3.1c is merged; P4.B.1.1 only after P4.A.4.1.
Pitfalls: heredocs with quotes in Bash abort silently, use Write; python text-mode writes CRLF on Windows (use newline=''), golangci gci then fails; export TEST_DATABASE_URL in the same call as go test; foreground commands cap near 120 s, use run_in_background; go test dirties go.sum, `git checkout go.sum`; QA builds use `docker compose --env-file C:\dev\hexascore-app\hexascore-vukem\.env` (never copy .env); flag env is FLAG_COMMUNITY_SUBMISSIONS; never `docker compose --profile app down`; k6 results swing 5x with host load (run via grafana/k6:2.3.0 with BASE_URL=http://host.docker.internal:<port>); QA db hexascore_qa_sa dropped, stack `hexascore-qa-sa` stopped.
