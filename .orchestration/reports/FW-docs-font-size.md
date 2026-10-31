# FW-docs-font-size docs site typography +25%
status: concluida
branch: fw: fix/docs-font-size @ 79c229f8 (not pushed, no PR; worktree wt/fw-FW-docs-font-size)
files: site/app/globals.css (html { font-size: 125% }, overflow-wrap on prose links), site/components/logo.tsx and site/components/search-dialog.tsx (text-[11px]/[15px] to the same rem values so they scale too). No content edits.
approach: every size on the site is rem-based (Tailwind text-*, prose rules, layout widths), so a 125% root scales body, code, tables, headings, sidebar, TOC and widths together and keeps the hierarchy and the line-height ratios. Breakpoints are unaffected (media queries ignore the root size). Measure: main column 705 px at 1440 px wide with sidebar and TOC, roughly 62 characters per line, so no max-width change was needed.
computed font-size before -> after (px, playwright, 1249/1440 wide):
- /docs/introduction (home redirects here): body p 14 -> 17.5; h1 34 -> 42.5; h2 22 -> 27.5; h3 17 -> 21.25; inline code 12.3 -> 15.4; sidebar and TOC links 14 -> 17.5
- /docs/rest-apis/typed-handlers (long, 21 code blocks, 4 tables): p 14 -> 17.5; h2 22 -> 27.5; pre code 13 -> 16.25; td/th 13 -> 16.25; inline code 12.3 -> 15.4; sidebar 14 -> 17.5
- every ratio is 1.25
mobile (390 px): found horizontal overflow (long pkg.go.dev URL in prose, 461 px scroll width); fixed with overflow-wrap: anywhere on .prose a; scrollWidth 378 <= 390 after.
screenshots: reports/img/fw-FW-docs-font-size-desktop-intro.png, ...-desktop-typed.png, ...-mobile-typed.png
acceptance:
- `npm run lint` ok, `npm run build` ok, `npm run check:mdx` 0 errors (12 existing warnings)
- `task ci` -> EXIT=0
changelog: Changed - docs site typography scaled by 25%.
roadmap: no box
risks / follow-ups: before screenshots not captured (before values are the computed sizes above). Fixed-px elements, if any are added later, will not scale. Tablet widths between 1024 and 1280 px have a narrower main column (sidebar grows with the root); main was 854 px at 1249 px.
