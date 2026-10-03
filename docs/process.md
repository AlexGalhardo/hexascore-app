# Processo: git, CI/CD, qualidade


**Git**: `feat/*`/`fix/*` → PR para `dev` (sandbox: preview Vercel + branch Neon + seed sintético) → CI verde → PR `dev` → `main` (produção, Rolling Releases). Push direto bloqueado por branch protection (`gh`). Conventional Commits via commitlint + husky. release-please na `main`.

**CI (GitHub Actions) — implementado**

| Workflow | Gatilho | Jobs / o que roda |
| --- | --- | --- |
| `.github/workflows/ci.yml` | PR e push em `dev`/`main` | `lint` (`bun run lint:ci`) · `typecheck` (`bun run typecheck`) · `test` (`bun test` em `apps/web`, PGlite) · `build` (`bun run build`, depende dos 3 anteriores; env dummy válido para `lib/env.ts`) · `audit` (`bun audit --audit-level=high`) |
| `.github/workflows/codeql.yml` | PR/push em `dev`/`main`, semanal, manual | CodeQL `javascript-typescript` + `python` (python só quando houver `.py`). Repo privado: só roda com a variável de repo `CODE_SCANNING_ENABLED=true` (exige GitHub Code Security) |
| `.github/workflows/release-please.yml` | push em `main` | `googleapis/release-please-action@v5` com `release-please-config.json` + `.release-please-manifest.json` |
| `.github/dependabot.yml` | semanal | `bun` (raiz), `uv` (`webscrappings/`), `github-actions`, `docker` + `docker-compose` (`infra/`); minor/patch agrupados, PRs para `dev`, prefixos `chore(deps)`/`ci` |

Todos os jobs usam `.github/actions/setup-bun` (Bun `1.4.2` via `oven-sh/setup-bun@v2` + cache de `~/.bun/install/cache` + `bun install --frozen-lockfile`), `concurrency` com `cancel-in-progress` e `permissions` mínimas. Actions fixadas por major: `actions/checkout@v7`, `actions/cache@v6`, `github/codeql-action@v4`. Validar localmente: `docker run --rm -v "$PWD:/repo" -w /repo rhysd/actionlint:1.7.12`.

**Releases**: release-please (tipo `simple`, manifest em `0.0.0`) abre/atualiza um PR de release na `main` a cada push; `feat`/`fix` antes da 1.0 sobem o patch (primeira release `0.0.1`), merge do PR cria tag `vX.Y.Z` + GitHub Release e escreve a seção no `CHANGELOG.md` (`feat`→Added, `fix`→Fixed, `perf`/`refactor`/`revert`→Changed; `docs`/`chore`/`ci`/`test`/`build`/`style` ocultos). A seção `[Unreleased]` do `CHANGELOG.md` continua manual. PRs abertos pelo `GITHUB_TOKEN` não disparam workflows, então os checks obrigatórios não rodam no PR de release: o dono faz o merge como admin (branch protection não força admins).

**CI — alvo (milestones seguintes)**
- Todo PR: Biome (lint + format) · `tsc --noEmit` · `bun test` (PGlite) · integração (Postgres 18 + Redis services) · diff do `openapi.json` · `next build` + smoke · `ruff` + `pytest` · `bun audit --audit-level=high` · CodeQL · Dependabot.
- PR `dev` → `main` (contra o preview, bloqueante): Playwright E2E (mobile + desktop, light/dark, 3 idiomas, axe) · Schemathesis · OWASP ZAP baseline/API · k6 curto.
- Agendado/manual: agente QA exploratório (dirige o browser por todas as rotas com edge cases: unicode, limites, payloads XSS) → relatório em `qa/reports/` + issue por bug.

**Segurança**: `docs/security.md` com tabela OWASP Top 10:2025 (A01–A10: superfícies, mitigação, commit, riscos aceitos), revisada a cada milestone.

**Agentes**: `AGENTS.md` é a fonte; `CLAUDE.md` importa `@AGENTS.md`. `.claude/rules/`: `karpathy.md`, `ponytail.md`, `graphify.md`. Skills `respondeae-*` reescritas como `hexascore-secure-endpoint` e `hexascore-local-verification`. Worktrees + agentes paralelos para `webscrappings/`, `sdks/` e `docs/`.

**Dependências**: só versões estáveis, fixadas sem `^`. Imagens Docker com tag fixa (`oven/bun:<versão>`, `postgres:18.x`). Nenhuma dependência nova sem aprovação.
