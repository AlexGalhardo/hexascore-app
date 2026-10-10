# API pública /v1

- Base `https://api.hexascore.app/v1` (proxy reescreve `api.<domínio>/v1/*` → `/api/v1/*` em `apps/web/proxy.ts`). A mesma API também responde em `https://<deploy>/api/v1`. Auth `Authorization: Bearer hx_live_…` (chave exibida 1 vez, só hash sha256 no banco). Só leitura (`GET`).

| Rota | Uso |
| --- | --- |
| `/v1/{games\|movies\|series\|books}` | lista + filtros: `genre`, `platform` (só games), `year`, `min_score`, `max_score`, `min_{categoria}` (chaves de `lib/categories.ts` por tipo, ex. `min_gameplay=80`), `sort` (`score\|-score\|release_date\|-release_date\|rating_count\|-rating_count\|popularity\|-popularity`, padrão `-popularity`), `limit` (1..`maxPageSize` do plano: Free 10, pagos 100, padrão 10), `cursor`, `locale` (`en\|pt-BR\|es`, padrão `en`). Resposta `{ data: ItemSummary[], next_cursor }` |
| `/v1/{type}/{id}` | `ItemDetail`: metadados, `score` + `categories` (null abaixo de 5 notas) + `rating_count`, `external_ids` (inclui `wikidata`), `trailers` (YouTube), `synopsis { text, locale, license, source_url }` ou null, `url` da página, `time_to_beat { main, extras, completionist: { median_hours, reports } }` (só jogos, `null` nos outros tipos; `median_hours` null abaixo de 3 relatos). Id de outro tipo → 404 |
| `/v1/{type}/genres` | `{ data: { slug, count }[] }` (mais comuns primeiro; `slug` = valor exato para `genre`) |
| `/v1/search?q=&type=&limit=&locale=` | busca fuzzy entre tipos (`type` = `game\|movie\|series\|book`), até 50, sem paginação |
| `/v1/lookup?source=imdb\|steam\|igdb\|tmdb\|isbn13\|openlibrary\|wikidata&id=` | `ItemDetail` ou 404 |
| `/v1/me/usage` | `{ plan, quota { limit, used, remaining, period, resets_at }, credits { balance }, rate_limit { requests_per_second } }` (conta como 1 request) |

- JSON snake_case, datas ISO 8601, ordenação estável (desempate por id). Parâmetros desconhecidos → 400 (`z.strictObject`).
- **Nunca** expor `external_scores` (notas de terceiros são só para o front, por licença) — coberto por teste.
- Erros RFC 9457 (`application/problem+json`) com `code` estável, `type` = `/docs/errors#<code>` e `request_id`.
- Headers: `RateLimit-Limit/Remaining/Reset` (IETF), `X-Quota-Remaining`, `X-Request-Id`, `Cache-Control: private, no-store`. 402 sem cota/créditos; 429 com `Retry-After`; 503 (com `Retry-After`) se o store de cota falhar — fail closed.
- Contagem: toda request autenticada que chega ao handler conta (inclusive 404 de id). Não contam: 401, 400 de validação de parâmetros, 404 de tipo/rota desconhecidos, 429 e 5xx (estornados).
- Versionamento por path; quebra → `/v2`, antiga com `Deprecation` + `Sunset` por ≥ 6 meses e changelog em `/docs/changelog`.
- Leituras do catálogo vão direto ao repositório (`lib/catalog/repository.ts`, queries indexadas, respostas `no-store`); `unstable_cache` só se o p95 do k6 (0.6) exigir. A checagem de cota sempre roda.

## Implementação

- `lib/api/handler.ts` — `apiRoute({ params?, query, response }, handler)`: request id, auth por chave (atualiza `last_used_at` no máx. 1×/min), plano, schema de query (pode depender do plano e dos params), cota/throttle, `response.parse` no corpo (contrato = fio), `NotFoundError`/`BadRequestError`, estorno em 5xx.
- `lib/api/schemas.ts` — DTOs zod (fonte única) + schemas de params/query + mapeadores repositório → snake_case.
- `lib/api/openapi.ts` — OpenAPI 3.1 a partir dos mesmos schemas (`z.toJSONSchema`, registry próprio). `bun run openapi` grava `apps/web/public/openapi.json` (en, base dos SDKs) + `openapi.pt-BR.json` + `openapi.es.json`; `bun run openapi:check` falha se algum dos três estiver desatualizado (job `lint` do CI).
- Textos do spec (info, tags, summaries/descriptions de operações, parâmetros, campos, headers, erros, security scheme) ficam em `apps/web/messages/<locale>/openapi.json` (fora dos `NAMESPACES` do next-intl). `lib/api/openapi-text.ts`: `openApiText(locale, key, params)` tipado pelas chaves do en; nos schemas zod, `.meta({ description: apiDoc(key) })` grava o inglês e registra a chave, e `localizeDescriptions` troca pelo idioma depois do `z.toJSONSchema` (descrição sem chave = erro). Estrutura idêntica entre idiomas (só description/summary mudam), garantida por `lib/api/openapi.test.ts`; paridade de chaves em `messages/messages.test.ts`. `example.detail` dos erros fica em inglês (é o corpo real).
- Rotas em `apps/web/app/api/v1/` (`[type]`, `[type]/[id]`, `[type]/genres`, `search`, `lookup`, `me/usage`). Testes de rota em `lib/api/v1.test.ts` (PGlite + store em memória).
- Passo a passo para novo endpoint: skill `.claude/skills/hexascore-api/SKILL.md`.

