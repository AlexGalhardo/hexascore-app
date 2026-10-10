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
| Scripts de setup e QA | `docker run --rm -v "$PWD:/mnt" -w /mnt koalaman/shellcheck:v0.11.0 -x setups/*.sh qa/*/run.sh` | — |
| Carga, fuzzing, DAST | `qa/k6/run.sh`, `qa/schemathesis/run.sh`, `qa/zap/run.sh` (ver abaixo) | `.github/workflows/qa.yml` |

`lint`, `typecheck`, `test` e `build` são os checks obrigatórios do ruleset de `dev` e `main` (`.github/rulesets/protect-main-dev.json`; aplicação pendente, ver `docs/process.md`). Antes de abrir PR rode os quatro localmente.

Banco para testar migrations à mão: `setups/setup-*-docker.*` (Postgres 18 em `localhost:5432`) ou `setups/setup-*-pglite.*`; ambos rodam `bun run db:migrate`.

E2E (Playwright + axe), k6, Schemathesis e OWASP ZAP rodam em `.github/workflows/qa.yml` no PR `dev` → `main` contra o preview da Vercel (ver abaixo). O job de integração com Postgres 18 + Redis ainda é lacuna.

## API: k6, Schemathesis e ZAP (0.6) — `qa/k6/`, `qa/schemathesis/`, `qa/zap/`

Os três rodam em imagens Docker fixas (nenhuma dependência npm): `grafana/k6:2.3.0`, `schemathesis/schemathesis:4.29.0`, `ghcr.io/zaproxy/zaproxy:2.17.0`. Cada um tem um `run.sh` que o CI e o uso local chamam igual. Relatórios em `qa/reports/{k6,schemathesis,zap}/` (gitignored; artifacts no CI).

| Variável | Uso |
| --- | --- |
| `QA_BASE_URL` | origem do deploy (`http://localhost:3100`, preview); a API é `$QA_BASE_URL/api/v1` |
| `QA_DOCKER_BASE_URL` | só no Docker Desktop (sem rede host): `http://host.docker.internal:3100` |
| `QA_API_KEY` | chave **Pro/Business** no banco do deploy (Free = 100 requests na vida e 1 req/s). Gasto por execução: k6 ~600, Schemathesis ~2k, ZAP alguns milhares |
| `QA_FREE_API_KEY` | opcional, só k6: chave Free reservada para o 402 (a 1ª execução gasta a cota restante em ~100 s; as seguintes recebem 402 direto) |
| `VERCEL_AUTOMATION_BYPASS_SECRET` | envia `x-vercel-protection-bypass` |

```bash
# servidor local como no E2E (NODE_ENV=test = store de cota em memória); chaves criadas em /profile/api,
# plano pago via linha em `subscriptions` (status active, current_period_end no futuro)
export QA_BASE_URL=http://localhost:3100 QA_DOCKER_BASE_URL=http://host.docker.internal:3100 QA_API_KEY=hx_live_…
qa/k6/run.sh             # QA_FREE_API_KEY=… para o 402; RATE=8 (req/s da fase de carga, abaixo do limite do plano)
qa/schemathesis/run.sh   # args extras vão para `st run` (ex.: --max-examples 5)
qa/zap/run.sh
```

- **k6** (`qa/k6/api.js`): `setup` valida a chave (`/me/usage`, recusa Free) e pega ids reais; fase `load` a `RATE` req/s por 60 s em listas, filtros, gêneros, busca e detalhe — thresholds `p(95)<300` ms e 99% de 200; fase `burst` (depois da carga) dispara 3× o limite por segundo e exige ≥ 1 `429` com `Retry-After` e corpo problem; fase `quota` (com `QA_FREE_API_KEY`) exige `limit=11` → 400 e, ao esgotar a cota, `402` problem.
- **Schemathesis** (`qa/schemathesis/schemathesis.toml`): fuzz de `apps/web/public/openapi.json` (o do commit) com todas as checagens de conformidade (status, content-type, headers, schema da resposta), `not_a_server_error`, `ignored_auth`, aceitação/rejeição de dados; `rate-limit = "8/s"`. Desligados com motivo no arquivo: `unsupported_method` (o servidor HTTP responde 400 a `TRACE` antes do Next; os demais métodos dão 405), `use_after_free`/`ensure_resource_availability` (API só leitura); `cursor` aleatório aceita 400 (é um token opaco). O `run.sh` falha antes de fuzzar se a chave for inválida ou Free. Aviso esperado: detalhe por id e `/lookup` respondem 404 para ids gerados (o 200 do detalhe é coberto por `lib/api/v1.test.ts` e pelo k6).
- **ZAP** (`qa/zap/plan.yaml`, Automation Framework): spider + passivo no site (exclui `/api/auth` e `/api/cron`), import do `openapi.json` e active scan só em `/api/v1` (2 threads, 150 ms de atraso), relatórios HTML/JSON e `exitStatus` com `errorLevel: High` — **falha só em alerta High**; Medium/Low ficam no relatório. Falsos positivos entram no job `alertFilter` do plano, cada um com a justificativa. Sem `QA_API_KEY` o scan da API vê só 401.
- CI: `k6` → `schemathesis` → `zap` em sequência (compartilham a chave e o limite por segundo), em paralelo ao `e2e`. Sem `QA_API_KEY`, k6 e Schemathesis são pulados com `::notice::`; o ZAP roda assim mesmo.

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

### Sandbox (dev.hexascore.app)

O sandbox roda num **branch Neon copiado da produção** depois do bootstrap de ~100k itens: catálogo real, usuários/avaliações/listas fictícios. Dados fictícios existem **só nesse branch**, nunca na produção.

