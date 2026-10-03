# Hexascore

Base de dados de **jogos, filmes, séries e livros** em que cada usuário avalia a obra em **6 categorias de 0 a 100**, com rede social simplificada e uma **API REST paga**. Codinome técnico `hexascore` (nome de marca final pendente). Plano completo em [`PLAN.md`](PLAN.md).

Stack: Bun 1.4.2 (workspaces) · Next.js 16 (`apps/web`) · Drizzle + Postgres 18 (PGlite sem Docker) · Better Auth · Upstash Redis · Vercel. Ingestão em Python (`webscrappings/`, uv).

## Quickstart

Requisitos: [Bun 1.4.2](https://bun.sh) e, opcionalmente, Docker.

```sh
# Com Docker (Postgres 18 + Redis + proxy REST Upstash)
./setups/setup-unix-docker.sh
# Sem Docker (PGlite em processo)
./setups/setup-unix-pglite.sh

bun run dev   # http://localhost:3000
```

No Windows: `powershell -ExecutionPolicy Bypass -File setups\setup-windows-docker.ps1` (ou `setup-windows-pglite.ps1`). Detalhes em [`setups/README.md`](setups/README.md).

## Comandos

| Comando | O que faz |
| --- | --- |
| `bun install` | instala as dependências do workspace |
| `bun run dev` | servidor de desenvolvimento (`apps/web`) |
| `bun run build` / `bun run start` | build e servidor de produção |
| `bun run lint` / `bun run lint:ci` | Biome (check local / modo CI) |
| `bun run format` | formata com Biome |
| `bun run typecheck` | `tsc --noEmit` em `apps/web` |
| `bun run test` | testes unitários (`bun test`, PGlite) |
| `bun run db:generate` | gera migration a partir do schema Drizzle |
| `bun run db:migrate` | aplica as migrations (Postgres se `DATABASE_URL`, senão PGlite) |
| `bun audit --audit-level=high` | auditoria de vulnerabilidades das dependências |
| `docker compose -f infra/docker-compose.yaml up -d --wait` | sobe Postgres, Redis e o proxy Upstash locais |

## Estrutura

- `apps/web/` — app Next.js (App Router, i18n en/pt-BR/es), Drizzle em `lib/db/`, env validado com zod em `lib/env.ts`
- `infra/` — `docker-compose.yaml` local e `Dockerfile` do app
- `setups/` — scripts de setup local (Unix/Windows × Docker/PGlite)
- `.github/` — CI, CodeQL, release-please e Dependabot
- `docs/` — documentação de referência

## Documentação

[Arquitetura](docs/architecture.md) · [Modelo de dados](docs/data-model.md) · [API](docs/api.md) · [Frontend](docs/frontend.md) · [Design system](docs/design-system.md) · [Pagamentos](docs/payments.md) · [Ingestão](docs/ingestion.md) · [Segurança](docs/security.md) · [Testes](docs/testing.md) · [Processo, CI/CD e releases](docs/process.md)

Mudanças notáveis em [`CHANGELOG.md`](CHANGELOG.md). Guia para agentes de IA em [`AGENTS.md`](AGENTS.md).
