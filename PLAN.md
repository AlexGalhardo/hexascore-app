# Hexascore — Plano do Projeto

> Codinome técnico: **hexascore**. Nome de marca final é decidido antes do launch (ver [Pendências](#pendências-antes-do-launch)); a marca vem de `APP_NAME` + mensagens i18n, então trocar custa pouco.

Plano vivo. Legenda: `[ ]` a fazer · `[~]` em andamento · `[x]` feito. Regra: **1 tarefa = 1 microcommit** (Conventional Commits). Cada milestone tem critérios de verificação — só está pronto quando todos passam.

---

## 1. Produto

Base de dados + rede social simplificada de **jogos, filmes, séries e livros** (inspiração: IMDb, IGDB, Goodreads), com **API REST paga** de DX referência (inspiração: Resend).

Diferencial: usuários **não escrevem resenhas** — avaliam cada obra em **6 categorias de 0 a 100**. A nota do usuário é a média das 6; o **metascore** do item é a média das notas dos usuários.

### Categorias por seção (todas obrigatórias, inteiros 0–100)

| Seção | c1 | c2 | c3 | c4 | c5 | c6 |
| --- | --- | --- | --- | --- | --- | --- |
| Jogos | Gameplay | Diversão | Qualidade Sonora | Qualidade Técnica | Qualidade Visual | Storytelling |
| Filmes | Roteiro | Atuação | Direção | Visual/Fotografia | Trilha/Som | Entretenimento |
| Séries | Roteiro | Atuação | Personagens | Ritmo | Produção/Visual | Entretenimento |
| Livros | Enredo | Personagens | Escrita/Estilo | Ritmo | Originalidade | Impacto |

### Regras de avaliação

- Avalia-se a **obra** (jogo em qualquer plataforma, série inteira, livro como obra — não edição). Plataforma/temporada/edição são metadados.
- 1 avaliação por usuário por item, editável. Exige e-mail verificado.
- Nota do usuário = `round(média das 6)`. Metascore = média das notas dos usuários + média por categoria.
- Metascore só aparece com **≥ 5 avaliações**; antes disso: "nota insuficiente".
- Conta excluída: avaliações permanecem anonimizadas ("usuário removido") para não distorcer o metascore.

### Social (v1.0.0)

Perfil público `/u/{nickname}` (opção privado) · status por item (quero / em andamento / concluído / abandonado) · seguir usuários + feed · listas curadas públicas · denúncias + moderação no `/admin`.

### Monetização

- **Afiliados** (Amazon, Mercado Livre) só em livros e jogos: links curados no `/admin` + fallback de busca por ISBN/título. Aviso de "link de afiliado" visível.
- **API paga** — modelo híbrido (assinatura com cota + créditos avulsos):

| Plano | Preço/mês | Cota | `limit` máx. | Rate |
| --- | --- | --- | --- | --- |
| Free | R$0 | 100 req **no total da conta** | 10 | 1 req/s |
| Pro | R$29 / US$6 / €6 | 50.000 req/mês | 100 | 10 req/s |
| Business | R$149 / US$29 / €27 | 500.000 req/mês | 100 | 50 req/s |
| Créditos extras | R$19 / US$4 / €4 | +25.000 req, não expiram | — | — |

Valores ficam em config (não hardcoded). PIX: 1 mês ou 12 meses (12 pelo preço de 10). Créditos só para assinantes Pro/Business. Free exige e-mail verificado (1 conta = 1 cota free).

---

## 2. Decisões travadas

| Tema | Decisão | Motivo |
| --- | --- | --- |
| Licença de dados | **Híbrido**: API expõe só dados próprios (notas, metascore), metadados de fontes livres (Wikidata CC0, Open Library) e IDs externos. Notas de terceiros só no frontend, com atribuição + link | IMDb/TMDB/IGDB/Metacritic/RT/Goodreads/HLTB proíbem redistribuição comercial |
| Fontes | **APIs + Wikidata, sem scraping**. IMDb/MC/RT/HLTB/Goodreads/Skoob/Epic = só deep-link via ID do Wikidata. TMDB/IGDB fora da v1 (exigem licença comercial) | Risco jurídico mínimo |
| Sequência | **Horizontal**: as 4 seções avançam juntas, camada por camada | Escolha do dono |
| Monorepo | **Bun workspaces** (`apps/web`, `sdks/ts`), **sem Turborepo** | Só 1 consumidor TS real; Turborepo entra com ≥ 3 pacotes JS com dependência de build |
| API | Dentro do Next (`app/api/v1`), `api.hexascore.com` → rewrite para `/api/v1` | 1 projeto Vercel |
| Banco | Postgres-only. Prod: **Neon** (branch por preview). Dev/testes: **PGlite**. Paridade/seed/QA: **Docker Postgres 18**. SQLite removido | 1 schema Drizzle, sem dupla manutenção |
| Busca | Postgres nativo: `pg_trgm` + `unaccent` + `tsvector` por idioma | Suficiente para 100k–1M itens, zero infra |
| Rate limit/cota | **Upstash Redis** + flush diário para Postgres | Não sobrecarregar o Neon com 1 write/request |
| Ingestão | Python 3.13 + uv em `webscrappings/`, rodando em **GitHub Actions cron**, staging `source_records` + normalizador | Custo zero, sem limite de 300s |
| i18n UI | `next-intl`, EN (padrão, sem prefixo) / `/pt-br` / `/es` | — |
| i18n conteúdo | Títulos via Wikidata (CC0); sinopse via Wikipedia (CC BY-SA, com `license` + `source_url`), fallback EN com aviso | — |
| Auth | **Better Auth**: e-mail+senha, Google, verificação de e-mail obrigatória, 2FA TOTP + backup codes, passkeys, plugin admin | — |
| Senha Google | Conta Google nasce sem senha; `/profile` oferece "Definir senha" (`setPassword`) | Senha invisível não tem uso e aumenta superfície |
| Exclusão de conta | 30 dias de carência; login mostra banner com "cancelar exclusão"; cron anonimiza após 30d | — |
| Pagamentos | **Stripe** Checkout + Customer Portal (BRL/USD/EUR) · **AbacatePay** PIX avulso | PIX não tem recorrência confiável |
| Vercel extra | Analytics + Speed Insights · BotID + Firewall · Blob · Flags SDK + Rolling Releases | — |
| Observabilidade | Vercel Observability + logger Telegram para alertas. Sem Sentry por ora | — |
| Release | **release-please** (Conventional Commits → semver, CHANGELOG, tag, GitHub Release), começando em `0.0.1` | — |
| Visual | **Dados editorial**: almanaque denso, serif nos títulos + sans/mono nos números, radar hexagonal das 6 categorias como assinatura, escala de cor de nota estilo Metacritic, 1 acento de cor por seção | — |
| Admin | `/admin` mínimo (role admin, só EN, tabelas) | — |
| Seeds | Prod: 100k itens reais, zero usuários fake. Dev/preview/sandbox: snapshot de itens + 10k usuários sintéticos determinísticos | Notas fake em prod distorcem o metascore |
| Diagramas | Mermaid nos `.md`; React Flow só em `/docs` públicas onde interatividade ajuda | GitHub renderiza Mermaid nativo |

---

## 3. Arquitetura

```
hexascore/                 Bun workspaces: apps/web, sdks/ts
├─ apps/web/               Next 16 — site + /api/v1 + /admin (todo o TS do produto)
│   ├─ app/[locale]/       páginas
│   ├─ app/api/v1/         API pública
│   ├─ app/api/{auth,webhooks,cron}/
│   ├─ actions/            Server Actions ("use server" só aqui)
│   ├─ lib/db/             Drizzle schema, migrations, client (Neon | PGlite), seed
│   ├─ lib/api/            schemas zod → openapi.json (fonte única do contrato)
│   ├─ lib/                metascore, cota, planos, pagamentos, auth, session
│   ├─ emails/             react-email (Resend)
│   ├─ components/         Base UI + Tailwind 4
│   └─ proxy.ts            locale, host api.* → /api/v1, CSP com nonce
├─ sdks/ts/  sdks/python/  gerados do openapi.json (zero código à mão)
├─ webscrappings/          Python 3.13 + uv (ingestão)
├─ qa/                     Playwright, agente QA, Schemathesis, ZAP, k6
├─ infra/                  docker-compose (Postgres 18 + Redis)
└─ docs/                   arquitetura, ADRs, security.md
```

```mermaid
flowchart LR
  GH[GitHub Actions cron] --> PY[webscrappings Python]
  PY -->|payload bruto| SR[(source_records)]
  SR -->|normalize.py| CAT[(items / translations / external_scores)]
  PY -->|POST /api/cron/revalidate| WEB
  U[Usuário] --> WEB[apps/web RSC]
  WEB --> CAT
  DEV[Dev cliente API] -->|Bearer hx_live_...| API[/api/v1]
  API --> R[(Upstash Redis: key cache, rate limit, cota)]
  API --> CAT
  R -->|flush diário| AU[(api_usage_daily / credit_ledger)]
```

### Regras de segurança (obrigatórias)

- Quem age vem sempre da sessão (`getSessionUser()`), nunca do body.
- `"use server"` só em `actions/` (em `lib/` vira endpoint público) — teste-guarda no CI.
- Nenhuma linha inteira de `user` vai ao browser: DTOs explícitos.
- Valores de dinheiro e cota calculados só no servidor.
- Segredo nunca em `NEXT_PUBLIC_*`.
- CSP estrita com nonce. Webhooks: assinatura/HMAC com comparação de tempo constante + idempotência.
- BotID em login, signup, contato e checkout. Regras de rate limit no Vercel Firewall.

---

## 4. Modelo de dados (Drizzle + Postgres)

**Catálogo**
- `items`: `id` uuid v7 (público), `type` (`game|movie|series|book`), `slug` (único por tipo), `wikidata_qid` (único), `release_date`, `external_ids jsonb` (imdb, igdb, steam, tmdb, isbn13, openlibrary, hltb, goodreads, metacritic, rt, epic, gog), `meta jsonb` por tipo (genres, platforms, runtime, pages, seasons, trailers; índice GIN), stats desnormalizadas `rating_count`, `score`, `score_c1..c6` (atualizadas na mesma transação da avaliação).
- `item_translations` (`item_id`, `locale`) PK: `title`, `synopsis`, `synopsis_license`, `synopsis_source_url`, `search tsvector` gerado, índice `pg_trgm` em `title`.
- `external_scores` (`item_id`, `source`, `score`, `url`, `fetched_at`) — **nunca sai pela API**.
- `affiliate_links` (`item_id`, `store`, `url`, `active`).

**Avaliação e social**
- `ratings` (`user_id`, `item_id`) PK: `c1..c6 smallint CHECK (0..100)`, `overall` gerado = `round((c1+…+c6)/6.0)`. Nomes das categorias vivem em código.
- `item_status` (`user_id`, `item_id`, `status`) · `follows` (`follower_id`, `followee_id`) · `lists`, `list_items` · `reports`.

**Auth** (gerado pelo Better Auth CLI): `user`, `session`, `account`, `verification`, `two_factor`, `passkey`. Extras em `user`: `nickname` único, `role`, `profile_private`, `deletion_scheduled_at`.

**API e billing**
- `api_keys`: plugin apiKey do Better Auth se servir; senão 1 tabela (`prefix`, `hash` sha256, `revoked_at`, `last_used_at`).
- `subscriptions` (`user_id`, `plan`, `provider`, `status`, `current_period_end`, `provider_ref`).
- `credit_ledger` (`user_id`, `delta`, `reason`, `ref`) — saldo = soma, append-only.
- `pix_charges` (`provider_ref` único, `amount_cents`, `purpose`, `status`).
- `api_usage_daily` (`user_id`, `date`, `count`).
- Últimas 100 requests por usuário ficam no Redis (`LPUSH` + `LTRIM`), não no Postgres.

**Outros**: `contact_messages` · `source_records` (`source`, `external_id`) PK com `payload jsonb`, `hash`, `fetched_at`, `item_id`.

Fora por ora: tabela de gêneros/plataformas (entra quando gênero ganhar página traduzida), `ingestion_runs` (logs do Actions bastam).

---

## 5. API pública `/v1`

- Base `https://api.hexascore.com/v1`. Auth `Authorization: Bearer hx_live_…` (chave exibida 1 vez, só hash no banco). Só leitura (`GET`).

| Rota | Uso |
| --- | --- |
| `/v1/{games\|movies\|series\|books}` | lista + filtros: `genre`, `platform`, `year`, `min_score`, `max_score`, `min_{categoria}` (ex. `min_gameplay=80`), `sort` (`score`, `-score`, `release_date`, `rating_count`), `limit`, `cursor`, `locale` |
| `/v1/{type}/{id}` | detalhe: metadados, metascore, médias por categoria, IDs externos, sinopse com `license` + `source_url` |
| `/v1/{type}/genres` | taxonomia para filtros |
| `/v1/search?q=&type=` | busca fuzzy entre tipos |
| `/v1/lookup?source=imdb&id=tt…` | busca por ID externo |
| `/v1/me/usage` | cota restante, plano, créditos |

- Lista: `{ "data": [...], "next_cursor": "…" | null }` (paginação por cursor).
- Erros RFC 9457 (`application/problem+json`) com `type` linkando a página do erro nas docs e `request_id`.
- Headers: `RateLimit-*` (IETF), `X-Quota-Remaining`, `X-Request-Id`. 402 sem cota/créditos (link para `/pricing`); 429 com `Retry-After`.
- Cota conta toda request autenticada, exceto 429 e 5xx.
- Versionamento por path; quebra → `/v2`, antiga com headers `Deprecation` + `Sunset` e changelog da API.
- Dados de catálogo com cache do Next (`"use cache"` + tags invalidadas pela ingestão); checagem de cota sempre roda.
- Contrato: zod 4 → `openapi.json` 3.1 gerado no build; CI falha se o arquivo mudar sem commit.
- Docs: `/docs` (quickstart, auth, erros, paginação, limites) + referência Scalar (snippets cURL/JS/Python automáticos). `/profile/api` = mesmo Scalar com a chave do usuário + dashboard de uso.
- SDKs gerados: `@hey-api/openapi-ts` (npm) e `openapi-python-client` (PyPI), publicados via release-please.

Fora por ora: endpoints de escrita, notas individuais de usuários, webhooks para clientes, GraphQL.

---

## 6. Frontend e rotas

Locale: EN sem prefixo, `/pt-br/…`, `/es/…`. Slugs de rota em inglês nos 3 idiomas. Tudo mobile-first, light/dark.

| Rota | Conteúdo |
| --- | --- |
| `/` | navbar com busca global (autocomplete) + blocos por seção (melhores, lançamentos, mais avaliados) |
| `/games` `/movies` `/series` `/books` | catálogo denso com filtros na URL |
| `/{type}/[slug]` | página do item |
| `/search?q=` | resultados entre tipos |
| `/u/[nickname]`, `/u/[nickname]/lists/[id]` | perfil público, listas |
| `/feed` | avaliações de quem você segue |
| `/pricing` | planos da API |
| `/docs/*` | guias + Scalar |
| `/login` `/signup` `/recover-password` `/reset-password` `/verify-email` | auth |
| `/profile` | dados, senha (definir/mudar), 2FA, passkeys, excluir conta |
| `/profile/api` · `/profile/api/checkout` | chaves, uso, Scalar · assinar plano/comprar créditos |
| `/contact` `/policy` `/terms` | contato, privacidade, termos e cookies |
| `/admin/*` | contato, denúncias, afiliados, edição/merge de itens, status da ingestão, banir usuário (só EN) |

**Página do item**: cabeçalho (título, ano, tipo, capa pequena, gêneros/plataformas) · metascore grande + radar hexagonal das 6 categorias + nº de avaliações · botão "ver notas dos usuários" (modal paginado, só perfis públicos) · formulário com 6 controles (slider + campo numérico, acessível por teclado, média ao vivo) · notas externas (MC, RT, IMDb, HLTB) como links com atribuição · até 3 trailers (`youtube-nocookie`, carregam só no clique) · tabela de metadados · status + adicionar à lista · "onde comprar" (livros/jogos).

**Formulários** (validação zod idêntica cliente/servidor):
- Senha: 8–32, ≥ 1 maiúscula, 1 minúscula, 1 número, 1 especial; botão olho. Rodapé do login: "Não tem conta ainda? Criar conta"; do signup: "Já tem conta? Entrar". Ambos com Google.
- Nickname: 4–16 `[a-zA-Z0-9_]`, único; avatar placeholder mostra a 1ª letra em MAIÚSCULA enquanto digita.
- Contato: nome + e-mail preenchidos e `disabled` se logado (servidor usa a sessão); select "Dúvidas" / "Sugestões & Feedback" / "Bugs e Problemas Técnicos"; textarea 5 linhas, 32–512 caracteres com contador `0/512`.
- Excluir conta: modal explicando os 30 dias, botão confirmar desabilitado por 10s com contagem.

**Preferências**: tema, consentimento e locale em cookies próprios; tema lido no servidor (sem flash, sem lib). Banner de cookies: só essenciais (Analytics é cookieless) + consentimento para embeds do YouTube.

**UI**: Base UI + Tailwind 4 (tokens por seção), `next-intl`, radar em SVG puro.

---

## 7. Pagamentos, cota e billing

**Stripe**: Checkout hospedado (`mode=subscription` com 1 Price por plano e `currency_options` BRL/USD/EUR; créditos `mode=payment`). Customer Portal para cancelar/trocar cartão/faturas. Webhook `/api/webhooks/stripe` verifica assinatura, idempotente por `event.id`, trata `checkout.session.completed`, `customer.subscription.updated|deleted`, `invoice.paid`. Moeda padrão pelo locale (pt-br→BRL, es→EUR, en→USD), trocável. Upgrade imediato (proration); downgrade no fim do período.

**AbacatePay (PIX)**: servidor cria cobrança com valor da config → QR + copia-e-cola com polling de status → webhook HMAC (tempo constante) idempotente por `provider_ref` → `current_period_end = max(agora, fim atual) + período`. Cron diário: e-mail de renovação 7d e 1d antes; vencido volta a Free.

**Cota por request (Upstash)**:
1. Resolve a chave pelo hash (cache Redis 60s: plano + `user_id`).
2. Rate limit por segundo (`@upstash/ratelimit`, janela deslizante).
3. Cota: `INCR q:{user}:life` (Free) ou `q:{user}:{yyyymm}` (pago).
4. Mensal esgotada → `DECR credits:{user}` (inicializado do ledger).
5. Sem cota e sem crédito → 402.
6. Cron diário grava em `api_usage_daily` e lança débitos no `credit_ledger`.

Fora por ora: faturas próprias, cupons, trial, PIX Automático recorrente.

---

## 8. Ingestão (`webscrappings/`)

| Fonte | Como | Fornece |
| --- | --- | --- |
| Wikidata | dump JSON (bootstrap) + SPARQL (incremental) | catálogo base, títulos por idioma, gêneros, datas, IDs externos, notas MC/RT (P444), trailer YouTube (P1651) |
| Wikipedia | REST `page/summary` | sinopse por idioma (CC BY-SA) |
| Open Library | dump + API | ISBN, páginas, autores, capas |
| Steam | `appdetails` público | preço, plataformas, trailers, capa |
| GOG | API pública de catálogo | disponibilidade, link |
| Epic, IMDb, MC, RT, HLTB, Goodreads, Skoob | sem fetch | deep-link pelo ID do Wikidata |
| TMDB, IGDB | fora da v1 | após licença comercial |

```
webscrappings/
├─ pyproject.toml   uv, ruff, pytest
├─ common/          http (rate limit, retry, User-Agent identificado), db (psycopg), upsert source_records
├─ connectors/      wikidata.py, wikipedia.py, openlibrary.py, steam.py, gog.py
│                   fetch(since) -> Iterable[Record(source, external_id, payload)]
├─ normalize.py     source_records → items / item_translations / external_scores (match por QID)
├─ covers.py        copia capas para Vercel Blob
└─ tests/           pytest com fixtures gravadas, sem rede
```

Padrão: conector grava payload bruto → `hash` igual = no-op → normalizador processa só alterados e faz upsert por `wikidata_qid` (sem QID = descartado na v1) → `POST /api/cron/revalidate` (segredo) invalida tags.

Agenda: **bootstrap** manual (`workflow_dispatch`): top 100k por sitelinks (~35k filmes, 15k séries, 25k jogos, 25k livros) · **incremental** noturno (Wikidata + Steam) · **semanal** notas externas + sinopses. Bootstrap exporta snapshot `.sql.gz` (GitHub Release) usado pelo seed de dev/preview.

Fora por ora: orquestrador (Airflow/Prefect), ORM Python.

---

## 9. Qualidade, segurança e processo

**Git**: `feat/*`/`fix/*` → PR para `dev` (sandbox: preview Vercel + branch Neon + seed sintético) → CI verde → PR `dev` → `main` (produção, Rolling Releases). Push direto bloqueado por branch protection (`gh`). Conventional Commits via commitlint + husky. release-please na `main`.

**CI (GitHub Actions)**
- Todo PR: Biome (lint + format) · `tsc --noEmit` · `bun test` (PGlite) · integração (Postgres 18 + Redis services) · diff do `openapi.json` · `next build` + smoke · `ruff` + `pytest` · `bun audit --audit-level=high` · CodeQL · Dependabot.
- PR `dev` → `main` (contra o preview, bloqueante): Playwright E2E (mobile + desktop, light/dark, 3 idiomas, axe) · Schemathesis · OWASP ZAP baseline/API · k6 curto.
- Agendado/manual: agente QA exploratório (dirige o browser por todas as rotas com edge cases: unicode, limites, payloads XSS) → relatório em `qa/reports/` + issue por bug.

**Segurança**: `docs/security.md` com tabela OWASP Top 10:2025 (A01–A10: superfícies, mitigação, commit, riscos aceitos), revisada a cada milestone.

**Agentes**: `AGENTS.md` é a fonte; `CLAUDE.md` importa `@AGENTS.md`. `.claude/rules/`: `karpathy.md`, `ponytail.md`, `graphify.md`. Skills `respondeae-*` reescritas como `hexascore-secure-endpoint` e `hexascore-local-verification`. Worktrees + agentes paralelos para `webscrappings/`, `sdks/` e `docs/`.

**Dependências**: só versões estáveis, fixadas sem `^`. Imagens Docker com tag fixa (`oven/bun:<versão>`, `postgres:18.x`). Nenhuma dependência nova sem aprovação.

---

## 10. Milestones

Horizontal: cada milestone cobre as 4 seções. Ordem fixa; cada uma fecha com uma release.

### 0.0.x — Fundação
- [ ] `git init`, `.gitignore`, remote `origin`, branches `main` + `dev`, branch protection
- [ ] `AGENTS.md`, `CLAUDE.md`, `.claude/rules/{karpathy,ponytail,graphify}.md`, reescrever skills `respondeae-*`
- [ ] Bun workspaces, `apps/web` Next 16 + Tailwind 4 + Base UI, Biome 2, tsconfig estrito
- [ ] Drizzle + client Neon/PGlite, `infra/docker-compose.yaml` (Postgres 18 + Redis)
- [ ] `next-intl` (en/pt-br/es), tema via cookie, layout base, navbar, footer, banner de cookies
- [ ] CI básico (lint, typecheck, test, build), husky, commitlint, release-please, Dependabot, CodeQL
- [ ] Projeto Vercel + Neon + Upstash via Marketplace, `vercel env pull`
- **Verificar**: `bun run lint:ci && bunx tsc --noEmit && bun test && bun run build` verdes local e no CI; preview da `dev` abre `/`, `/pt-br`, `/es` com tema persistido.

### 0.1 — Catálogo
- [ ] Schema do catálogo + `source_records` + migrations
- [ ] `webscrappings/`: common, conectores Wikidata/Wikipedia/Open Library/Steam/GOG, normalizador, covers, testes
- [ ] Workflows bootstrap/incremental/semanal + snapshot `.sql.gz`
- [ ] Páginas `/`, catálogos das 4 seções, página do item (sem avaliação ainda), `/search`, autocomplete
- **Verificar**: bootstrap local carrega ≥ 100k itens; `pytest` verde; busca "zelda" (com typo "zleda") retorna o jogo; página do item renderiza nos 3 idiomas com atribuição de licença.

### 0.2 — Auth e conta
- [ ] Better Auth (e-mail+senha, Google, verificação, 2FA, passkeys, admin), BotID
- [ ] `/login`, `/signup`, `/recover-password`, `/reset-password`, `/verify-email`
- [ ] `/profile` (dados, definir/mudar senha, 2FA, passkeys, exclusão com modal de 10s) + cron de anonimização
- [ ] `/contact`, `/policy`, `/terms`, templates react-email
- **Verificar**: testes de regras de senha/nickname (casos válidos e limites 7/8/32/33); E2E de signup → verificação → login → 2FA → exclusão → cancelamento.

### 0.3 — Avaliação e social
- [ ] Ratings + stats na transação, metascore com limiar 5, radar hexagonal, modal de notas
- [ ] Status por item, follows, `/feed`, listas curadas, perfil público/privado, denúncias
- [ ] `/admin` (contato, denúncias, afiliados, edição/merge, status da ingestão, banir)
- [ ] Afiliados na página do item
- **Verificar**: testes do cálculo (média, arredondamento, limiar 4→5 avaliações, edição, exclusão de conta); E2E avaliar → metascore muda.

### 0.4 — API
- [ ] API keys, middleware de auth/rate limit/cota (Upstash), erros RFC 9457, headers
- [ ] Endpoints `/v1/*`, `openapi.json` gerado + check no CI
- [ ] `/docs` + Scalar, `/profile/api` com dashboard de uso e últimas requests
- [ ] SDKs TS e Python gerados + publicação via release-please
- **Verificar**: Free bloqueia na 101ª request (402) e `limit=11` é rejeitado; Schemathesis sem falhas; quickstart das docs funciona copiando e colando.

### 0.5 — Billing
- [ ] Config de planos, Products/Prices Stripe (test mode via CLI), Checkout, Portal, webhook
- [ ] AbacatePay PIX (cobrança, polling, webhook), créditos, cron de renovação/expiração
- [ ] `/pricing`, `/profile/api/checkout`
- **Verificar**: `stripe trigger` e simulação PIX ativam o plano; reenvio do mesmo evento não duplica; valor adulterado no client é ignorado.

### 0.6 — Hardening
- [ ] Seed 10k usuários sintéticos com notas realistas
- [ ] Playwright E2E completo + axe, agente QA, Schemathesis, ZAP, k6 no gate `dev` → `main`
- [ ] Auditoria OWASP em `docs/security.md`, CSP, Firewall, Flags
- [ ] Performance (Core Web Vitals no Speed Insights), a11y
- **Verificar**: gate `dev` → `main` todo verde; ZAP sem alertas High; p95 da API < 300 ms no k6.

### 1.0.0 — Launch
- [ ] Nome/domínio finais, `APP_NAME`, i18n revisado
- [ ] Termos/política revisados, avisos de afiliado e licenças
- [ ] Bootstrap de 100k itens em produção, Stripe live, AbacatePay live
- **Verificar**: smoke em produção; checkout real de R$ mínimo + estorno; release `v1.0.0` publicada.

---

## Pendências antes do launch

- [ ] **Nome final** (codinome `hexascore`). `.com` livres consultados em 2026-10-02: hexascore.com, scorara.com, averago.com, kultscore.com, talescore.com, crowdcrit.com, hexmeter.com. Checar marca registrada (INPI/USPTO/EUIPO) antes de comprar.
- [ ] Revisar preços dos planos.
- [ ] Licenças comerciais TMDB/IGDB (se quiser enriquecer além de Wikidata).
- [ ] Revisão jurídica de termos/privacidade (LGPD/GDPR).