1. Criar o branch (console Neon ou `neonctl branches create --name sandbox --parent <branch de produção>`) e pegar a connection string (`neonctl connection-string sandbox`).
2. Rodar o seed sobre os itens existentes (em `apps/web`, com `DATABASE_URL` do branch):
   ```bash
   SEED_TARGET=sandbox bun lib/db/seed.ts --existing-items --users 10000 --batch 1000
   # se a cópia já tem > 50 usuários reais verificados: acrescente --i-know-this-is-sandbox
   ```
   `--existing-items` **não cria nem altera itens** (só `rating_count`/`score`/`score_c_*`): cria os usuários seed, avalia os itens do banco em lotes keyset por id (o catálogo nunca fica inteiro na memória; o rank de popularidade por tipo vem de `row_number()` no Postgres) com a mesma curva Zipf do seed de fixture aplicada ao rank (~2M avaliações em 10k × 100k), qualidade por item derivada do hash do id, viés por usuário e atividade heavy-tailed como antes; recalcula as estatísticas em SQL no mesmo lote (fórmula de `lib/ratings.ts`). Depois, por lote de usuários, SQL set-based: status (70% concluído, 15% em andamento, 10% abandonado; "quero" a partir das avaliações do vizinho), time to beat para metade dos jogos concluídos (30–600000 min, só com status `completed`), 2–21 follows por usuário enviesados para os power raters, e listas (1/3 dos pares usuário×tipo com 5+ avaliações; 1–10 por tipo = `LIST_LIMITS.perType`; ids `md5('seed-list:…')`; até 25 itens do mesmo tipo, os mais bem avaliados pelo dono). Idempotente: rodar de novo converge nas mesmas linhas.
3. Deploy do sandbox com `DATABASE_URL` do branch e `NEXT_PUBLIC_SANDBOX=true`: faixa fixa (não dispensável, `role="status"`) acima da navbar em en/pt-BR/es, `<meta name="robots" content="noindex, nofollow">` e `robots.txt` com `Disallow: /` (além do bloqueio que já vale fora de `VERCEL_ENV=production`). Depois do seed: `POST /api/cron/revalidate` `{"tags":["catalog"]}`.
4. Reset: `SEED_TARGET=sandbox bun lib/db/seed.ts --purge` apaga todos os usuários seed (cascade em avaliações, status, time to beat, follows e listas) e recalcula as estatísticas de todos os itens em lotes. Para recomeçar do zero com o catálogo atual da produção, `neonctl branches reset sandbox --parent` é mais rápido.

**Marcador**: usuário seed = id `seed-user-NNNNNN` **e** e-mail `seed-user-NNNNNN@example.test`; todo dado social fictício pertence a um deles. O `--purge` exige os dois.

**Guarda** (`lib/db/seed-guard.ts`, testada em `seed-guard.test.ts`), antes de abrir conexão: `VERCEL_ENV=production` → recusa sempre; `DATABASE_URL` vazio (PGlite) ou host `localhost`/`127.0.0.1`/`::1` → ok; qualquer outro host só com `SEED_TARGET=sandbox` (único valor aceito). Depois de conectar (seed, não purge): recusa se houver **> 50 usuários verificados fora do marcador** (id não `seed-user-*` e e-mail fora de domínio `.test`, que cobre E2E) sem `--i-know-this-is-sandbox`, com a explicação no terminal.

**Produção** (decisão do dono, 2026-10-04: dados de demonstração antes do launch): `SEED_TARGET=production` **e** `--i-know-this-is-production` juntos liberam o alvo mesmo com `VERCEL_ENV=production`; só `--existing-items` ou `--purge` são aceitos, a checagem de usuários reais é dispensada e o admin padrão (senha pública) **não** é criado. Rodar localmente com um env mínimo (o `.env.production.local` puxado da Vercel tem placeholders `[SENSITIVE]` que o `env.ts` recusa): `DATABASE_URL` de produção, `VERCEL_ENV=production`, `BETTER_AUTH_URL=https://hexascore.app`.

```sh
SEED_TARGET=production bun --env-file=<env mínimo> lib/db/seed.ts --existing-items --users 10000 --batch 1000 --i-know-this-is-production
SEED_TARGET=production bun --env-file=<env mínimo> lib/db/seed.ts --purge --i-know-this-is-production   # remover
```

Enquanto os dados existirem, `FLAG_SEED_DATA=on` na Vercel liga a flag `seed-data`, que mostra o lembrete em `/admin`; desligue depois do `--purge` e rode `POST /api/cron/revalidate` `{"tags":["catalog"]}`.

Medição local (Docker Postgres 18 no Windows, 2026-10-03): 5k itens + `--existing-items --users 2000 --batch 1000` = **~60–76 s** (205k avaliações ~30–44 s; status/time to beat/follows/listas ~30 s, metade nos triggers de FK de `item_status`); rerun ~33 s; `--purge` ~2–4 s. Volume e tempo escalam ~linear com as avaliações: **10k × 100k ≈ 2M avaliações → estimativa de ~12–15 min** local. No Neon a rede pesa pouco (poucos statements por lote de 1000 itens/usuários); o que manda é o tamanho do compute.
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
- O seed é determinístico (faker 10.6.0, seed fixa; cada item depende só de `(tipo, índice)`), idempotente (upserts) e passa pela guarda de alvo (ver "Sandbox" acima: produção nunca, remoto só com `SEED_TARGET=sandbox`). Usuários `seed-user-NNNNNN@example.test`; ids de trailer começam com `FAKE` (não existem no YouTube).
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
- Dev/preview/sandbox: snapshot de ~100k itens + 10k usuários sintéticos (faker com seed fixa). Em produção só como dado de demonstração pré-launch, com opt-in duplo (ver "Produção" abaixo da Guarda).

Admin padrão do seed (local/sandbox): ver [`docs/admin.md`](admin.md).
