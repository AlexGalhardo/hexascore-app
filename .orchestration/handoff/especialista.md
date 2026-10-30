# Handoff especialista (2026-10-29, session 13)
STATUS: working. Done this session (PRONTO_PARA_MERGE self): P2.D.2.2 (agent/especialista/P2.D.2.2), P2.D.2.3 (agent/especialista/P2.D.2.3). Next: 4-P2.F.4.1, 5-P2.G.1.1, 6-P2.E.1.2, 7-P2.C.3.3. Skip 9-*.
Pitfalls added: biome format on TS; signup now needs form_token (tests: signUpForm(app,...)/signupToken(app)); python edits via heredoc break if text has unbalanced quotes (use Write tool); vukem MethodOverride swallows 413 for urlencoded (VF-NEW-P2.D.2.2).
(older pitfalls: golangci-lint --allow-parallel-runners + golangci-lint fmt; pgtest Fatal without TEST_DATABASE_URL; vitest under node; rm qa/e2e/test-results before markdownlint.)
