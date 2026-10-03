# Pagamentos, cota e billing

Código em `apps/web/lib/billing/`, ações em `apps/web/actions/billing.ts`, testes `lib/billing/*.test.ts` (PGlite em memória). Como testar webhooks localmente: skill `.claude/skills/hexascore-billing`.

## Catálogo (`lib/billing/catalog.ts`)

Fonte única de preços (centavos). Nada de valor vem do browser: o client manda só `product`/`purpose`/`currency`.

| Item | BRL | USD | EUR |
| --- | --- | --- | --- |
| Pro / mês | 2900 | 600 | 600 |
| Business / mês | 14900 | 2900 | 2700 |
| Pacote de créditos (+25.000 req) | 1900 | 400 | 400 |

PIX: `pro_1m`, `pro_12m`, `business_1m`, `business_12m`, `credits` — 12 meses = 10 × mensal (`PIX_ANNUAL_CHARGED_MONTHS`). Moeda padrão pelo locale (pt-BR→BRL, es→EUR, en→USD), trocável em `/pricing` e no checkout.

## Stripe (cartão)

- `bun run stripe:setup` (em `apps/web`, usa o `STRIPE_SECRET_KEY` do `.env.local`) cria/atualiza de forma idempotente os Products `hexascore_{pro,business,credits}`, Prices com `currency_options` BRL/USD/EUR e `lookup_key` (`hexascore_pro_monthly`, `hexascore_business_monthly`, `hexascore_credits_25k`) e a configuração do Customer Portal (troca de plano: upgrade com proration imediata, downgrade no fim do período; cancelamento no fim do período). Mudou um valor no catálogo → rode de novo: um Price novo herda o `lookup_key` e o antigo é arquivado.
- Runtime resolve o Price pelo `lookup_key` (cache 10 min): nenhum price id em env.
- `startStripeCheckout({ product, currency })`: Checkout Session `mode=subscription` (planos) ou `mode=payment` (créditos), Customer único por usuário (`stripe_customers`), `client_reference_id` + `metadata.user_id`/`kind`, `subscription_data.metadata.user_id`, retorno para `/profile/api/checkout?status=success|canceled`.
- `openBillingPortal()`: sessão do Customer Portal com a configuração do setup.
- Webhook `POST /api/webhooks/stripe` (corpo bruto + `Stripe-Signature`): `checkout.session.completed`, `checkout.session.async_payment_succeeded`, `customer.subscription.created|updated|deleted`, `invoice.paid`, `invoice.payment_failed`. Para assinaturas, o estado é sempre relido via `subscriptions.retrieve` (eventos chegam fora de ordem). Plano vem do `lookup_key` do Price; status: `active|trialing→active`, `canceled|incomplete_expired→canceled`, demais → `past_due` (não concede plano). `current_period_end` = maior `items[].current_period_end` (API `2026-09-30.endive`, SDK `stripe@23.0.0`).
- Respostas: 400 assinatura inválida · 200 aplicado/duplicado/ignorado · 500 falha transitória (Stripe reenvia) · 503 sem `STRIPE_WEBHOOK_SECRET` (checkout continua funcionando).

## AbacatePay (PIX, API v2)

- `createPixCharge({ purpose })` → `POST /v2/transparents/create` (`method: "PIX"`, valor do catálogo, expira em 1 h) e grava `pix_charges`; devolve `brCode`, imagem do QR (base64) e expiração. Um QR pendente do mesmo `purpose` com > 5 min de validade é reaproveitado.
- `GET /api/billing/pix/[id]` (só o dono): polling da página a cada 4 s; se pendente, reconsulta `GET /v2/transparents/check` e liquida/expira.
- O QR abre num modal (Base UI Dialog). Depois de pago, o modal fica travado por 10 s (botão Fechar, ESC e clique fora desativados, contagem anunciada em `role="status"`).
- **Sandbox (só testes)**: com `ABACATEPAY_SANDBOX=true` (default `false`, validado em `lib/env.ts`; só tem efeito com `ABACATEPAY_API_KEY` configurada) o modal mostra o botão “PAGAR PIX SANDBOX”. A Server Action `simulatePixPaymentAction(chargeId)` exige sessão + e-mail verificado, recusa cobrança de outro usuário (`not_found`) ou que não esteja `pending`/válida (`not_pending`) e chama `POST /v2/transparents/simulate-payment?id=<providerRef>` ([docs](https://docs.abacatepay.com/pages/transparents/simulate-payment); só funciona com chave de dev, em produção a AbacatePay devolve erro). A action **não** ativa o plano: a liquidação continua vindo do webhook ou do polling acima (que vê `PAID` no `/check`, mesmo sem webhook alcançando o localhost). O botão tem cooldown de 10 s. Nunca ligue a flag em produção real.
- Webhook `POST /api/webhooks/abacatepay?webhookSecret=…`: confere o secret da URL (`ABACATEPAY_WEBHOOK_SECRET`) e `X-Webhook-Signature` = base64(HMAC-SHA256(corpo bruto, chave pública da AbacatePay)), ambos em tempo constante → 401. Só `transparent.completed` tem efeito; o valor pago tem de bater com `amount_cents`.
- Plano pago: `current_period_end = max(agora, fim atual do mesmo plano PIX) + 1 ou 12 meses`, `provider = abacatepay`.
- Conflitos (checados na criação; a liquidação também nunca sobrescreve): assinatura Stripe ativa bloqueia PIX de plano (créditos liberados); PIX ativo só renova o mesmo plano; assinatura Stripe nova é recusada enquanto houver período PIX ativo (`pix_active`).

## Créditos

Só para Pro/Business ativos. Compra → `credit_ledger` (+25.000, `reason = purchase`, `ref` único = id da Checkout Session ou `pix:<chargeId>`) e, **depois do commit**, `addCachedCredits` (`lib/api/quota.ts`) faz `INCRBY` atômico em `hx:credits:<userId>` **só se a chave existir** — se não existir, o próximo request reconstrói o saldo a partir do ledger, que já inclui a compra.

## Idempotência

`processed_events.id` = id do evento (Stripe `evt_…`, AbacatePay `log_…`, lembretes `reminder:<user>:<fim>:<dias>`), inserido na mesma transação dos efeitos: se a transação falha, nada fica marcado e o reenvio reprocessa. Segunda barreira: `credit_ledger.ref` único e a transição `pix_charges.status pending → paid`.

## Cron `GET /api/cron/billing-reminders` (diário, 12:00 UTC)

E-mail (pt-BR, template `emails/action-email.tsx`) para assinantes PIX 7 dias e 1 dia antes de `current_period_end`; marca `expired` os períodos PIX vencidos (a chave volta para Free porque o handler da API só considera `active` com fim no futuro).

## Cota por request (Upstash)

1. Resolve a chave pelo hash (plano + `user_id`).
2. Rate limit por segundo (`@upstash/ratelimit`, janela deslizante).
3. Cota: `INCR hx:q:{user}:life` (Free) ou `hx:q:{user}:{yyyymm}` (pago).
4. Mensal esgotada → `DECR hx:credits:{user}` (inicializado do ledger).
5. Sem cota e sem crédito → 402.
6. Cron diário grava em `api_usage_daily` e lança débitos no `credit_ledger`.

Fora por ora: faturas próprias, cupons, trial, PIX Automático recorrente, troca de plano via PIX durante um período ativo, estorno automático de PIX pago em conflito (é logado para estorno manual).
