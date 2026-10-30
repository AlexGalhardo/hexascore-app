# Task DOCS-SITE-REFRESH (priority 1, framework; owner request 2026-10-30)

Objective: bring the whole vukem docs site (`site/content/docs/**`) up to date with vukem `main`, and make "Last updated" reliable, with clear and explicit rules so it never goes stale again.

How "Last updated" works today (site/README.md "Last updated"): the date is the last git commit of the page's MDX file; Vercel clones shallowly, so pages git cannot date fall back to the committed `site/content/last-updated.json`, which is only refreshed by hand with `npm run last-updated`. Result: stale dates on the live site.

1. Content audit, one PR per area if large: every package or option merged since v0.33.0 is documented on the site, with nav entries, and no page describes removed or renamed APIs. At least: passkeys (`auth/webauthn`, `install:auth --passkeys`), React Email (`mail.React`, `make:mail --react`), `jsonrpc` + OpenRPC client, HTTPS/HTTP/2 + `contrib/http3`, `contrib/cloudflare` (IP, origin lock, Turnstile action, visitor headers, cache/purge, Access JWT), Betterleaks (`secrets:scan`, `vukem doctor`), `notify`, `api.PageOf/CursorOf/CollectionOf`, `totp.HashRecoveryCode`, `sitemap`, `httpclient` Throttle/Retry/Breaker, impersonation, group error handlers, OpenAPI `StripPrefix`/`RequiredOutputs`/`TypedOnly`, `LOG_TRUNCATE_IP`, typed handler status semantics, logging and releases pages. Cross-check each page's code samples against the current API (snippetcheck must stay at 0 problems).
2. Refresh `site/content/last-updated.json` from a full-history clone (`git fetch --unshallow` if needed, then `npm run last-updated`) as the final commit of the last PR, so every page touched gets today's date.
3. Explicit rules (write them, do not just mention them):
   - `AGENTS.md` and `CLAUDE.md` of vukem, and `.claude/rules/docs-content.md`: "Any PR that adds, changes or removes a page under `site/content/docs/` must refresh `site/content/last-updated.json` (`npm run last-updated`) in the same PR. The 'Last updated' date is part of the page contract; never edit it by hand."
   - `site/README.md` "Last updated": the same rule, plus the release checklist step.
   - `docs/maintenance.md` release checklist: refresh `last-updated.json` before tagging.
4. Enforcement: a CI check (script + job, stdlib/Node only, no new dependency) that fails a PR when an MDX file under `site/content/docs/` changed but its entry in `last-updated.json` is older than the PR's last commit touching it (or is missing). Add a test for the check script.
Gate: site build, check-mdx 0 errors, snippetcheck 0 problems, markdownlint 0, the new check green. Do not run `gh pr merge`; no release or tag. Report `$ORCH/reports/DOCS-SITE-REFRESH.md` with PR numbers.
