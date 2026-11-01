# FIX-nojs-auth-post No-JS login/signup POST
status: concluida
branch: agent/senior/FIX-nojs-auth-post @ 5388590
files: resources/js/pages/Auth/{Login,Signup}.tsx, resources/js/pages/Auth/NoJsForm.test.tsx, tests/pow_hardening_test.go (challenge.go and app/challenge untouched)
finding: the premise was only half right. Browsers send Sec-Fetch-Site/Origin on a form post, so CSRF passes without a token, and PoW is demanded only past the ADR 0016 thresholds. The real defect was in the SSR markup: Field inputs had no `name` (name/email/identifier posted nothing), the signup `form_token` was not a hidden input (a no-JS signup hit the honeypot branch: fake success, no user), and `remember` had no name/value.
acceptance:
- `go test -tags integration ./tests -run 'TestSecurity_A07|TestSecurity_A10'` -> ok (39 s); new TestSecurity_A07_NoJSFormPost: urlencoded post without X-Challenge/Inertia signs up under the threshold (3 users stored), 428 + no user past it (fail closed), cross-site post 403
- `bunx vitest run resources/js/pages/Auth` -> 3 passed; NoJsForm.test fails (2) on the old markup, passes on the new
validation: biome ci resources/js/pages/Auth ok, typecheck ok, golangci-lint ./tests (see note), browser not run (markup-only, covered by jsdom test)
changelog: Fixed: login and signup forms now post their named fields, the signed form token and `remember` without JavaScript.
roadmap: no box (closes the no-JS half of P2.H.2.2)
risks / follow-ups:
- Past the PoW threshold a no-JS user still gets the raw 428 problem JSON (fail closed, per SDD 13 §7.2 the page offers Google sign-in). A friendly HTML page for that case would be a UX task; no ADR change needed, SDD/ADR untouched.
- `bun run lint` fails on main for qa/e2e innerText warnings (not this diff).

## Review (especialista, 2026-10-31)
verdict: APROVADO (merge queued: main frozen for v0.12.0)
re-run: detached worktree at 5388590: `bunx vitest run resources/js/pages/Auth` -> 2 files / 3 tests; golangci-lint --build-tags integration ./tests 0 issues; `go test -tags integration ./tests -run 'TestSecurity_A07|TestSecurity_A10'` -> ok, TestSecurity_A07_NoJSFormPost PASS
checked: A07, SDD 13 §7.2: every SSR input named (identifier, password, remember=true, name, email, password, password_confirmation, website honeypot, hidden form_token); Field spreads `name` onto the input; `challenge.go` and app/challenge untouched (diff is 4 files); past the PoW threshold a no-JS post stays 428 with no user (fail closed); cross-site post 403
notes: the raw 428 JSON for no-JS users past the threshold is a UX follow-up, not a security gap

PRONTO_PARA_MERGE agent/senior/FIX-nojs-auth-post (merged dc99c4d; gate: go build ok, golangci tests 0 issues, TestSecurity_A07_NoJSFormPost pass, biome/tsc ok, vitest 23 files / 79 tests)
