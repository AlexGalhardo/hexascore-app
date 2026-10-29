# senior-b handoff
Done this session (all pushed, reviews queued): P1.E.1.1, P1.C.2.1b, P1.C.3.2. Earlier: P1.B.2.2 (merged), P1.B.4.1, P1.C.3.1.
Worktree wt/especialista is on agent/senior-b/P1.C.3.2. Next: `git fetch && git switch -c agent/senior-b/<ID> origin/main`.
Note: P1.C.3.2 branch predates P1.C.2.1b merge (login field email vs identifier, /onboarding redirect): whoever merges second must adjust tests/deletion_test.go.
Queue (senior): 2-P1.C.4.1, 2-P1.D.2.2, 2-P1.D.4.1, 2-P1.D.4.2, 3-P1.C.2.2, 3-P1.C.5.2, 3-P2.B.1.1-3 (check before claiming).
Pitfalls: NEVER write in $APP. Bash heredocs with quotes/backticks break: use the Write tool for Go files; python scripts from a file. vukem i18n catalogue rejects JSON arrays (use {"0":..}). golangci-lint --fix fixes golines/gci; do not run two lint runs in parallel. vukemtest Client has Delete(), no Request(). Test client non-Inertia gets 401/422 JSON. username CHECK is ^[a-z]{4,16}$. Heartbeat: heartbeat/senior-b.txt.
