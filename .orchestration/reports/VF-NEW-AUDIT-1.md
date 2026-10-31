# VF-NEW-AUDIT-1 full-page error render in Inertia apps
status: concluida
branch: fw: fix/vf-inertia-html-error-page @ 742b3e51 (not pushed, no PR; worktree wt/fw-VF-NEW-AUDIT-1)
files: inertia/errors.go, inertia/inertia.go (renderRoot extracted from RenderStatus), inertia/errors_test.go, site docs (frontend/inertia.mdx, basics/error-handling.mdx), site/content/last-updated.json
acceptance:
- `go test ./inertia/` → ok (TestErrorFullPages, TestSecurity_A10_InertiaFullErrorPageHidesInternals; existing "full loads keep the HTML status page" test unchanged = default off)
- `task ci` → EXIT=0, no FAIL
validation: golangci-lint ok, vet ok, tests ok
changelog: Added - inertia.Config.ErrorFullPages: renders ErrorComponent in the root view for full page loads (also unmatched routes), so service-raised errors keep the app shell.
roadmap: no box
risks / follow-ups: opt-in, default false, so no behaviour change. JSON clients and htmx unchanged (handled earlier in the error handler). If the root view render fails after writing starts the fallback could double-write; same exposure as RenderStatus. `task api:check` not run (additive field, expected clean; specialist runs it). docs/frontend.md not edited. HexaScore adoption: set ErrorFullPages: true and drop pageError after the release.
vukem-feedback status line: VF-NEW-AUDIT-1 -> fixed in vukem branch fix/vf-inertia-html-error-page (Config.ErrorFullPages), pending PR/release.

## Review (especialista, 2026-10-31)
verdict: APROVADO
re-run: `task ci` exit 0, `task api:check` rc 0 (additive `Config.ErrorFullPages`); `go test ./inertia/` ok after merging origin/main
checked: opt-in, default false (existing full-load test unchanged); JSON clients keep problem+json; `Cache-Control: no-store` on the rendered error; A10 test proves no internals leak; a render failure logs and falls back to the framework page; `renderRoot` is a pure extraction of RenderStatus
note: REV-VF-NEW-AUDIT-3 touches the same lines of `inertia/inertia.go`; merge it after this PR and move its shell into `renderRoot`
fixes: branch merged origin/main and last-updated.json regenerated (432b5867), no code change
PR: #107
merged: PR #107 33/33 green, squash 8fbe0f2a
