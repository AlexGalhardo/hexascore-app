# DOC-link-check Relative links and anchors in docs and root guides
status: concluida (report only; no edits)
branch: none (main @ 246ced8 at read time)
files: .orchestration/reports/DOC-link-check.md
acceptance:
- command: `bun <scratchpad>/links.js` (walks `docs/**/*.md`, `ROADMAP.md`, `AGENTS.md`, `PROMPT-WORKERS.md`; resolves each `[text](target)` relative to its file; checks the file exists and, for `#anchor`, that the target file has the heading slug or `<a id>`; fenced code blocks skipped; http(s)/mailto skipped)
  - output: `files: 61 relative links checked: 73 broken: 2`
  - both hits are false positives inside code spans, not links: `docs/sdd/02-scoring.md:239` (`QueryOf[Item](tx)`) and `docs/spike-report.md:113` (`orm.JSON[T]` inside backticks). Real broken links: 0.
- reference-style links (`[x]: path`): `rg -n "^\[[^\]]+\]:\s*\S+" docs ROADMAP.md AGENTS.md PROMPT-WORKERS.md` → none
- raw HTML `href="..."` relative links: only `ROADMAP.md:2238`, inside a code span (curl example), not a link
validation: lint n/a, typecheck n/a, tests n/a, browser n/a
changelog: none
roadmap: no box
risks / follow-ups: the slug rule is GitHub-like (lowercase, punctuation stripped, spaces to `-`, duplicates suffixed `-1`); a heading with unusual unicode could differ from GitHub's rendering, so an anchor flagged in the future should be checked in a browser. Links to `https://vukem.dev/docs` were not fetched (not in scope of a local check).

## Review (especialista, 2026-10-31)
verdict: APROVADO
re-run: sampled both false positives (`docs/sdd/02-scoring.md:239`, `docs/spike-report.md:113`): code spans, not links; every relative link target in `AGENTS.md` exists (`rg -o '\]\(...\)'` + `test -e`); markdownlint: `.orchestration/**` is excluded by the repo config
checked: facts sampled, scope (report only, no branch), nothing to merge
