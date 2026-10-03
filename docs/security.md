# Segurança — OWASP Top 10:2025

Auditoria viva. Revisar a cada milestone; cada mitigação aponta para o código/commit que a implementa quando existir.

## Superfícies

| Superfície | Entrada não confiável | Proteções |
| --- | --- | --- |
| Server Actions (`actions/`) | form data | sessão via `getSessionUser()`, zod, BotID nos fluxos públicos |
| API pública `/api/v1` | query string, header `Authorization` | hash da chave, rate limit + cota (Upstash), zod, só leitura |
| Webhooks Stripe / AbacatePay | body + assinatura | verificação de assinatura/HMAC em tempo constante, idempotência |
| Cron `/api/cron/*` | header | `Authorization: Bearer $CRON_SECRET` |
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

## Como manter

Toda nova rota, action ou webhook passa pela skill `hexascore-secure-endpoint` e ganha uma linha na tabela de superfícies.
