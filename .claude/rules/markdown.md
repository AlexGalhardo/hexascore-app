---
paths:
  - "**/*.md"
---

# Markdown rules

- Must pass `bunx markdownlint-cli2@0.24.0 "**/*.md"` with the repo config `.markdownlint.jsonc`.
- ATX headings (`#`), dash lists, fenced code blocks with a language, blank lines around blocks.
- Indent with 2 spaces; hard tabs are allowed only inside `go` / `makefile` fences.
- English, concise, link to the owning doc instead of duplicating content.
- `AGENTS.md` stays short and links to `ROADMAP.md` and `docs/`.
