# Processo: git, CI/CD, qualidade


**Git**: `feat/*`/`fix/*` → PR para `dev` (sandbox: preview Vercel + branch Neon + seed sintético) → CI verde → PR `dev` → `main` (produção, Rolling Releases). Push direto bloqueado por branch protection (`gh`). Conventional Commits via commitlint + husky. release-please na `main`.

**CI (GitHub Actions)**
- Todo PR: Biome (lint + format) · `tsc --noEmit` · `bun test` (PGlite) · integração (Postgres 18 + Redis services) · diff do `openapi.json` · `next build` + smoke · `ruff` + `pytest` · `bun audit --audit-level=high` · CodeQL · Dependabot.
- PR `dev` → `main` (contra o preview, bloqueante): Playwright E2E (mobile + desktop, light/dark, 3 idiomas, axe) · Schemathesis · OWASP ZAP baseline/API · k6 curto.
- Agendado/manual: agente QA exploratório (dirige o browser por todas as rotas com edge cases: unicode, limites, payloads XSS) → relatório em `qa/reports/` + issue por bug.

**Segurança**: `docs/security.md` com tabela OWASP Top 10:2025 (A01–A10: superfícies, mitigação, commit, riscos aceitos), revisada a cada milestone.

**Agentes**: `AGENTS.md` é a fonte; `CLAUDE.md` importa `@AGENTS.md`. `.claude/rules/`: `karpathy.md`, `ponytail.md`, `graphify.md`. Skills `respondeae-*` reescritas como `hexascore-secure-endpoint` e `hexascore-local-verification`. Worktrees + agentes paralelos para `webscrappings/`, `sdks/` e `docs/`.

**Dependências**: só versões estáveis, fixadas sem `^`. Imagens Docker com tag fixa (`oven/bun:<versão>`, `postgres:18.x`). Nenhuma dependência nova sem aprovação.
