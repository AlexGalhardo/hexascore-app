# Handoff especialista (2026-10-29, session 3)
STATUS: HANDOFF (heavy context after the P1.H.2.1a implementation). Queue left: 1-REV-P1.F.2.1.md, 1-REV-P1.G.3.2.md (untouched).
Done this session:
- P1.H.2.1a concluida -> agent/especialista/P1.H.2.1a (pushed, PRONTO_PARA_MERGE self, needs orchestrator sampling). Report reports/P1.H.2.1a.md has the pattern for H.2.1b/c.
Pitfalls (new):
- Python text-mode writes on Windows emit CRLF: use newline='\n' or run gofmt -w afterwards.
- /api/v1 tests: free plan is 1 rps and the memory limiter ignores the fake clock; use newKeyRigWith(t, vukem.WithRateLimiters(... ratelimit.NewMemory(1000, window))).
- AssertMatchesOpenAPI needs OPENAPI_PREFIX=/api/v1 (operationId collision dashboard/admin.dashboard); newKeyRigWith sets it.
- Typed handler 304/301: c.NotModified + return errs.New(304) / errs.New(301).WithHeader("Location") (VF-NEW-P1.H.2.1a).
- Previous pitfalls (postgres container, golangci-lint parallel retry, merge-before-review) still apply.
Worktree C:\dev\hexascore-app\wt\especialista on branch agent/especialista/P1.H.2.1a, clean.
Skills: karpathy-guidelines/caveman not loaded this session (context economy).
