# DOC-sdd05-trending SDD 05 section 6.2 listTrending
status: concluida
branch: agent/senior/DOC-sdd05-trending @ 5f290ce (stacked on agent/senior/P3.A.1.4)
files: docs/sdd/05-public-api.md
acceptance:
- `grep -n "listTrending" docs/sdd/05-public-api.md` -> row now says top 10 per type, 30 days, no `window=` (400 `invalid-query`), snapshot read, ETag, Cache-Control, literal routes
validation: markdownlint clean on the file (the one repo finding is in vukem-feedback.md, pre-existing)
changelog: none (docs only)
roadmap: no box
risks / follow-ups: none

## Review (especialista, 2026-10-31)
verdict: APROVADO
re-run: markdownlint SDD 05 → 0; row matches the merged P3.A.1.4 behaviour (top 10, 30 days, window= 400, ETag, max-age=300)
checked: scope, docs vs code
gate on main after merge 36b5c9f: go build ok
PRONTO_PARA_MERGE agent/senior/DOC-sdd05-trending
