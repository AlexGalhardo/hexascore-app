# Segurança — OWASP Top 10:2025

Auditoria viva. Revisar a cada milestone; cada mitigação aponta para o código/commit que a implementa quando existir.

## Superfícies

| Superfície | Entrada não confiável | Proteções |
| --- | --- | --- |
| Server Actions (`actions/`) | form data | sessão via `getSessionUser()`, zod, BotID nos fluxos públicos |
| `actions/account.ts` (nome, definir senha, perfil privado, agendar/cancelar exclusão) | argumentos da action | exige sessão; alvo sempre é o usuário da sessão; zod (`nameSchema`, `passwordSchema`, `z.boolean()`); `setPassword` só no servidor; exclusão faz sign-out e apaga todas as sessões |
| `actions/contact.ts` | form data | BotID; usuário logado tem nome/e-mail tirados da sessão (body ignorado); `contactSchema`; rate limit 5/h por usuário ou IP na tabela `rate_limit` |
| Better Auth `/api/auth/*` | JSON | origin check, rate limit em banco (429), hook de senha em todo endpoint que aceita senha, BotID em `/sign-up/email`, `/sign-in/email`, `/sign-in/username`; verificação de e-mail obrigatória; recuperação e reenvio sempre respondem igual (sem enumeração) |
| `/api/cron/purge-accounts` | header `Authorization` | `Bearer $CRON_SECRET` comparado em tempo constante (sha256 + `timingSafeEqual`); anonimiza só contas com `deletion_scheduled_at` < agora − 30 dias; revoga API keys e cancela assinatura Stripe ativa antes (falha no Stripe mantém a conta agendada para o próximo dia) |
| API pública `/api/v1` (e `api.<domínio>/v1` via proxy) | query string, params de path, header `Authorization` | chave só por header (nunca query), hash sha256 + lookup por hash, chave revogada/conta banida → 401/403; `z.strictObject` em params/query (desconhecido → 400), `limit` limitado pelo plano; throttle + cota (Upstash) **fail closed** (503 se o store falhar); corpo passa por `response.parse` (nenhum campo fora do contrato, `external_scores` nunca sai — teste); só leitura, `no-store`; erros RFC 9457 sem stack |
| Página `/profile/api` + `actions/api-keys.ts` | nome da chave, id para revogar | sessão via `getSessionUser()` (redirect para login), e-mail verificado para criar, máx. 5 chaves ativas, revogação filtrada por dono; DTO `ApiKeyView` sem hash; chave completa só na resposta da criação e no estado do React (Scalar com `persistAuth: false`, sem IA/MCP de terceiros) |
| Webhook Stripe `POST /api/webhooks/stripe` | body bruto + `Stripe-Signature` | `constructEventAsync` (tolerância de 5 min), 400 em assinatura inválida, idempotência por `event.id` em `processed_events` na mesma transação dos efeitos, estado da assinatura relido da API (eventos fora de ordem), crédito com `ref` único = id da Checkout Session; 503 com log se `STRIPE_WEBHOOK_SECRET` faltar |
| Webhook AbacatePay `POST /api/webhooks/abacatepay` | body bruto, `?webhookSecret=`, `X-Webhook-Signature` | secret da URL e HMAC-SHA256 base64 (chave pública da AbacatePay) comparados em tempo constante → 401; zod só nos campos lidos; idempotência por `id` do evento + transição `pending → paid`; valor pago tem de bater com `amount_cents` gravado |
| Server Actions `actions/billing.ts` | `{ product, currency }` / `{ purpose }` | sessão + e-mail verificado; zod; preço por `lookup_key` (Stripe) ou catálogo (PIX), nunca do client; `client_reference_id`/metadata = id da sessão; regras de conflito Stripe×PIX no servidor; QR pendente reaproveitado (sem spam de cobranças) |
| `GET /api/billing/pix/[id]` | `id` | sessão, zod `uuid`, filtro por dono (404 para os demais), `no-store`; reconsulta a AbacatePay com timeout |
| Páginas `/pricing`, `/profile/api/checkout` | query `currency`/`plan`/`status` | allowlist via zod; DTO `BillingSummary` sem ids de provedor |
| Cron `/api/cron/*` (`flush-usage`, `billing-reminders`) | header | `Authorization: Bearer $CRON_SECRET` (tempo constante); lembrete gravado em `processed_events` antes do envio (sem e-mail duplicado) |
| `actions/ratings.ts` (`saveRating`, `deleteRating`) | `{ itemId, scores[6] }` / `itemId` | zod (uuid, 6 inteiros 0–100); `asMember()` = sessão + não banido + e-mail verificado + nickname; rate limit 30/min por usuário; alvo sempre o usuário da sessão; stats recalculadas no servidor numa transação com lock na linha do item (sem lost update); CHECK 0..100 no banco |
| `actions/item-status.ts`, `actions/follows.ts`, `actions/lists.ts` | ids, status, nickname, título/descrição | zod; `asMember()`; rate limit (60/min, 30/min, 60/min); escrita filtrada por dono (`user_id` da sessão) — lista de outro usuário = `notFound`; limites 50 listas / 500 itens / título 80 / descrição 500 no servidor (cap de listas com advisory lock); não segue a si mesmo (zod + CHECK) nem perfil privado/banido |
| `actions/reports.ts` | `{ target, reason, details }` | zod (alvo lista=uuid ou usuário=nickname, reason enum, details ≤ 500); `asMember()`; 10/h; denúncia aberta duplicada ignorada (índice único parcial); não denuncia a si/à própria lista |
| `actions/account.ts#setNickname` (`/onboarding`) | nickname | sessão; só quando `username` é nulo; `nicknameSchema` + validação/unicidade do plugin username |
| Páginas `/u/[nickname]`, `/u/[nickname]/lists/[id]`, `/feed` | params/query | zod com `.catch` para `tab/sort/page/cursor`; perfil privado → página mínima sem dados (200), banido/inexistente → 404; listas privadas/ocultas/de perfil privado → 404 para terceiros; feed só inclui perfis públicos; DTOs sem e-mail nem id de usuário |
| `/admin/*` + `actions/admin.ts` | form data | `isAdmin()` (role `admin`, não banido) no layout, em cada página e em cada action → 404/`notFound` para os demais; zod em todo form; ban via `auth.api.banUser` (revoga sessões); afiliados só `https:`, sem credenciais/porta, host no allowlist da loja; merge em transação com lock nos dois itens e mesmo tipo |
| `scripts/admin-grant.ts` | e-mail (CLI) | roda só com acesso ao banco; zod `email` |
| Ingestão (Python) | payloads de terceiros | escrita só em `source_records`; normalizador valida antes do upsert; texto renderizado sempre escapado |

