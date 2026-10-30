# Orchestration

Three agents, any runtime. The briefs do not depend on Claude Code, a vendor skill tool, or a named model.

| Role | Queue | Does | Does not |
| --- | --- | --- | --- |
| `framework` | `queue/framework/` | vukem only: merge green PRs, read GitHub Actions failures, open issues | app features, tags, `BREAKING CHANGE` |
| `implementer` | `queue/implementer/` | app code and tests for one microtask | merge to `main`, review its own work |
| `reviewer` | `queue/reviewer/` | re-run acceptance, fix the branch, merge to local `main` | start a feature the implementer has not delivered |

Run all three at once. They must not edit the same files. Each agent reads `briefs/common.md`, then `briefs/<role>.md`, then claims the lowest-numbered file in its queue.

Old names, kept only so old reports stay readable: `senior` and `estagiario` are `implementer`; `especialista` is `reviewer`.
