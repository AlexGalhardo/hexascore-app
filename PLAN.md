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
| API | Dentro do Next (`app/api/v1`), `api.hexascore.app` → rewrite para `/api/v1` | 1 projeto Vercel |
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

## 3. Documentação (`docs/`)

Toda documentação de referência vive em `docs/`; este plano só guarda decisões e milestones. Ao mudar um subsistema, atualize o doc correspondente no mesmo PR.

| Doc | Conteúdo |
| --- | --- |
| [`docs/architecture.md`](docs/architecture.md) | monorepo, fluxo de dados, regras de segurança obrigatórias |
| [`docs/data-model.md`](docs/data-model.md) | tabelas Drizzle/Postgres |
| [`docs/api.md`](docs/api.md) | API pública `/v1`, contrato, docs e SDKs |
| [`docs/frontend.md`](docs/frontend.md) | rotas, página do item, formulários, preferências |
| [`docs/design-system.md`](docs/design-system.md) | direção visual, tokens, tipografia, componentes |
| [`docs/payments.md`](docs/payments.md) | Stripe, AbacatePay, cota e créditos |
| [`docs/ingestion.md`](docs/ingestion.md) | fontes, conectores Python, agenda |
| [`docs/security.md`](docs/security.md) | OWASP Top 10:2025, riscos aceitos |
| [`docs/testing.md`](docs/testing.md) | camadas de teste e QA/pentest, como rodar |
| [`docs/process.md`](docs/process.md) | git, CI/CD, releases, regras de agentes |

---

## 4. Milestones

Horizontal: cada milestone cobre as 4 seções. Ordem fixa; cada uma fecha com uma release.

### 0.0.x — Fundação
- [x] `git init`, `.gitignore`, remote `origin`, branches `main` + `dev` (branch protection: desligada por decisão do dono; ruleset pronto em `.github/rulesets/`)
- [x] `AGENTS.md`, `CLAUDE.md`, `.claude/rules/{karpathy,ponytail,graphify}.md`, reescrever skills `respondeae-*`
- [x] Bun workspaces, `apps/web` Next 16 + Tailwind 4 + Base UI, Biome 2, tsconfig estrito
- [x] Drizzle + client Neon/PGlite, `infra/docker-compose.yaml` (Postgres 18 + Redis)
- [x] `next-intl` (en/pt-br/es), tema via cookie, layout base, navbar, footer, banner de cookies
- [x] CI básico (lint, typecheck, test, build), husky, commitlint, release-please, Dependabot, CodeQL
- [x] Projeto Vercel + Neon + Upstash via Marketplace, `vercel env pull`
- **Verificar**: `bun run lint:ci && bunx tsc --noEmit && bun test && bun run build` verdes local e no CI; preview da `dev` abre `/`, `/pt-br`, `/es` com tema persistido.

### 0.1 — Catálogo
- [x] Schema do catálogo + `source_records` + migrations
- [x] `webscrappings/`: common, conectores Wikidata/Wikipedia/Open Library/Steam/GOG, normalizador, covers, testes
- [x] Workflows bootstrap/incremental/semanal + snapshot `.sql.gz`
- [x] Páginas `/`, catálogos das 4 seções, página do item (sem avaliação ainda), `/search`, autocomplete
- **Verificar**: bootstrap local carrega ≥ 100k itens; `pytest` verde; busca "zelda" (com typo "zleda") retorna o jogo; página do item renderiza nos 3 idiomas com atribuição de licença.

### 0.2 — Auth e conta
- [x] Better Auth (e-mail+senha, Google, verificação, 2FA, passkeys, admin), BotID
- [x] `/login`, `/signup`, `/recover-password`, `/reset-password`, `/verify-email`
- [x] `/profile` (dados, definir/mudar senha, 2FA, passkeys, exclusão com modal de 10s) + cron de anonimização
- [x] `/contact`, `/policy`, `/terms`, templates react-email
- **Verificar**: testes de regras de senha/nickname (casos válidos e limites 7/8/32/33); E2E de signup → verificação → login → 2FA → exclusão → cancelamento.