## A01–A10 (auditoria 0.6, 2026-10-03)

Evidência = arquivo:linha em `apps/web/` (salvo indicação), teste que prova e commit. Status: ✅ implementado e testado · ⚠️ parcial (ver lacunas).

| Risco | Mitigação | Evidência | Status |
| --- | --- | --- | --- |
| A01 Broken Access Control | identidade só da sessão; dono em toda escrita; `/admin` exige `role=admin` (404 caso contrário); perfis privados filtrados no repositório | `lib/session.ts:20` (`getSessionUser`), `app/[locale]/admin/guard.ts:5` (`requireAdmin`), `lib/social/repository.ts:43` (`profile_private = false`); revogação de chave filtrada por dono em `actions/api-keys.ts`; E2E `qa/e2e/specs/profile.spec.ts` "profile requires login", `pricing.spec.ts` "checkout requires login" | ✅ |
| A02 Security Misconfiguration | CSP por request com nonce + `strict-dynamic` (sem `unsafe-inline`/`unsafe-eval` em script), `frame-ancestors 'none'`, iframes só `youtube-nocookie.com`, HSTS 2 anos + preload, `nosniff`, `Referrer-Policy: strict-origin-when-cross-origin`, `Permissions-Policy` restritiva, COOP `same-origin`; `/api/*` com `default-src 'none'`; `poweredByHeader: false` | `lib/security-headers.ts`, `proxy.ts`, `next.config.ts` (commit `8a35c7a`); testes `lib/security-headers.test.ts` (7) e E2E `qa/e2e/specs/headers.spec.ts` (página + rota de API; todo `<script>` carrega o nonce); toda spec E2E falha em violação de CSP no console (`qa/e2e/helpers/fixtures.ts`) | ✅ (ZAP em `qa/zap`, ver abaixo) |
| A03 Software Supply Chain Failures | versões fixadas sem `^`, lockfile, `bun audit --audit-level=high` no CI, override `js-yaml` 4.3.2, Dependabot, CodeQL; actions fixadas por major | `.github/workflows/ci.yml` job `audit`; `package.json` `overrides`; `bun audit --audit-level=high` local em 2026-10-03: sem achados high | ✅ |
| A04 Cryptographic Failures | hash de senha do Better Auth; API keys só como sha256; comparações de segredo em tempo constante; segredos nunca `NEXT_PUBLIC_*`; HSTS | `lib/api/keys.ts:15`, `lib/cron-auth.ts:1` (`timingSafeEqual`), `lib/env.ts` (zod; segredos obrigatórios em produção) | ✅ |
| A05 Injection | Drizzle parametrizado; arrays do seed como parâmetro único; React escapa saída; caracteres de controle nunca chegam ao Postgres | `lib/catalog/search-params.ts` (rejeita `\p{Cc}`), `lib/catalog/repository.ts` `searchItems` (remove `\p{Cc}`) — commit `566c919`, testes `search-params.test.ts` e `repository.test.ts` "control characters (NUL)…"; E2E `search.spec.ts` (XSS refletido não executa, `' OR 1=1 --`, NUL, 500 chars) e `catalog.spec.ts` "a tampered cursor or filter never breaks the page" | ✅ |
| A06 Insecure Design | dinheiro e cota calculados no servidor; limiar de 5 avaliações aplicado na leitura; cota Free amarrada a e-mail verificado; kill switches `pix-checkout`/`public-api-signup` | `lib/billing/catalog.ts`, `lib/api/quota.ts`; `lib/flags.ts` + `actions/billing.ts`/`actions/api-keys.ts` (commit `5a91c5b`) | ✅ |
| A07 Authentication Failures | regras de senha em todo endpoint com senha, verificação de e-mail obrigatória, 2FA TOTP, passkeys, rate limit em banco (429), BotID nos fluxos públicos | `lib/auth.ts:86-95` (hook `passwordSchema`), `lib/bot.ts:10`; E2E `auth.spec.ts` (signup → link → login; regras no client; senha errada genérica; login com TOTP e código inválido) e `profile.spec.ts` (troca de senha, exclusão com contagem de 10 s e cancelamento) | ✅ |
| A08 Software or Data Integrity Failures | webhooks assinados (Stripe `constructEventAsync`, AbacatePay HMAC) e idempotentes em `processed_events` na mesma transação; migrations só novas; release-please com tags | `lib/billing/stripe-webhook.ts:127`; testes `lib/billing/*.test.ts` | ✅ |
| A09 Logging & Alerting Failures | `x-request-id` em toda resposta da API; logs estruturados com o id; Vercel Observability; Telegram para falhas de cron/webhook | `lib/api/handler.ts:88,175` | ⚠️ alertas de pico de 4xx/5xx ainda manuais |
| A10 Mishandling of Exceptional Conditions | RFC 9457 sem stack; store de cota fail closed (503); BotID nega em erro na Vercel; locale inválido vira 404 antes de consultar (antes: ZodError logado a cada `/favicon.ico`) | `lib/api/handler.ts:102-105`, `i18n/locale.ts` (`asLocale`, commit `7362487`); E2E `not-found.spec.ts` (404 localizados + problemas da API) | ✅ |

