# senior-a handoff

In review: P4.A.3.1 (296a3db), P4.A.3.1b (2923481), P4.A.3.1c (4d346a9), P3.A.4.1 (fce88fc), FIX-sdk-python-regen (5b2ff0d), P2.H.2.1 (eb1c281), FIX-nojs-auth-post (5388590), DOC-sdd-flag-names (ad46202). Community chain: 3.1 -> 1b -> 1c, merge in order.
Blocked: FIX-sdk-ts-generator (Q-A8), vukem push (Q-A7), Q-A11.
Next: P2.H.3.5 once FIX-covers-base-state is merged (routes/web.go lock); FIX-community-problem-codes and P4.A.3.1d after 1c merges; P4.A.3.2 after 3.1 merges; P4.B.1.1 only after P4.A.4.1.
Pitfalls: heredocs with quotes in Bash abort silently, use Write; export TEST_DATABASE_URL in the same call as go test; foreground commands cap near 120 s, use run_in_background; go test dirties go.sum, `git checkout go.sum` before committing; QA builds use `docker compose --env-file C:\dev\hexascore-app\hexascore-vukem\.env` (never copy .env; two stray copies in wt/senior-P4.A.3.1b and 1c); flag env is FLAG_COMMUNITY_SUBMISSIONS; never `docker compose --profile app down`; `bun run lint` fails on main for qa/e2e innerText warnings (lint only changed dirs with `bunx biome ci <dir>`).
