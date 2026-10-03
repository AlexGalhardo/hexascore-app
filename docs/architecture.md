# Arquitetura


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
