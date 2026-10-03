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

## Comandos implementados

| Verificação | Local (raiz do repo) | CI (`.github/workflows/ci.yml`) |
| --- | --- | --- |
| Lint + format | `bun run lint:ci` | job `lint` |
| Contrato OpenAPI | `bun run openapi:check` (regerar: `bun run openapi`) | job `lint` |
| Tipos | `bun run typecheck` | job `typecheck` |
| Unitário (PGlite, sem Docker) | `bun run test` (ou `bun test` em `apps/web`) | job `test` |
| Build | `bun run build` | job `build` (depois de lint/typecheck/test) |
| Vulnerabilidades | `bun audit --audit-level=high` | job `audit` |
| SAST | — | `.github/workflows/codeql.yml` |
| Workflows | `docker run --rm -v "$PWD:/repo" -w /repo rhysd/actionlint:1.7.12` | — |
| Scripts de setup | `docker run --rm -v "$PWD:/mnt" -w /mnt koalaman/shellcheck:v0.11.0 -x setups/*.sh` | — |

`lint`, `typecheck`, `test` e `build` são os checks obrigatórios do ruleset de `dev` e `main` (`.github/rulesets/protect-main-dev.json`; aplicação pendente, ver `docs/process.md`). Antes de abrir PR rode os quatro localmente.

Banco para testar migrations à mão: `setups/setup-*-docker.*` (Postgres 18 em `localhost:5432`) ou `setups/setup-*-pglite.*`; ambos rodam `bun run db:migrate`.

Ainda **não** existem workflows de E2E (Playwright), Schemathesis, OWASP ZAP nem k6: entram no milestone em que as respectivas camadas forem criadas (gatilho: PR `dev` → `main`), junto com o job de integração com Postgres 18 + Redis como service containers.
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

## API pública (0.4)

```bash
cd apps/web && bun test lib/api          # rotas /v1 (PGlite + store em memória) + cota
bun run openapi:check                    # openapi.json commitado == schemas zod
bunx @redocly/cli@2 lint apps/web/public/openapi.json   # validação do spec (sem virar dependência)
```

- `lib/api/v1.test.ts` chama os `GET` exportados das rotas com o `db` mockado (`mock.module("@/lib/db/client")`) para um PGlite em memória e o store de cota em memória (`getStore()` com `NODE_ENV=test`); o throttle é desligado por padrão e religado no teste de 429. Cobre 401, 400 (inclui `limit=11` no Free), 402 na 101ª request, 429 + `Retry-After`, 404, 503 com store quebrado, headers, cada endpoint e a ausência de `external_scores`.
- Smoke manual com `next start`: o store em memória só existe fora de `NODE_ENV=production`, e o Redis comum (inclusive o `serverless-redis-http` do `infra/`) rejeita o script do `@upstash/ratelimit` (`allow-key-locking`, específico do Upstash) — a API responde 503 (fail closed). Para exercitar tudo localmente: `bun run build` e depois `NODE_ENV=test bun --bun next start` em `apps/web` (store em memória), criar uma chave em `/profile/api` e chamar `curl -H "Authorization: Bearer hx_live_…" localhost:3000/api/v1/games`. Para o host da API: `curl -H "Host: api.localhost:3000" localhost:3000/v1/games ...`.
- PGlite + `next start`: cada bundle (route handler × página) abre sua própria instância do `.pglite` e não enxerga escritas da outra até reiniciar (o cache em `globalThis` de `lib/db/client.ts` some no build de produção). Ex.: login via `/api/auth/*` só vale nas páginas depois de reiniciar o servidor. Com Postgres (Docker/Neon) não acontece.
- SDKs: `cd sdks/ts && bun run typecheck`; Python: `uv run --no-project --with ./sdks/python/hexascore python -c "import hexascore"`.

## Dados de teste

- Unitário/integração: fixtures mínimas por teste.
- Dev/preview/sandbox: snapshot de ~100k itens + 10k usuários sintéticos (faker com seed fixa). Nunca em produção.
