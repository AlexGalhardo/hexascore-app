# Testes e QA

| Camada | Ferramenta | Onde | Quando roda |
| --- | --- | --- | --- |
| Unitário | `bun test` + PGlite | `apps/web/**/*.test.ts` | todo PR |
| Integração | `bun test` + Postgres 18 + Redis (service containers) | `apps/web/tests/integration/` | todo PR |
| Contrato | diff de `openapi.json` | `apps/web/lib/api/` | todo PR |
| Smoke | `bun test` contra `next start` | `apps/web/tests/smoke/` | todo PR |
| Python | `pytest` com fixtures gravadas (sem rede) + `ruff` | `webscrappings/tests/` | todo PR |
| E2E + a11y | Playwright + axe (mobile/desktop, light/dark, en/pt-br/es) | `qa/e2e/` | PR `dev` → `main` |
| Fuzzing da API | Schemathesis a partir do `openapi.json` | `qa/schemathesis/` | PR `dev` → `main` |
| DAST | OWASP ZAP baseline + API scan | `qa/zap/` | PR `dev` → `main` |
| Carga | k6 (cota e rate limit sob carga) | `qa/k6/` | PR `dev` → `main` |
| QA exploratório | agente dirigindo o browser por todas as rotas com edge cases | `qa/agent/` | agendado/manual, relatório em `qa/reports/` |

Comandos exatos de cada camada são adicionados aqui quando a camada for implementada.

## Comandos

```bash
bun run test                          # unit + repositório (PGlite em memória), na raiz
bun test lib/catalog                  # só o catálogo (dentro de apps/web)
bun run lint:ci && bun run typecheck
BETTER_AUTH_SECRET=<64 hex> CRON_SECRET=<64 hex> bun run build

# Ambiente local com dados (PGlite em apps/web/.pglite, sem Docker)
bun run db:migrate
bun run db:seed                       # 50 usuários, 400 itens (100 por tipo), ~4.6k avaliações
(cd apps/web && bun lib/db/seed.ts --users 10000 --items 100000 --batch 1000)   # escala maior
bun run dev
```

- Testes de repositório usam `createTestDatabase()` (`apps/web/lib/db/testing.ts`): PGlite em memória com todas as migrations aplicadas, sem tocar em `.pglite/`. Cada arquivo cria o seu banco no `beforeAll` (~2 s).
- O seed é determinístico (faker 10.6.0, seed fixa; cada item depende só de `(tipo, índice)`), idempotente (upserts) e recusa rodar com `VERCEL_ENV=production`. Usuários `seed-user-NNNNNN@example.test`; ids de trailer começam com `FAKE` (não existem no YouTube).
- Smoke manual: `/`, `/games`, `/pt-br/movies?sort=-score`, uma página de item nos 3 idiomas, `/es/search?q=house`, slug inexistente (404), `GET /api/items/<id>/ratings`, `GET /api/search/suggest?q=...`, `POST /api/cron/revalidate` com `Authorization: Bearer $CRON_SECRET`.

## Dados de teste

- Unitário/integração: fixtures mínimas por teste.
- Dev/preview/sandbox: snapshot de ~100k itens + 10k usuários sintéticos (faker com seed fixa). Nunca em produção.
