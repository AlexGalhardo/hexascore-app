# DOC-research-letterboxd: Letterboxd social model (follow, feed, reviews, likes, privacy)
status: parcial (the core items are not verified; support.letterboxd.com returned 403 for every article, most letterboxd.com pages too)
branch: none (report only)
files: `.orchestration/reports/DOC-research-letterboxd.md`
acceptance:
- researcher run, access date 2026-10-31: read 4 pages (below); 10 help-center article IDs and 7 letterboxd.com pages → HTTP 403
- Verified facts (quoted, with source):
  - `letterboxd.com/about/faq/`: spoiler flag "Contains spoilers"; moderated text "A moderator has locked spoilers for this review"; "Like" checkbox; Close Friends ratings "don't contribute to stats"; account default "Privacy mode" in Settings
  - `letterboxd.com/about/importing/`: "Draft entries don't appear in your watched films or diary for others"; diary date and watched status "remain publicly visible"
  - `letterboxd.com/legal/privacy-notice/`: an "account export option" in Settings; data is "either delete or anonymize[d]" on deletion, with retention for legal, tax or accounting duties
- Not verified: follow/approval model, activity feed contents and ordering, review length and Markdown, comments and threading, blocking, reporting, profile and review visibility, what happens to reviews and likes on deletion, export format.
validation: links were opened by the researcher, not by me; no claim here is from a third-party site
changelog: none (report only)
roadmap: no box
risks / follow-ups: the Phase 7 SDD cannot rely on this report for follows, feed, reviews or blocking. Needs a second pass by browser (claude-in-chrome or agent-browser) on `support.letterboxd.com` articles, or a human reading them. Until then, `docs/web-research.md` §11 stays missing and P7.A.1.1 / P7.B.1.1 refs to Letterboxd remain open (DOC-roadmap-p7-refs patch item 3 and 7).

## Review (especialista, 2026-10-31)
verdict: APROVADO as parcial (report only)
checked: the three verified facts are quoted with their page and access date; unverified items are listed as such, nothing is asserted beyond the sources; 403s are a site block, not a method error
follow-up: the core Phase 7 items (follow model, feed, reviews, blocking, export) stay open under owner question Q-A9; a browser pass or a human read is needed before SDD 14 §12 cites Letterboxd
