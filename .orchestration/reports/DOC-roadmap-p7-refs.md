# DOC-roadmap-p7-refs: proposed patch for Phase 7 references
status: concluida (report only; ROADMAP.md not edited)
branch: none (no repository file changed)
files: `.orchestration/reports/DOC-roadmap-p7-refs.md`
acceptance:
- `grep -n "SDD 14 §8" ROADMAP.md` → 7 hits (86, 3786, 3796, 3805, 3821, 3830, 3854); §8 is "Edge cases" in `docs/sdd/14-user-library.md` (line 210)
- `grep -n "web-research §4" ROADMAP.md` → 2 hits (3788, 3821); `docs/web-research.md` §4 is Metacritic (line 245), Goodreads §6 (378), Skoob §7 (444); no Letterboxd section
- SDD 13 table (`docs/sdd/13-security.md` lines 131-139): A03 = supply chain, A04 = crypto, A05 = injection, A06 = limits
validation: markdownlint not applicable (`.orchestration/**` is ignored by the lint config; `bunx markdownlint-cli2@0.24.0` on `qa/e2e/test-results` reports pre-existing errors, out of scope)
changelog: none (report only)
roadmap: no box
risks / follow-ups: assumes SDD 14 §12 "Social layer" (AUDIT-social-sdd §2); if SDD 15 is chosen, replace `§12`. Letterboxd refs wait for DOC-research-letterboxd.
Proposed patch (old → new, in `ROADMAP.md`):
1. L86 table: `SDD 14 (new §8)` → `SDD 14 (new §12)`
2. L3786 P7.A.1.1: `write SDD 14 §8 "Follows and feed"` → `write SDD 14 §12 "Social layer" (follows and feed subsection)`
3. L3788 P7.A.1.1 Refs: `web-research §4 (Letterboxd, Goodreads, Skoob feeds)` → `web-research §6 (Goodreads), §7 (Skoob); Letterboxd: DOC-research-letterboxd report`
4. L3796 P7.A.1.2 Refs: `SDD 14 §8; SDD 13 A01, A04 (rate limit 60/h)` → `SDD 14 §12; SDD 13 A01, A06 (rate limit 60/h)`
5. L3805 P7.A.1.3 Refs: `SDD 14 §8; SDD 08 §6 (activity primitives…)` → `SDD 14 §12; SDD 08 §7.1 (primitives…)`
6. L3806 P7.A.1.3 Depends on: `P7.A.1.2` → `P7.A.1.2, P6.C.1.1` (uses the P6.C.1 primitives; P7.A depends only on P2.G today)
7. L3821 P7.B.1.1 Refs: `SDD 14 §8; SDD 13 A03 (…); …; web-research §4.2 (Letterboxd review model)` → `SDD 14 §12; SDD 13 A05 (…); …; Letterboxd review model: DOC-research-letterboxd report`
8. L3825 P7.B.1.1 acceptance: `TestSecurity_A03_ReviewMarkdown` → `TestSecurity_A05_ReviewMarkdown`
9. L3830 P7.B.1.2 Refs: `SDD 14 §8; SDD 13 A04; SDD 04 §6 (notifications…)` → `SDD 14 §12; SDD 13 A06 (rate limit); SDD 14 §7.5 (notifications)`
10. L3854 P7.C.1.1 Refs: `SDD 08 §6; SDD 14 §8` → `SDD 08 §6; SDD 14 §12`

Notes from the check (not in AUDIT-social-sdd): `docs/sdd/04-community-edits.md` §6 is "Contracts", not notifications; notification rules are SDD 14 §6.5 and §7.5 (SDD 04 §5 line 157 already cites them). SDD 14 §5 "Deferred shape" (line 123) matches the audit's defect 4.
