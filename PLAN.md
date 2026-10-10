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

Perfil público `/u/{nickname}` (opção privado) · status por item (quero / em andamento / concluído / abandonado) · seguir usuários + feed · listas curadas públicas (uma categoria por lista, até 10 por categoria) · denúncias + moderação no `/admin`.

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
| Fontes | **APIs + Wikidata, sem scraping**. IMDb/MC/RT/HLTB/Goodreads/Skoob/Epic = só deep-link via ID do Wikidata. TMDB (filmes, séries) e IGDB (jogos) são buscados desde 2026-10-10 para capa oficial e campos que o Wikidata não tem, com atribuição em `meta.coverAttribution`; sinopses deles nunca são copiadas | Decisão do dono em 2026-10-10: seguir com TMDB/IGDB também em produção e na API paga, assumindo o risco de licença (ambos restringem uso comercial sem acordo; ver pendências do launch) |
| Sequência | **Horizontal**: as 4 seções avançam juntas, camada por camada | Escolha do dono |
| Monorepo | **Bun workspaces** (`apps/web`, `sdks/ts`), **sem Turborepo** | Só 1 consumidor TS real; Turborepo entra com ≥ 3 pacotes JS com dependência de build |
| API | Dentro do Next (`app/api/v1`), `api.hexascore.app` → rewrite para `/api/v1` | 1 projeto Vercel |
| Banco | Postgres-only. Prod: **Neon** (branch por preview). Dev/testes: **PGlite**. Paridade/seed/QA: **Docker Postgres 18**. SQLite removido | 1 schema Drizzle, sem dupla manutenção |
| Busca | Postgres nativo: `pg_trgm` + `unaccent` + `tsvector` por idioma | Suficiente para 100k–1M itens, zero infra |
| Rate limit/cota | **Upstash Redis** + flush diário para Postgres | Não sobrecarregar o Neon com 1 write/request |
| Ingestão | Python 3.13 + uv em `webscrappings/`, rodando em **GitHub Actions cron**, staging `source_records` + normalizador | Custo zero, sem limite de 300s |
| i18n UI | `next-intl`, EN (padrão, sem prefixo) / `/pt-br` / `/es` | — |
| Fetch no cliente | Server Components por padrão; **TanStack Query** (`@tanstack/react-query`) só para fetch disparado no browser (autocomplete, paginação de modal, polling) | Cache, cancelamento e polling sem `useEffect` manual |
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
| [`docs/sdk-publish-npm.md`](docs/sdk-publish-npm.md) · [`sdk-publish-jsr.md`](docs/sdk-publish-jsr.md) · [`sdk-publish-pypi.md`](docs/sdk-publish-pypi.md) | Passo a passo para publicar os SDKs no npm, no JSR e no PyPI |
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
- [x] SDKs prontos para publicar, versão `0.1.0`: `@hexascore/sdk` (npm e JSR, `sdks/ts`) e `hexascore` (PyPI, `sdks/python`), com testes, licença MIT e guias `docs/sdk-publish-npm.md`, `docs/sdk-publish-jsr.md`, `docs/sdk-publish-pypi.md`. Conferido em 2026-10-10: `npm pack --dry-run`, `jsr publish --dry-run`, `uv build` + `twine check` + instalação do wheel em ambiente limpo.
- [ ] Publicar os SDKs (manual, pelo dono): criar organização `hexascore` no npm, escopo `@hexascore` no JSR e contas PyPI/TestPyPI; seguir os três guias. Nomes livres em 2026-10-10.
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
- [x] Tempo para zerar colaborativo (jogos, substitui o scraping do HowLongToBeat): relatos de quem marcou "Jogado", mediana por estilo (≥ 3 relatos) na página e em `ItemDetail.time_to_beat` na API
- **Verificar**: smoke em produção; checkout real de R$ mínimo + estorno; release `v1.0.0` publicada.

---

## Próxima sessão (ingestão em produção `hexascore.app`, estado em 2026-10-04 20:15 local)

O branch sandbox `ep-falling-sky` se apagou sozinho (expiração de 1 dia configurada pelo dono); a ingestão passou a rodar **localmente contra produção** (`ep-blue-hall`). GitHub Actions segue parado por falha de cobrança. Comando de cada etapa em `webscrappings/`:
`uv run --env-file ../apps/web/.env.production.local --env-file .env.local python -m <módulo> [--missing-only]` (`.env.production.local` vem de `vercel env pull --environment production`; os segredos chegam como `[SENSITIVE]`, só `DATABASE_URL` é usado). Tudo é retomável.

Estado medido no banco de produção (99.960 itens):

- Wikidata 99.960, Wikipedia en 90.569 / es 45.837 / pt 35.107, Open Library 10.187, Steam 6.049, GOG 1.974; `normalize` com 0 pendentes.
- Capas WebP (capa + miniatura, HTTP 200 conferido por tipo): **27.469** — filmes 6.902, séries 1.006, livros 12.711, jogos 6.850. Filmes e séries estão no teto das fontes (só itens com imagem no Commons, P18/P3383; pôsteres costumam ter copyright); 2 itens falharam (503 do Internet Archive) e entram na próxima execução de `covers`.
- Seed de demonstração: 10.000 usuários `seed-user-*`, 2.145.882 avaliações + status/follows/listas (`docs/testing.md` → "Produção"); remover com `--purge` antes do launch.

