# Hexascore — Guide for AI agents

Media database (games, movies, TV series, books) with 6-category 0–100 user scores and a paid public REST API.
Codename `hexascore`; final brand name is pending.

**Status: planning.** No application code yet. The single source of truth is [`PLAN.md`](PLAN.md): decisions, architecture, data model, API, milestones. Read it before any work and keep its checkboxes up to date.

## Rules

- Communicate with the owner in pt-BR; code, identifiers and commit messages in English.
- Conventional Commits (commitlint). Changes go through PR to `dev`, then `dev` → `main`.
- Only stable, pinned dependency versions (no `^`, no beta/rc/canary). Ask before adding a dependency.
- Bun is the package manager. Python code lives in `webscrappings/` and uses uv.
- Follow the Karpathy guidelines (think first, simplicity, surgical changes, verifiable goals), `ponytail` (YAGNI, minimal solution) and `graphify` (map the code before broad searches). Vendored skills live in `.claude/skills/`.
- Security rules from `PLAN.md` §3 are mandatory.
- Record notable changes in `CHANGELOG.md` (Keep a Changelog).