## Docs e dashboard

- `/docs` (chrome traduzido en/pt-BR/es, guias em inglês): quickstart, authentication, errors (âncora por `code`), pagination, rate-limits (tabela de `PLAN_LIMITS`), changelog; `/docs/reference` = Scalar (`@scalar/api-reference-react` 0.9.77) carregando o spec do idioma do site (`openApiPath(locale)` em `lib/api/spec-path.ts`: `/openapi.json`, `/openapi.pt-BR.json`, `/openapi.es.json`) com a UI do Scalar no mesmo idioma, tema seguindo `data-theme`, IA/MCP do Scalar desligados. A raiz de `api.<domínio>` lista os três em `openapi: { en, "pt-BR", es }`.
- `/profile/api` (login obrigatório → `/login?next=/profile/api`): chaves (criar com exibição única + copiar, revogar com confirmação), plano/cota, uso de hoje (Redis) + 30 dias (`api_usage_daily`) em gráfico SVG, links para `/pricing` e `/profile/api/checkout`, "Try it" com Scalar e a chave só em memória (persistência de auth desligada).

## SDKs

Três destinos, dois pacotes, todos na versão `0.1.0` e ainda não publicados (falta criar as contas e os escopos):

| Destino | Pacote | Pasta | O que vai | Guia |
| --- | --- | --- | --- | --- |
| npm | `@hexascore/sdk` | `sdks/ts` | `dist/` (ESM + `.d.ts`) | [`sdk-publish-npm.md`](sdk-publish-npm.md) |
| JSR | `@hexascore/sdk` | `sdks/ts` | o TypeScript de `src/` (`jsr.json`) | [`sdk-publish-jsr.md`](sdk-publish-jsr.md) |
| PyPI | `hexascore` | `sdks/python` | sdist + wheel (`uv build`) | [`sdk-publish-pypi.md`](sdk-publish-pypi.md) |

- `sdks/ts` (workspace): `@hey-api/openapi-ts` 0.99.0 (client fetch), `bun run sdk:generate`. Único código manual: `src/index.ts` (`createHexascoreClient`, `HEXASCORE_BASE_URL`). O gerador emite imports relativos com `.ts` (`importFileExtension`), exigência do JSR; o build do npm reescreve para `.js` (`rewriteRelativeImportExtensions`), o que faz o pacote funcionar no ESM do Node sem bundler. Testes: `bun test` em `sdks/ts` (fetch simulado). Sem dependências de runtime.
- `sdks/python`: `openapi-python-client` 0.29.1 via `uvx`, `bun run sdk:generate:python`, com `--meta none`: só o pacote `sdks/python/hexascore/` é gerado; `pyproject.toml`, `README.md`, `LICENSE` e `tests/` são escritos à mão e sobrevivem à regeneração (o script recria o `py.typed`). Testes: `uv run pytest -q` em `sdks/python`. Dependências: `httpx`, `attrs`.
- Versão: a mesma em `sdks/ts/package.json`, `sdks/ts/jsr.json` e `sdks/python/pyproject.toml`, independente da versão do site. Sobe quando o contrato muda.
- Publicação é manual (GitHub Actions desativado); cada guia traz o caminho de automação por OIDC para quando voltar.
- Os testes dos SDKs não entram no `ci:local`; rode-os à mão antes de publicar (estão no checklist de cada guia).

Fora por ora: endpoints de escrita, notas individuais de usuários, webhooks para clientes, GraphQL, lista de últimas requests no dashboard.
- Parâmetro de query vazio (`?min_score=`) é sempre 400 `bad_request` (`errors[].message = "Must not be empty."`), validado em `apiRoute` antes do zod: `z.coerce` leria `""` como `0`.