### OWASP ZAP (`qa/zap`, gate `qa.yml`)

Execução local em 2026-10-03 (ZAP 2.17.0, `next start` + seed pequeno, chave Pro; spider 2.225 URLs, active scan de `/api/v1` em 9 min 41 s): **0 High**. Medium: CSP `style-src 'unsafe-inline'` e "wildcard directive" (aceitos: estilos inline do Next/Tailwind; scripts seguem só por nonce), CSP ausente em `/robots.txt` e `/sitemap.xml` (não-HTML). Low: cookie `locale` sem `HttpOnly` (lido no client por design), timestamps. Corrigido: formulários com senha ou código 2FA (`/login`, `/signup`, senha, 2FA, reset) usam `method="post"`, então um submit antes da hidratação nunca põe o segredo na URL.

### Vercel Firewall (configuração a aplicar no dashboard)

As ferramentas disponíveis nesta sessão não aplicaram a configuração; aplicar em **Project → Firewall → Configure** (ou `vercel firewall`/API `PUT /v1/security/firewall/config`):

| Regra | Condição | Ação |
| --- | --- | --- |
| `auth-rate-limit` | `path` começa com `/api/auth/` | rate limit fixed window 60 s, 30 req por IP (`ip`), ação `deny` 429 por 5 min |
| `api-v1-rate-limit` | `path` começa com `/api/v1/` **ou** `host` começa com `api.` | rate limit fixed window 60 s, 600 req por IP — teto anti-abuso acima do throttle por chave (Upstash) |
| Bot Protection (managed) | — | `challenge` |
| AI Bots (managed) | — | `deny` |
| OWASP core ruleset | sqli, xss, rce, lfi, rfi | `log` por 7 dias, depois `deny` |

O bypass de automação (`x-vercel-protection-bypass`) usado pelo CI não contorna o Firewall; regras com `deny` em `/api/auth/*` podem afetar a suíte E2E contra o preview — rode o gate com no máximo 2 workers (padrão no CI).

## Riscos aceitos

| Risco | Motivo |
| --- | --- |
| Contas múltiplas para multiplicar a cota Free | limitado por e-mail verificado + BotID; 100 req não compensa o abuso |
| Signup revela "nickname já em uso" | nickname é público (`/u/{nickname}`); e-mail duplicado recebe a mesma resposta genérica de sucesso (Better Auth com `requireEmailVerification`) |
| BotID não roda fora da Vercel | `lib/bot.ts` só faz bypass quando `VERCEL !== "1"`; na Vercel qualquer falha da checagem nega (fail closed) |
| `style-src 'unsafe-inline'` na CSP | atributos `style` do React/Base UI e os `<style>` em runtime do Scalar não aceitam nonce; injeção de CSS sem execução de script tem impacto baixo. Scripts continuam só por nonce |
| `img-src https:` | capas vêm de várias origens (Open Library, Steam, GOG, Commons, Vercel Blob); imagens não executam código |
| Todas as páginas são dinâmicas | exigência do nonce por request (guia de CSP do Next 16); o catálogo segue no Data Cache (`unstable_cache`), então o custo é só o SSR |
| Toolbar/feedback da Vercel em previews bloqueados pela CSP | só afeta membros do time em previews; não se relaxa a CSP de produção por isso |

## Como manter

Toda nova rota, action ou webhook passa pela skill `hexascore-secure-endpoint` e ganha uma linha na tabela de superfícies.
