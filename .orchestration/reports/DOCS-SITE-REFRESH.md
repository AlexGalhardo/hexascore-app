# DOCS-SITE-REFRESH report (session 13)

- Items 3-4: PR #88 (rules + `Site / Last updated` CI check), open.
- Item 1 (content audit) done: #89 security (overview OWASP rows for passkeys, impersonation, Cloudflare IP trust, Betterleaks/doctor, LOG_TRUNCATE_IP; 6 security events missing from building-blocks), #90 REST (JSON-RPC in overview, struct{}+Op.Status no-body, signed cursors), #91 basics (routing links Group.ErrorHandler). Audit method: per-feature grep of every item in the task list, inline `pkg.Ident` refs checked against declared identifiers (0 stale), env vars added since v0.33.0 all in configuration.mdx, nav validated by build. Every merged feature since v0.33.0 already had its page/section. Unmerged features (#65-67 Cloudflare, #71, #77, #82, #87) carry their own docs.
- Each audit PR refreshes last-updated.json for its changed pages only (from full history), so parallel PRs do not conflict; satisfies #88's check.
- Gates (each PR): check-mdx 0 errors, `next build --webpack` ok, snippetcheck 578 blocks 0 problems. No .md file changed, so markdownlint had nothing to lint.
- Item 2 (full `npm run last-updated` refresh) PENDING: must branch from main after #88-#91 merge, else it conflicts with them. Orchestrator: send "refresh last-updated" when they land.
