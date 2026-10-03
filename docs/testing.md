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

## Dados de teste

- Unitário/integração: fixtures mínimas por teste.
- Dev/preview/sandbox: snapshot de ~100k itens + 10k usuários sintéticos (faker com seed fixa). Nunca em produção.
