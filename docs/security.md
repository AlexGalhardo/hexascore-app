# Segurança — OWASP Top 10:2025

Auditoria viva. Revisar a cada milestone; cada mitigação aponta para o código/commit que a implementa quando existir.

## Superfícies

| Superfície | Entrada não confiável | Proteções |
| --- | --- | --- |
| Server Actions (`actions/`) | form data | sessão via `getSessionUser()`, zod, BotID nos fluxos públicos |
| `actions/account.ts` (nome, definir senha, perfil privado, agendar/cancelar exclusão) | argumentos da action | exige sessão; alvo sempre é o usuário da sessão; zod (`nameSchema`, `passwordSchema`, `z.boolean()`); `setPassword` só no servidor; exclusão faz sign-out e apaga todas as sessões |
| `actions/contact.ts` | form data | BotID; usuário logado tem nome/e-mail tirados da sessão (body ignorado); `contactSchema`; rate limit 5/h por usuário ou IP na tabela `rate_limit` |
| Better Auth `/api/auth/*` | JSON | origin check, rate limit em banco (429), hook de senha em todo endpoint que aceita senha, BotID em `/sign-up/email`, `/sign-in/email`, `/sign-in/username`; verificação de e-mail obrigatória; recuperação e reenvio sempre respondem igual (sem enumeração) |
| `/api/cron/purge-accounts` | header `Authorization` | `Bearer $CRON_SECRET` comparado em tempo constante (sha256 + `timingSafeEqual`); anonimiza só contas com `deletion_scheduled_at` < agora − 30 dias |
| API pública `/api/v1` | query string, header `Authorization` | hash da chave, rate limit + cota (Upstash), zod, só leitura |
| Webhook Stripe `POST /api/webhooks/stripe` | body bruto + `Stripe-Signature` | `constructEventAsync` (tolerância de 5 min), 400 em assinatura inválida, idempotência por `event.id` em `processed_events` na mesma transação dos efeitos, estado da assinatura relido da API (eventos fora de ordem), crédito com `ref` único = id da Checkout Session; 503 com log se `STRIPE_WEBHOOK_SECRET` faltar |
| Webhook AbacatePay `POST /api/webhooks/abacatepay` | body bruto, `?webhookSecret=`, `X-Webhook-Signature` | secret da URL e HMAC-SHA256 base64 (chave pública da AbacatePay) comparados em tempo constante → 401; zod só nos campos lidos; idempotência por `id` do evento + transição `pending → paid`; valor pago tem de bater com `amount_cents` gravado |
| Server Actions `actions/billing.ts` | `{ product, currency }` / `{ purpose }` | sessão + e-mail verificado; zod; preço por `lookup_key` (Stripe) ou catálogo (PIX), nunca do client; `client_reference_id`/metadata = id da sessão; regras de conflito Stripe×PIX no servidor; QR pendente reaproveitado (sem spam de cobranças) |
| `GET /api/billing/pix/[id]` | `id` | sessão, zod `uuid`, filtro por dono (404 para os demais), `no-store`; reconsulta a AbacatePay com timeout |
| Páginas `/pricing`, `/profile/api/checkout` | query `currency`/`plan`/`status` | allowlist via zod; DTO `BillingSummary` sem ids de provedor |
| Cron `/api/cron/*` (`flush-usage`, `billing-reminders`) | header | `Authorization: Bearer $CRON_SECRET` (tempo constante); lembrete gravado em `processed_events` antes do envio (sem e-mail duplicado) |
| Ingestão (Python) | payloads de terceiros | escrita só em `source_records`; normalizador valida antes do upsert; texto renderizado sempre escapado |

## A01–A10

| Risco | Mitigação planejada |
| --- | --- |
| A01 Broken Access Control | identidade só da sessão; checagem de dono em toda escrita; `/admin` exige `role=admin` no servidor; perfis privados filtrados no repositório |
| A02 Security Misconfiguration | CSP estrita com nonce, HSTS, `X-Content-Type-Options`, `Referrer-Policy`, cookies `HttpOnly`/`Secure`/`SameSite=Lax`; ZAP no gate `dev` → `main` |
| A03 Software Supply Chain Failures | versões fixadas, lockfile commitado, `bun audit`, Dependabot, CodeQL |
| A04 Cryptographic Failures | senhas com o hash do Better Auth; API keys só como sha256; segredos fora de `NEXT_PUBLIC_*` |
| A05 Injection | Drizzle com parâmetros; zero SQL concatenado; React escapa saída; sinopses de terceiros como texto puro |
| A06 Insecure Design | valores de dinheiro e cota calculados no servidor; limiar de 5 avaliações; cota Free amarrada a e-mail verificado |
| A07 Authentication Failures | regras de senha, verificação de e-mail, 2FA, passkeys, rate limit em login/reset, BotID |
| A08 Software or Data Integrity Failures | webhooks assinados e idempotentes; migrations revisadas; release-please com tags |
| A09 Logging & Alerting Failures | Vercel Observability, `request_id` em toda resposta da API, alertas Telegram para falhas de webhook/cron |
| A10 Mishandling of Exceptional Conditions | erros RFC 9457 sem stack trace; falhas de captcha/BotID e de verificação fecham (deny); timeouts nas chamadas externas |

## Riscos aceitos

| Risco | Motivo |
| --- | --- |
| Contas múltiplas para multiplicar a cota Free | limitado por e-mail verificado + BotID; 100 req não compensa o abuso |
| Signup revela "nickname já em uso" | nickname é público (`/u/{nickname}`); e-mail duplicado recebe a mesma resposta genérica de sucesso (Better Auth com `requireEmailVerification`) |
| BotID não roda fora da Vercel | `lib/bot.ts` só faz bypass quando `VERCEL !== "1"`; na Vercel qualquer falha da checagem nega (fail closed) |

## Como manter

Toda nova rota, action ou webhook passa pela skill `hexascore-secure-endpoint` e ganha uma linha na tabela de superfícies.