Pendências:

- [x] **Store do Vercel Blob público**: o store privado `hexascore-blob` foi apagado e trocado pelo público `hexascore-covers` (`store_a9K51nn0eemlqu9Z`, iad1, `a9k51nn0eemlqu9z.public.blob.vercel-storage.com`), ligado ao projeto em Production/Preview/Development; `webscrappings/.env.local` atualizado; `covers --limit 3` no sandbox gravou 3 capas + miniaturas (HTTP 200). CSP `img-src https:` e sem `remotePatterns`: nada a mudar no app.
- [ ] Atualizar o secret `BLOB_READ_WRITE_TOKEN` do GitHub com o token do store novo (`gh secret set BLOB_READ_WRITE_TOKEN`) antes de reativar os workflows de ingestão.
- [x] `normalize` e `covers` completos em produção (26.758 capas gravadas nesta rodada, ~3,5 itens/s; a seleção do `covers` agora traz só as chaves usadas de cada fonte).
- [ ] Rodar de novo `connectors.steam --missing-only` e depois `normalize` + `covers` para os jogos que faltarem.
- [ ] Regularizar a cobrança do GitHub Actions e reativar os workflows (`gh workflow enable <nome>`), ou manter a ingestão local e documentar isso em `docs/ingestion.md`.
- [ ] Confirmar para qual banco aponta o secret `DATABASE_URL` dos workflows de ingestão (sandbox ou produção) antes de reativá-los.
- [ ] Remover as worktrees antigas em `.claude/worktrees/` (todas mescladas e limpas; `git worktree remove` foi bloqueado pelo modo automático).
- [x] Produção (`hexascore.app`): ingestão completa (enriquecimento, `normalize`, `covers`) e seed de demonstração.
- [ ] `POST /api/cron/revalidate` `{"tags":["catalog"]}` com o `CRON_SECRET` (dono) para o site refletir capas e notas novas.
- [ ] Deploy das mudanças do seed/flag e `FLAG_SEED_DATA=on` na Vercel; `seed.ts --purge --i-know-this-is-production` + flag off antes do launch.
- [x] Fonte de pôsteres para filmes/séries: conector TMDB (`connectors/tmdb.py`), capa oficial com prioridade sobre o Commons.
- [x] Rodada TMDB em produção (2026-10-10): 48.485 registros; capa em 34.657 de 34.968 filmes (99%) e 13.300 de 14.911 séries (89%), antes 6.902 e 1.006.
- [x] Rodada IGDB em produção (2026-10-10): 22.848 registros; capa em 22.799 de 24.873 jogos (92%), antes 6.850 (credenciais em `docs/igdb-credentials.md`).
- [x] Rodada Open Library em produção (2026-10-10): 2.702 registros refeitos. Livros ficaram com 8.148 capas de 24.089 (34%): as 4.635 capas vindas do P18 do Commons foram removidas e a Open Library não tem imagem para esses livros.
- [ ] Capas de livros: 15.941 livros sem capa. Avaliar outra fonte de capa oficial (ex.: Google Books) ou busca na Open Library por título + autor.
- [ ] Sinopses próprias onde não há Wikipedia: escrever o passo que gera o registro `synopsis` (LLM a partir do `overview` de TMDB/IGDB/Open Library); depende de chave de API, modelo e teto de custo. O `normalize` já lê esse registro. Alvo medido em 2026-10-10: 3.021 itens sem sinopse em inglês que têm `overview` (2.121 jogos, 475 séries, 258 livros, 167 filmes); outros 5.251 não têm texto em fonte nenhuma.
- [ ] Aviso de atribuição do TMDB no site ("This product uses the TMDB API but is not endorsed or certified by TMDB") e incluir os conectores novos nos workflows de ingestão.

## Pendências antes do launch

Ações do dono (não automatizáveis):

- [x] Domínio verificado no Resend: `hexascore.app` (comprado na Vercel em 2026-10-03, com `api.` e `www.` → apex) está `verified` com envio ativo (us-east-1), `EMAIL_FROM` = `HexaScore <no-reply@hexascore.app>`. Conta Resend: `aleexgalhardoo@gmail.com`.
- [ ] Chaves live do Stripe e da AbacatePay (hoje sandbox) e teste real de checkout + estorno.
- [ ] Contas e escopos para publicar os SDKs: organização `hexascore` no npm (com 2FA), escopo `@hexascore` no JSR (login pelo GitHub), contas PyPI e TestPyPI (com 2FA e token). Passo a passo em `docs/sdk-publish-*.md`.
- [ ] Confirmar a carga completa de 100k itens (limite de armazenamento do Neon free; avaliar plano pago).

- [ ] **Nome final** (codinome `hexascore`). `.com` livres consultados em 2026-10-02: hexascore.com, scorara.com, averago.com, kultscore.com, talescore.com, crowdcrit.com, hexmeter.com. Checar marca registrada (INPI/USPTO/EUIPO) antes de comprar.
- [ ] Revisar preços dos planos.
- [ ] Licenças comerciais TMDB/IGDB: os dois já são usados em produção e na API paga (decisão de 2026-10-10); fechar o acordo comercial com cada um antes do launch.
- [ ] Revisão jurídica de termos/privacidade (LGPD/GDPR).
