# HexaScore: guide for AI agents

HexaScore is a catalogue, rating and recommendation platform for games, films, series and books. This repository is its Go rewrite on vukem; the legacy Next.js app is on branch `legacy`.

## Start here

1. Read [ROADMAP.md](ROADMAP.md), then [docs/ROADMAP-V1.md](docs/ROADMAP-V1.md) and [docs/BACKLOG.md](docs/BACKLOG.md) for milestones and task order.
2. Check the latest [review snapshot](docs/review/) and relevant [design documents](docs/sdd/), [ADRs](docs/adrs/) and [owner decisions](docs/brainstorming.md) before changing behavior.
3. Read every matching rule in [.agents/rules/](.agents/rules/) before editing; read the applicable skill in [.agents/skills/](.agents/skills/) when its description fits. Rules are not loaded automatically.
4. For Railway operations, follow [.agents/skills/use-railway/SKILL.md](.agents/skills/use-railway/SKILL.md) and [the Railway runbooks](docs/runbooks/). Never deploy unless the owner explicitly asks. A push to `main` deploys production.

## Working rules

- Keep repository code, comments and docs in English; chat with the owner in pt-BR. Put UI text in the `en`, `pt-BR` and `es` catalogues.
- Treat [brainstorming decisions](docs/brainstorming.md) and [ADRs](docs/adrs/) as final: update the decision, then the SDD, then code when a change is approved.
- Prefer native vukem features. Do not add Go/npm dependencies without owner approval and an ADR. Pin exact versions; never use `latest` or prereleases.
- Use official [vukem documentation](https://vukem.dev/docs) for framework facts. Do not invent APIs; record undocumented gaps in the local framework feedback process.
- Make surgical changes. Do not weaken security defaults. Update owning docs and `CHANGELOG.md` when behavior, commands or conventions change.
- Follow the acceptance commands in [ROADMAP.md](ROADMAP.md). Run focused checks for the task; the full release gate is in [the local CI script](scripts/ci/local.sh).
- Use Conventional Commits per [.agents/rules/commits.md](.agents/rules/commits.md). Never rewrite history or force-push. Workers do not push; only the orchestrator may push `main`, after the required gate.
- Never commit `go.work` or a local `replace` directive. Keep `legacy` immutable and never delete it.
- Before production data changes, get owner approval and follow the relevant [runbook](docs/runbooks/).

## Project map

- [docs/sdd/](docs/sdd/): product and domain specifications; [docs/sdd/open-issues.md](docs/sdd/open-issues.md): unresolved design points.
- [docs/rfcs/](docs/rfcs/): catalogue worker, trending and offers proposals; [docs/data/SOURCES.md](docs/data/SOURCES.md): source access, terms and decisions.
- [docs/runbooks/](docs/runbooks/): operations, releases, incidents and recovery; [docs/setups/](docs/setups/): local and self-hosted setup.
- [docs/audit/](docs/audit/): project map, reviews and cleanup ledger; [docs/review/](docs/review/): dated state snapshots.
- `.agents/rules/`: path-scoped coding rules; `.agents/skills/`: project skills. `CLAUDE.md` points here.
