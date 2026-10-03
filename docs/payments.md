# Pagamentos, cota e billing


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
