# CLAUDE.md

@AGENTS.md

At the start of every session, read `.agents/skills/use-railway/SKILL.md` and <https://railway.com/agents/grok> before any Railway command. The skill in this repo wins over a global copy. Do not run `railway up` unless the owner asked for a deploy.

Rules and skills live in `.agents/rules/` and `.agents/skills/` (provider-agnostic). Claude Code does not load them on its own here: follow the "Rules" section of AGENTS.md and read the matching files by path.
