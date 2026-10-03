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

E2E (Playwright + axe) roda em `.github/workflows/qa.yml` no PR `dev` → `main` (ver abaixo). Schemathesis, OWASP ZAP e k6 **ainda não** têm configuração nem workflow (lacuna do 0.6), assim como o job de integração com Postgres 18 + Redis.

## E2E + a11y (0.6) — `qa/e2e/`

Workspace Bun `qa/e2e` (`@playwright/test` 1.63.0, `@axe-core/playwright` 4.13.0). Projetos: `desktop-light`, `desktop-dark` (1280×800) e `mobile-light`, `mobile-dark` (360×740); specs públicas percorrem en/pt-br/es. Fluxos de conta (tag `@auth`) rodam só em `desktop-light` e `mobile-dark` por causa do rate limit por IP do Better Auth.

```bash
# 1. Postgres de QA (container próprio, porta livre) + migrations + seed
docker run -d --name hexascore-qa-pg -e POSTGRES_USER=hexascore -e POSTGRES_PASSWORD=hexascore -e POSTGRES_DB=hexascore -p 127.0.0.1:55432:5432 postgres:18.6-alpine
cd apps/web
export DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:55432/hexascore BETTER_AUTH_SECRET=<64 hex> CRON_SECRET=<64 hex> BETTER_AUTH_URL=http://localhost:3100 NEXT_PUBLIC_APP_URL=http://localhost:3100
bun run db:migrate && bun lib/db/seed.ts --users 10000 --items 100000 --batch 1000
# 2. build + start (NODE_ENV=test = store de cota em memória, a API responde em vez de 503)
bun run build && NODE_ENV=test bun --bun next start -p 3100
# 3. suíte (outro terminal)
cd qa/e2e && bunx playwright install chromium
E2E_BASE_URL=http://localhost:3100 E2E_DATABASE_URL=$DATABASE_URL E2E_BETTER_AUTH_SECRET=$BETTER_AUTH_SECRET bunx playwright test
bunx playwright test --project=desktop-light   # um projeto; --last-failed para repetir falhas
```

- `E2E_BASE_URL` (padrão `http://localhost:3000`), `VERCEL_AUTOMATION_BYPASS_SECRET` (envia `x-vercel-protection-bypass`), `E2E_DATABASE_URL` e `E2E_BETTER_AUTH_SECRET` (opcionais; sem eles as specs `@auth` são puladas com o motivo).
- Link de verificação: o Better Auth 1.7 **não grava** o token de verificação de e-mail (é um JWT HS256 `{ email }` assinado com `BETTER_AUTH_SECRET`), então `qa/e2e/helpers/auth.ts` gera o mesmo token com o secret do deploy e passa pelo `/api/auth/verify-email` real. Tokens de reset de senha estão na tabela `verification` (`helpers/db.ts` → `latestResetToken`). Nenhum backdoor no app.
- Toda spec falha se o browser registrar violação de CSP ou exceção não tratada (`helpers/fixtures.ts`); `expectAccessible()` roda axe (WCAG 2.2 AA) e falha em `serious`/`critical`. Scalar é excluído do axe (markup de terceiros).
- `global-setup.ts` aquece as listagens (o `unstable_cache` serve a versão antiga uma vez depois de um seed novo). Depois de mexer no banco por fora: `POST /api/cron/revalidate` com `{"tags":["catalog"]}`.
- Fluxos do 0.3 (tag `@auth`, exigem `E2E_BETTER_AUTH_SECRET` **e** `E2E_DATABASE_URL`, senão são pulados): `ratings.spec.ts` (avaliar nas 6 categorias, editar, score da comunidade só na 5ª avaliação; status quero/em andamento/concluído/abandonado + estante no perfil; lista criada pela página do item), `social.spec.ts` (perfil público mostra avaliações, privado esconde de todos menos do dono; seguir e ver no `/feed`; `/feed` exige login), `admin.spec.ts` (`/admin` responde 404 para visitante e membro comum; membro denuncia usuário, admin resolve em `/admin/reports`).
- Dados de apoio vão direto no banco (`helpers/db.ts`): `createItem()` (jogo novo sem avaliações, slug `e2e-game-*`), `createMember()` (membro verificado sem credenciais), `insertRating()` (grava e recalcula as estatísticas como `lib/ratings.ts`; inserir antes da primeira visita ao item, que fica em cache) e `grantAdmin()` (igual a `bun run admin:grant`, antes do login). Itens e usuários de teste não são apagados.
- Resultado local (2026-10-03, `next start` + Docker Postgres com seed padrão, 4 projetos): **423 passaram, 0 falharam, 0 pulados** (3 falhas por crash do Chromium com a máquina sob carga passaram com `--last-failed`).

### Seed em escala

`bun lib/db/seed.ts --users 10000 --items 100000 --batch 1000` em Docker Postgres 18 (Docker Desktop/WSL2, Windows 11): **~7,3 min** (437 s: 10k usuários 1,5 s; ~105 s por tipo de 25k itens; recálculo das estatísticas 16 s) → 100k itens, **2.175.836 avaliações**, mediana 27 / p90 965 avaliações por usuário (pool de usuários ativos), 99% dos itens com ≥ 5 avaliações. As avaliações entram em `unnest` colunar de 20k linhas (9 parâmetros por statement) com `synchronous_commit off`; ~25% do tempo é geração em JS (faker por avaliação). Idempotente: rodar de novo faz upsert das mesmas linhas.
## Comandos

```bash
bun run test                          # unit + repositório (PGlite em memória), na raiz
bun test lib/catalog                  # só o catálogo (dentro de apps/web)
bun run lint:ci && bun run typecheck
BETTER_AUTH_SECRET=<64 hex> CRON_SECRET=<64 hex> bun run build

# Ambiente local com dados (PGlite em apps/web/.pglite, sem Docker)
bun run db:migrate
bun run db:seed                       # 50 usuários, 400 itens (100 por tipo), ~2.4k avaliações
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
