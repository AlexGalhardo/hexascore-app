# Commit rules

- Conventional Commits 1.0, English: `type(scope): description`.
- Types: feat, fix, perf, refactor, docs, test, build, ci, chore, style, revert.
- Scope = area, lowercase: `auth`, `catalog`, `ratings`, `lists`, `billing`, `api`, `admin`, `ui`, `infra`, `deps`, `repo`, `docs`.
- Description: imperative, lowercase start, no trailing period, header ≤ 100 chars.
- Breaking change: `!` after type/scope plus a `BREAKING CHANGE:` footer.
- Reference the ROADMAP.md task in the body: `Refs: P1.A.1.2`.
- One logical change per commit. Branches are `agent/<role>/<task-id>` and merge straight into `main` (no `dev` branch).
- The commit-msg hook is planned in P1.A.2.2; commitlint needs the owner's yes first (ROADMAP P1.A.2.2 notes). Until then, follow these rules by hand.
