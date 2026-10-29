# CLAUDE.md

@AGENTS.md

Claude Code specifics:

- Every session starts by loading the skills `karpathy-guidelines`, `graphify` and `caveman` (see "Session start" in AGENTS.md; `graphify` is CLI-only in this environment, run `graphify update .` when `graphify-out/graph.json` is missing or stale, then `graphify query` before broad code searches).
- Code location: delegate to `cavecrew-investigator`; diff review: `cavecrew-reviewer`; external docs: `researcher`.
- Zen/caveman chat style applies to chat only; files stay in normal English prose.
- Before marking a roadmap microtask done, run its acceptance commands and paste the decisive output line in the session summary.
