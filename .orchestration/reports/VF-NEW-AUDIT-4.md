# VF-NEW-AUDIT-4 data-hydrating hook on the Inertia scaffold
status: concluida
branch: fw: fix/vf-inertia-hydrating-hook @ b3776f6c (not pushed, no PR; worktree wt/fw-VF-NEW-AUDIT-4)
files: internal/cli/scaffolds/inertia-{react,vue,svelte}/resources/views/app.html and resources/js/app.(ts|tsx), internal/cli/inertia_test.go, site/content/docs/frontend/inertia.mdx, last-updated.json
acceptance:
- `go test -run TestInertiaScaffoldsExposeHydratingHook ./internal/cli` -> ok (all three scaffolds set and clear data-hydrating)
- `task ci` -> ci=0; `task api:check` -> api=0
validation: lint ok, tests ok
changelog: see commit message (CHANGELOG is generated per release)
roadmap: no box
risks / follow-ups: Scaffold-level hook, not a core API: existing apps and custom root views must add `data-hydrating` themselves (documented). It is removed in `.finally` of createInertiaApp, so a failed boot also clears it. TS typecheck/build of generated projects not run here (no node_modules in the worktree; createInertiaApp returns Promise<void> per @inertiajs/react types). Specialist: push, PR, merge on green CI. last-updated.json already refreshed in the commit (committer date preserved on amend).
vukem-feedback status line: VF-NEW-AUDIT-4 -> fixed in vukem branch fix/vf-inertia-hydrating-hook (scaffold data-hydrating hook + docs), pending PR/release.
PR: vukem #116 open (pushed, awaiting GitHub CI)

## Review (especialista, 2026-10-31)
verdict: APROVADO
re-run: `go test -count=1 -run TestInertia ./internal/cli/` ok
checked: all three starters (react, vue, svelte) set and clear the attribute; removal in `.finally` so a failed mount does not leave the page styled as pending; CSS-only hook keeps the CSP strict; docs section
fixes: branch merged origin/main, last-updated.json regenerated; no code change
PR: #116
PR: #116 squash c7b5efe4 (33/33)
