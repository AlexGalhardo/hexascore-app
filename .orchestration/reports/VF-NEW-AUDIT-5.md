# VF-NEW-AUDIT-5 custom admin pages inside the panel layout
status: concluida
branch: fw: fix/vf-admin-custom-pages @ 727284f0 (not pushed, no PR; worktree wt/fw-VF-NEW-AUDIT-5)
files: admin/admin.go, admin/templates/page.html, admin/pages_test.go, admin/helpers_test.go, docs/admin.md, site/content/docs/digging-deeper/admin.mdx, site/content/last-updated.json
acceptance:
- `go test ./admin/` -> ok (TestAdminCustomPage, TestAdminCustomPageIsGuarded, TestAdminAddPageErrors, TestTemplatesListedInDocs)
- `task ci` -> EXIT=0 (after retries for "parallel golangci-lint")
- `task api:check` -> clean (only "no stable v1 tag" notices)
changelog: Added - admin.Panel.AddPage(slug, label, handler): custom GET page rendered in the panel layout with the same navigation, title and guard as resources; new overridable template page.html.
roadmap: no box
risks / follow-ups: GET only (POST via router + Panel.Guard, documented). Handler returns trusted template.HTML. Slugs share the resource namespace. No per-page ability beyond the panel ability. Additive API.
HexaScore adoption: replace the hand-made /admin/ingestion page with AddPage and drop the duplicated nav (P2.H.3.3).
vukem-feedback status line: VF-NEW-AUDIT-5 -> fixed in vukem branch fix/vf-admin-custom-pages (Panel.AddPage), pending PR/release.
