# Processo: git, CI/CD, qualidade


**Git**: `feat/*`/`fix/*` → PR para `dev` (sandbox: preview Vercel + branch Neon + seed sintético) → CI verde → PR `dev` → `main` (produção, Rolling Releases). Push direto deve ser bloqueado por ruleset (ver abaixo; hoje pendente). Conventional Commits via commitlint + husky. release-please na `main`.

**Proteção de `main` e `dev` — pendente (plano do GitHub)**: o repo é privado no plano Free, e tanto branch protection quanto rulesets respondem `HTTP 403 Upgrade to GitHub Pro or make this repository public`. A regra está pronta em `.github/rulesets/protect-main-dev.json` (PR obrigatório com 0 aprovações — risco aceito: projeto com um único dono —, checks `lint`/`typecheck`/`test`/`build`/`audit` em modo estrito (branch atualizada com a base), sem bypass, sem force push nem deleção); aplicar após upgrade ou ao tornar o repo público:

```sh
gh api -X POST repos/AlexGalhardo/claude-code-internet-api/rulesets --input .github/rulesets/protect-main-dev.json
```

Decisão do dono (2026-10-03): seguir sem proteção por enquanto; o fluxo `feat/*` → `dev` → `main` é convenção, não bloqueio. "Allow GitHub Actions to create and approve pull requests" já está habilitado, então o release-please abre o PR de release.

**CI (GitHub Actions) — implementado**

| Workflow | Gatilho | Jobs / o que roda |
| --- | --- | --- |
| `.github/workflows/ci.yml` | PR e push em `dev`/`main` | `lint` (`bun run lint:ci`) · `typecheck` (`bun run typecheck`) · `test` (`bun test` em `apps/web`, PGlite) · `build` (`bun run build`, depende dos 3 anteriores; env dummy válido para `lib/env.ts`) · `audit` (`bun audit --audit-level=high`) |
| `.github/workflows/codeql.yml` | PR/push em `dev`/`main`, semanal, manual | CodeQL `javascript-typescript` + `python` (python só quando houver `.py`). Repo privado: só roda com a variável de repo `CODE_SCANNING_ENABLED=true` (exige GitHub Code Security) |
| `.github/workflows/release-please.yml` | push em `main` | `googleapis/release-please-action@v5` com `release-please-config.json` + `.release-please-manifest.json` |
| `.github/workflows/qa.yml` | PR `dev` → `main`, manual (input `url` opcional) | `preview` espera o deploy Vercel do SHA do PR via GitHub Deployments API (até 20 min) · `e2e` roda `qa/e2e` (Playwright + axe, 4 projetos) contra a URL com `x-vercel-protection-bypass` (`secrets.VERCEL_AUTOMATION_BYPASS_SECRET`); specs `@auth` precisam de `secrets.E2E_DATABASE_URL` + `secrets.E2E_BETTER_AUTH_SECRET` (banco e secret do preview), senão são puladas. Relatório HTML como artifact |

**Gate `dev` → `main`**: o workflow `QA` é o gate de merge por convenção. Branch protection está desligada por decisão do dono (ver acima), então ele é consultivo: não mesclar `dev` → `main` com o `QA` vermelho. Schemathesis, ZAP e k6 ainda não fazem parte dele (lacuna do 0.6).

**Feature flags** (`apps/web/lib/flags.ts`, `flags` 4.3.1, avaliadas só no servidor, padrão ligado):

| Flag | Desliga | Como desligar |
| --- | --- | --- |
| `pix-checkout` | opção PIX em `/profile/api/checkout` e `createPixCharge` | env `FLAG_PIX_CHECKOUT=off` no projeto Vercel + redeploy |
| `public-api-signup` | criação de novas chaves em `/profile/api` (chaves existentes seguem válidas) | env `FLAG_PUBLIC_API_SIGNUP=off` + redeploy |

Override por navegador (sem redeploy): definir `FLAGS_SECRET` (32 bytes base64url, `openssl rand -base64 32 | tr '+/' '-_' | tr -d =`) no projeto; a Vercel Toolbar lista as flags via `/.well-known/vercel/flags`.
| `.github/dependabot.yml` | semanal | `bun` (raiz), `uv` (`webscrappings/`), `github-actions`, `docker` + `docker-compose` (`infra/`); minor/patch agrupados, PRs para `dev`, prefixos `chore(deps)`/`ci` |

