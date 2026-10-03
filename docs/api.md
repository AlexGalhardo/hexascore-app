# API pública /v1


- Base `https://api.hexascore.com/v1`. Auth `Authorization: Bearer hx_live_…` (chave exibida 1 vez, só hash no banco). Só leitura (`GET`).

| Rota | Uso |
| --- | --- |
| `/v1/{games\|movies\|series\|books}` | lista + filtros: `genre`, `platform`, `year`, `min_score`, `max_score`, `min_{categoria}` (ex. `min_gameplay=80`), `sort` (`score`, `-score`, `release_date`, `rating_count`), `limit`, `cursor`, `locale` |
| `/v1/{type}/{id}` | detalhe: metadados, metascore, médias por categoria, IDs externos, sinopse com `license` + `source_url` |
| `/v1/{type}/genres` | taxonomia para filtros |
| `/v1/search?q=&type=` | busca fuzzy entre tipos |
| `/v1/lookup?source=imdb&id=tt…` | busca por ID externo |
| `/v1/me/usage` | cota restante, plano, créditos |

- Lista: `{ "data": [...], "next_cursor": "…" | null }` (paginação por cursor).
- Erros RFC 9457 (`application/problem+json`) com `type` linkando a página do erro nas docs e `request_id`.
- Headers: `RateLimit-*` (IETF), `X-Quota-Remaining`, `X-Request-Id`. 402 sem cota/créditos (link para `/pricing`); 429 com `Retry-After`.
- Cota conta toda request autenticada, exceto 429 e 5xx.
- Versionamento por path; quebra → `/v2`, antiga com headers `Deprecation` + `Sunset` e changelog da API.
- Dados de catálogo com cache do Next (`"use cache"` + tags invalidadas pela ingestão); checagem de cota sempre roda.
- Contrato: zod 4 → `openapi.json` 3.1 gerado no build; CI falha se o arquivo mudar sem commit.
- Docs: `/docs` (quickstart, auth, erros, paginação, limites) + referência Scalar (snippets cURL/JS/Python automáticos). `/profile/api` = mesmo Scalar com a chave do usuário + dashboard de uso.
- SDKs gerados: `@hey-api/openapi-ts` (npm) e `openapi-python-client` (PyPI), publicados via release-please.

Fora por ora: endpoints de escrita, notas individuais de usuários, webhooks para clientes, GraphQL.
