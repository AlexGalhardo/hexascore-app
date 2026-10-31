# DOC-design-system: design system reference and concept-art prompts
status: concluida
branch: agent/senior/DOC-design-system @ 21cea07 (not pushed, worktree wt/senior-DOC-design-system)
files: docs/design-system.md, docs/prompts-concept-site-pages.md, docs/design-system/*.png (8 screenshots)
acceptance:
- `bunx markdownlint-cli2@0.24.0 docs/design-system.md docs/prompts-concept-site-pages.md` → 0 errors (run on the whole repo glob, 96 files)
- `ls resources/js/pages` → 33 Inertia components, all in the catalogue of design-system.md section 6; prompts file covers them in 35 entries (adds the global overlays and the admin panel, merges notice pages) x 3 styles = 105 prompts
- screenshots (production, read-only, logged out): home, /games, /games/diablo-v, /login at 1440x900; home, /games, /games/diablo-v, /login at 390x844 in `docs/design-system/`
validation: documentation only; no code, tokens or tests touched. Facts cited from `packages/tokens/`, `resources/js/`, `resources/css/app.css`, `resources/views/app.html`, `app/admin/views/`, `lighthouserc.cjs`, `qa/e2e/`, SDD 09. Production probes: `/docs/api` and `/games/diablo-v/changes` return 404.
changelog: none (docs only, no behaviour change)
roadmap: no box
styles: Editorial Press (`editorial`), Neon Noir Arcade (`noir`), Soft Paper Play (`paper`); images go to `docs/design-concepts/<page-id>-<style>[-m].png`
risks / follow-ups:
- Not verified: logged-in pages, dark mode and real focus colours were not captured (no login on production); authenticated page descriptions come from code only.
- Gaps listed (not fixed) in design-system.md section 11, notable: no fonts, many hard-coded indigo/zinc/red colours, score band tokens unused, no ScoreBadge/Skeleton/Button components, full-width auth forms, broken `/docs/api` link on API keys page, `Catalog/RecentChanges` missing, hard-coded English in `Auth/TwoFactor`.
- prompts file is about 400 kB (105 self-contained prompts); regenerate from the page data if styles change (generator was a scratch script, not committed).
