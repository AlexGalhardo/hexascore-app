# VF-NEW-AUDIT-3 shell for client-rendered Inertia pages
status: concluida
branch: fw: fix/vf-inertia-csr-shell @ bf0036c7 (not pushed, no PR; worktree wt/fw-VF-NEW-AUDIT-3)
files: inertia/inertia.go, inertia/ssr_test.go, site/content/docs/frontend/inertia.mdx, docs/frontend.md, site/content/last-updated.json
acceptance:
- `go test ./inertia/` -> ok (TestClientShell, TestClientShellSkippedWhenServerRendered)
- `task ci` -> EXIT=0, no FAIL
- `task api:check` -> only "no stable v1 tag" notices, no breaking change
changelog: Added - inertia.Config.ClientShell: head markup (for example the page title from props) and a placeholder inside the mount element for client-rendered pages (DisableSSR, Except, no SSR, SSR fallback), so a CSR page is never blank.
roadmap: no box
risks / follow-ups: additive, default nil = no change. The client must use createApp for these pages (placeholder is not server-rendered output; documented). Shell HTML is trusted, caller escapes. Not applied to Inertia JSON responses.
HexaScore adoption: set ClientShell with title from props and the skeleton (P2.H.2.1), drop the app-side workaround.
vukem-feedback status line: VF-NEW-AUDIT-3 -> fixed in vukem branch fix/vf-inertia-csr-shell (Config.ClientShell), pending PR/release.

## Review (especialista, 2026-10-31)
verdict: CORRIGIDO_E_APROVADO
re-run: `task ci` → CI_EXIT=0; `task api:check` → rc 0 (additive; "no stable v1 tag" notices only); `go test ./inertia/` ok
checked: scope (inertia only), API additive (default nil = old output), trusted-HTML contract documented, tests cover CSR, SSR skip and Inertia JSON
fixes: 1. merged origin/main (#107 `renderRoot` extraction): shell moved into `renderRoot`, `st.in.cfg` → `in.cfg`; error pages via `RenderStatus` get the shell too 2. docs: the new "A shell for client-rendered pages" section had been inserted mid-list, splitting the SSR helper bullets; moved after the list 3. `last-updated.json` regenerated after the merge
PR: AlexGalhardo/vukem#109 (branch fix/vf-inertia-csr-shell)