Todos os jobs usam `.github/actions/setup-bun` (Bun `1.4.2` via `oven-sh/setup-bun@v2` + cache de `~/.bun/install/cache` + `bun install --frozen-lockfile`), `concurrency` com `cancel-in-progress` e `permissions` mínimas. Actions fixadas por major: `actions/checkout@v7`, `actions/cache@v6`, `github/codeql-action@v4`. Validar localmente: `docker run --rm -v "$PWD:/repo" -w /repo rhysd/actionlint:1.7.12`.

**Releases**: release-please (tipo `simple`, manifest em `0.0.0`) abre/atualiza um PR de release na `main` a cada push; `feat`/`fix` antes da 1.0 sobem o patch (primeira release `0.0.1`), merge do PR cria tag `vX.Y.Z` + GitHub Release e escreve a seção no `CHANGELOG.md` (`feat`→Added, `fix`→Fixed, `perf`/`refactor`/`revert`→Changed; `docs`/`chore`/`ci`/`test`/`build`/`style` ocultos). A seção `[Unreleased]` do `CHANGELOG.md` continua manual. PRs abertos pelo `GITHUB_TOKEN` não disparam workflows, então os checks obrigatórios não rodam no PR de release: para os checks rodarem no PR de release, configure o secret `RELEASE_PLEASE_TOKEN` (fine-grained PAT com `contents` e `pull-requests` de escrita) — o workflow usa ele quando existir; sem ele, feche e reabra o PR para disparar o CI. Não há bypass no ruleset.

**CI — alvo (milestones seguintes)**
- Todo PR: Biome (lint + format) · `tsc --noEmit` · `bun test` (PGlite) · integração (Postgres 18 + Redis services) · diff do `openapi.json` · `next build` + smoke · `ruff` + `pytest` · `bun audit --audit-level=high` · CodeQL · Dependabot.
- PR `dev` → `main` (contra o preview): ~~Playwright E2E (mobile + desktop, light/dark, 3 idiomas, axe)~~ (feito, `qa.yml`) · Schemathesis · OWASP ZAP baseline/API · k6 curto.
- Agendado/manual: agente QA exploratório (dirige o browser por todas as rotas com edge cases: unicode, limites, payloads XSS) → relatório em `qa/reports/` + issue por bug.

**Segurança**: `docs/security.md` com tabela OWASP Top 10:2025 (A01–A10: superfícies, mitigação, commit, riscos aceitos), revisada a cada milestone.

**Agentes**: `AGENTS.md` é a fonte; `CLAUDE.md` importa `@AGENTS.md`. `.claude/rules/`: `karpathy.md`, `ponytail.md`, `graphify.md`. Skills do projeto: `hexascore-*` em `.claude/skills/` (lista em `AGENTS.md`). Worktrees + agentes paralelos para `webscrappings/`, `sdks/` e `docs/`.

**SDKs da API (0.4)**: `sdks/ts` (`@hexascore/sdk`, workspace Bun, `@hey-api/openapi-ts` 0.99.0) e `sdks/python` (`hexascore`, `openapi-python-client` 0.29.1 via `uvx`), gerados de `apps/web/public/openapi.json` (`bun run sdk:generate` / `bun run sdk:generate:python`) e commitados. Regerar sempre que `bun run openapi` mudar o spec. **Publicação ainda não feita** (sem credenciais): metadados prontos (`name`, `version` 0.0.0, `files`, `publishConfig.access=public` no npm; `pyproject.toml` gerado com backend `uv_build`). Para ligar: secrets `NPM_TOKEN` e `PYPI_API_TOKEN` (ou Trusted Publishing do PyPI), adicionar os pacotes ao `release-please-config.json` (`node` para `sdks/ts`, `python` para `sdks/python/hexascore`) e um job de publish disparado pela release (`bun run build && npm publish --provenance` / `uv build && uv publish`). O gerador TS exige o compilador JS do TypeScript, por isso `sdks/ts` fixa `typescript` 6.0.3 (o app usa o 7.0.2 nativo, sem API JS).

**Dependências**: só versões estáveis, fixadas sem `^`. Imagens Docker com tag fixa (`oven/bun:<versão>`, `postgres:18.x`). Nenhuma dependência nova sem aprovação.
