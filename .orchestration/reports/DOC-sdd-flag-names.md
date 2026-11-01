# DOC-sdd-flag-names SDD 04 flag names
status: concluida
branch: agent/senior/DOC-sdd-flag-names @ ad46202
files: docs/sdd/04-community-edits.md
acceptance:
- `grep -rn "community\.submissions\|community\.trusted_moderators" docs` -> no match after the change (was 2 lines, SDD 04 §6)
- `bunx markdownlint-cli2@0.24.0 docs/sdd/04-community-edits.md` -> 0 errors
validation: markdownlint ok; docs only
changelog: none (docs only)
roadmap: no box
risks / follow-ups: flag names now `community_submissions` / `community_trusted_moderators` (a-z, 0-9, _ per https://vukem.dev/docs/digging-deeper/feature-flags), env `FLAG_COMMUNITY_SUBMISSIONS`, matching what P4.A.3.1b shipped.

## Review (especialista, 2026-10-31)
verdict: APROVADO (merge queued: main frozen for v0.12.0)
checked: vukem flag names match `^[a-z][a-z0-9_]{0,99}$` (vukem flags/flags.go nameRe; docs page cited in the diff); `app/community/page.go` declares `flags.Bool("community_submissions", false)`; both SDD 04 §6 lines changed, nothing else

PRONTO_PARA_MERGE agent/senior/DOC-sdd-flag-names (merged d21e0db; markdownlint ok, no stale flag names in docs)
