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