### 0.3 — Avaliação e social
- [x] Ratings + stats na transação, metascore com limiar 5, radar hexagonal, modal de notas
- [x] Status por item, follows, `/feed`, listas curadas, perfil público/privado, denúncias
- [x] `/admin` (contato, denúncias, afiliados, edição/merge, status da ingestão, banir)
- [x] Afiliados na página do item
- **Verificar**: testes do cálculo (média, arredondamento, limiar 4→5 avaliações, edição, exclusão de conta); E2E avaliar → metascore muda.

### 0.4 — API
- [x] API keys, middleware de auth/rate limit/cota (Upstash), erros RFC 9457, headers
- [x] Endpoints `/v1/*`, `openapi.json` gerado + check no CI
- [x] `/docs` + Scalar, `/profile/api` com dashboard de uso (lista de últimas requests adiada: não há log por request)
- [ ] SDKs TS e Python gerados + publicação via release-please (gerados em `sdks/`; publicação pendente de credenciais npm/PyPI)
- **Verificar**: Free bloqueia na 101ª request (402) e `limit=11` é rejeitado; Schemathesis sem falhas; quickstart das docs funciona copiando e colando.

### 0.5 — Billing
- [x] Config de planos, Products/Prices Stripe (test mode via CLI), Checkout, Portal, webhook
- [x] AbacatePay PIX (cobrança, polling, webhook), créditos, cron de renovação/expiração
- [x] `/pricing`, `/profile/api/checkout`
- **Verificar**: `stripe trigger` e simulação PIX ativam o plano; reenvio do mesmo evento não duplica; valor adulterado no client é ignorado.

### 0.6 — Hardening
- [x] Seed 10k usuários sintéticos com notas realistas (`unnest` colunar, ~7 min para 10k/100k/2,2M no Docker; `docs/testing.md`)
- [~] Playwright E2E completo + axe (incluindo specs do 0.3), k6 (p95 + 429/402), Schemathesis e OWASP ZAP (falha em High) no gate `dev` → `main` (`qa/e2e`, `qa/k6`, `qa/schemathesis`, `qa/zap`, `qa.yml`) — feitos; faltam secrets `QA_API_KEY` (Pro/Business) e `QA_FREE_API_KEY` no repo e o agente QA
- [~] Auditoria OWASP em `docs/security.md`, CSP com nonce, Flags (`pix-checkout`, `public-api-signup`) — feitos; Firewall documentado, falta aplicar no dashboard
- [~] Performance (Core Web Vitals no Speed Insights), a11y — Analytics + Speed Insights no layout, axe em todas as páginas públicas; leitura dos CWV pendente (precisa de tráfego)
- **Verificar**: gate `dev` → `main` todo verde; ZAP sem alertas High; p95 da API < 300 ms no k6.

### 1.0.0 — Launch
- [ ] Nome/domínio finais, `APP_NAME`
- [x] i18n revisado (paridade de chaves e parágrafos legais testada, sem codinome nos textos, e-mails de auth traduzidos)
- [x] Termos/política revisados, avisos de afiliado e licenças (crédito de capa na página do item; revisão por advogado segue em Pendências)
- [ ] Bootstrap de 100k itens em produção, Stripe live, AbacatePay live
- **Verificar**: smoke em produção; checkout real de R$ mínimo + estorno; release `v1.0.0` publicada.

---

## Pendências antes do launch

Ações do dono (não automatizáveis):

- [ ] Nome/domínio final e domínio verificado no Resend (domínio padrão `hexascore.app` comprado na Vercel em 2026-10-03, com `api.` e `www.` → apex; falta verificar no Resend e trocar `EMAIL_FROM`) (hoje `onboarding@resend.dev` só entrega para o dono da conta).
- [ ] Chaves live do Stripe e da AbacatePay (hoje sandbox) e teste real de checkout + estorno.
- [ ] Credenciais npm/PyPI para publicar os SDKs.
- [ ] Confirmar a carga completa de 100k itens (limite de armazenamento do Neon free; avaliar plano pago).

- [ ] **Nome final** (codinome `hexascore`). `.com` livres consultados em 2026-10-02: hexascore.com, scorara.com, averago.com, kultscore.com, talescore.com, crowdcrit.com, hexmeter.com. Checar marca registrada (INPI/USPTO/EUIPO) antes de comprar.
- [ ] Revisar preços dos planos.
- [ ] Licenças comerciais TMDB/IGDB (se quiser enriquecer além de Wikidata).
- [ ] Revisão jurídica de termos/privacidade (LGPD/GDPR).
