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

## Comandos implementados

| Verificação | Local (raiz do repo) | CI (`.github/workflows/ci.yml`) |
| --- | --- | --- |
| Lint + format | `bun run lint:ci` | job `lint` |
| Tipos | `bun run typecheck` | job `typecheck` |
| Unitário (PGlite, sem Docker) | `bun run test` (ou `bun test` em `apps/web`) | job `test` |
| Build | `bun run build` | job `build` (depois de lint/typecheck/test) |
| Vulnerabilidades | `bun audit --audit-level=high` | job `audit` |
| SAST | — | `.github/workflows/codeql.yml` |
| Workflows | `docker run --rm -v "$PWD:/repo" -w /repo rhysd/actionlint:1.7.12` | — |
| Scripts de setup | `docker run --rm -v "$PWD:/mnt" -w /mnt koalaman/shellcheck:v0.11.0 -x setups/*.sh` | — |

`lint`, `typecheck`, `test` e `build` são os checks obrigatórios do ruleset de `dev` e `main` (`.github/rulesets/protect-main-dev.json`; aplicação pendente, ver `docs/process.md`). Antes de abrir PR rode os quatro localmente.

Banco para testar migrations à mão: `setups/setup-*-docker.*` (Postgres 18 em `localhost:5432`) ou `setups/setup-*-pglite.*`; ambos rodam `bun run db:migrate`.

Ainda **não** existem workflows de E2E (Playwright), Schemathesis, OWASP ZAP nem k6: entram no milestone em que as respectivas camadas forem criadas (gatilho: PR `dev` → `main`), junto com o job de integração com Postgres 18 + Redis como service containers.

## Dados de teste

- Unitário/integração: fixtures mínimas por teste.
- Dev/preview/sandbox: snapshot de ~100k itens + 10k usuários sintéticos (faker com seed fixa). Nunca em produção.
