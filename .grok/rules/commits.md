# Commit rules

- Conventional Commits 1.0, English: `type(scope): description`.
- Types: feat, fix, perf, refactor, docs, test, build, ci, chore, style, revert.
- Scope = area, lowercase: `auth`, `catalog`, `ratings`, `lists`, `billing`, `api`, `admin`, `ui`, `infra`, `deps`, `repo`, `docs`.
- Description: imperative, lowercase start, no trailing period, header ≤ 100 chars.
- Breaking change: `!` after type/scope plus a `BREAKING CHANGE:` footer.
- Reference the ROADMAP.md task in the body: `Refs: P1.A.1.2`.
- One logical change per commit. Branches are `agent/<role>/<task-id>` and merge straight into `main` (no `dev` branch).
- The POSIX hook `.githooks/commit-msg` enforces the header (enabled by `git config core.hooksPath .githooks`, done by the setup scripts). CI runs commitlint (D-37) on PR commits and the PR title.
- Style matches vukem (<https://github.com/AlexGalhardo/vukem/commits/main/>): short lowercase imperative subject that says what changed for a user or operator (`fix(auth): block sign-in for seed users`), not the task id or the process (`chore(orchestration): mark X ready` is board noise and stays out of release notes).
- Never write a subject that commitlint reads as a footer: a body line starting with `Word:` (for example `Central:`) or a `Refs:` line over 100 characters fails `footer-max-line-length`; split long `Refs:` into several lines.
- Merge commits: the hook rejects git's default `Merge branch ...`; use `chore(scope): merge <branch> into main`.
- Release notes are generated from these subjects (git-cliff and `scripts/release/history-notes.ts`); docs, test, ci, build, chore and style commits are left out of the notes, so a user-visible change must be `feat`, `fix`, `perf` or `refactor`.
