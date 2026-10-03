# Ingestão de dados

Sem scraping de HTML: só APIs públicas e Wikidata.

| Fonte | Como | Fornece |
| --- | --- | --- |
| Wikidata | SPARQL (seleção de QIDs) + `wbgetentities` (mesmo JSON do dump) | catálogo base, títulos por idioma, gêneros, datas, IDs externos, notas MC/RT (P444 + P447), trailer YouTube (P1651), imagem Commons (P3383/P18), popularidade (sitelinks) |
| Wikipedia | Action API `prop=extracts` em lotes de 20 (en, pt, es) | sinopse por idioma (`CC BY-SA 4.0` + URL da página) |
| Open Library | API (`/isbn`, `/books`, `/works`) | páginas (fallback do P1104), capa |
| Steam | `appdetails` público | capa (`header_image`), plataformas (fallback do P400) |
| GOG | API pública de catálogo (`catalog.gog.com/v1/catalog`) | link da loja, disponibilidade, capa vertical, plataformas (fallback) |
| Epic, IMDb, MC, RT, HLTB, Goodreads | sem fetch | deep-link pelo ID do Wikidata em `items.external_ids` |
| Skoob | sem fetch | Wikidata não tem propriedade de ID do Skoob: chave `skoob` fica vazia por ora |
| TMDB, IGDB | só ID | fetch fora da v1 (exige licença comercial) |

```
webscrappings/
├─ pyproject.toml   uv, Python ≥ 3.14 (uuid.uuid7), deps fixas: httpx, psycopg[binary], pydantic; dev: pytest, ruff
├─ common/          http.py (rate limit por host, retry 429/5xx com Retry-After, timeout, User-Agent),
│                   db.py (Record, canonical_hash, upsert_source_record, store_records), cli.py (--since)
├─ connectors/      wikidata.py, wikipedia.py, openlibrary.py, steam.py, gog.py → fetch(since, ...) -> Iterable[Record]
├─ normalize.py     source_records → items / item_translations / external_scores (build_item puro + upsert)
├─ bootstrap.py     carga completa (top-N por sitelinks) + --export snapshot.sql.gz via COPY
├─ covers.py        meta.coverUrl + meta.coverThumbUrl (WebP no Vercel Blob se houver token)
├─ revalidate.py    POST ${APP_URL}/api/cron/revalidate {"tags":["catalog"]}
└─ tests/           pytest com respostas reais gravadas em fixtures/, sem rede; test_db.py opcional
```

## Fluxo

1. **Conector** valida a resposta com pydantic e grava em `source_records` um payload enxuto (só os campos usados, para que edições irrelevantes não mudem o hash). `upsert_source_record` calcula sha256 do JSON canônico (chaves ordenadas): hash igual = no-op (nem `fetched_at` muda).
2. **`external_id` é sempre o QID do Wikidata**, em todas as fontes (o ID nativo fica dentro do payload). Assim a PK `(source, external_id)` serve de índice para juntar todas as fontes de um item. `source` ∈ `wikidata`, `wikipedia:en|pt|es`, `openlibrary`, `steam`, `gog`.
3. **Normalizador** pega os QIDs com registro pendente (`normalized_at IS NULL OR normalized_at < fetched_at`), relê *todos* os registros desses QIDs e reconstrói o item do zero com `build_item` (função pura, testada). Upsert por `wikidata_qid`, commit a cada 200 QIDs (ids existentes do lote buscados numa query; o lote é gravado em pipeline mode do psycopg, sem esperar a resposta de cada statement; só o slug de um item novo espera), marca `normalized_at` e `item_id`. Sem registro Wikidata ou sem título em inglês = pulado.
4. **Capas** (`covers.py`): prioridade Open Library (livros) → Steam → GOG (jogos) → Commons P3383/P18 (todos, com `meta.coverAttribution` = fonte, licença, autor, URL). Com `BLOB_READ_WRITE_TOKEN` baixa a imagem (máx. 15 MB), aplica a orientação EXIF, redimensiona para caber em 600×600 (mantém proporção, nunca amplia), codifica WebP (Pillow, `quality=78`, `method=6`, sem EXIF/XMP/ICC) e sobe para o Vercel Blob em `covers/<item_id>-<sha256(fonte)[:16]>.webp` (`meta.coverUrl`). Do mesmo download gera a miniatura das listas e do autocomplete: cabe em 96×144 (2x da miniatura de 48 px de altura, mantém proporção, nunca amplia), `quality=70`, `method=6`, sem metadados, em `covers/<item_id>-<hash16>-t.webp` (`meta.coverThumbUrl`). Ambos públicos, `cache-control` 1 ano, pois o conteúdo de um pathname nunca muda. Idempotente por tamanho: se `coverUrl` e `coverThumbUrl` já apontam para os pathnames do hash da imagem baixada, não recodifica nem sobe (com `--refresh`, só baixa a fonte); itens gravados antes das miniaturas (ou sem token) são selecionados de novo quando há token e recebem só o que falta. Sem token grava a URL da fonte em `coverUrl`, deixa `coverThumbUrl` vazio (o front cai na capa completa) e registra um aviso. Os itens são processados em paralelo (`WORKERS` = 8 threads para download, encode e upload, sob o throttle por host do `HttpClient`); a conexão do banco fica na thread principal, que grava cada bloco de 64 itens com um `executemany` e um commit. Falha num item (HTTP, imagem inválida, resposta do Blob fora do schema) só gera log; o item fica sem capa e é tentado de novo na próxima execução. Média medida em 12 capas reais (6 Open Library L, 5 headers Steam, 1 fixture Commons de `tests/fixtures`): fonte 44,4 KB → capa 18,2 KB + miniatura 1,8 KB (fixture Commons: 1,5 KB; faixa 0,9–4,3 KB). Por 1M itens: ~18,6 GB de capas + ~1,8 GB de miniaturas ≈ **20,5 GB** e 2M puts. Blobs antigos de uma capa trocada ficam órfãos (raro). O normalizador preserva `coverUrl`/`coverThumbUrl`/`coverAttribution`.
5. **Revalidação**: `normalize` e `bootstrap` chamam `POST ${APP_URL}/api/cron/revalidate` com `Authorization: Bearer ${CRON_SECRET}`; falha só gera log.

## Mapeamento

- `type`: classe usada na seleção SPARQL (`P31`): filme Q11424/Q24869/Q202866 · série Q5398426/Q581714/Q1259759/Q526877 · jogo Q7889 · livro Q7725634/Q571/Q8261. Não muda depois de criado.
- `slug`: título em inglês em ASCII minúsculo; colisão no mesmo tipo → `-<ano>`, depois `-<qid>`; título sem caracteres latinos → `<qid>`. Não muda depois de criado (URL estável).
- `release_date`: menor P577 (séries: P580 primeiro; fallback P571). Precisão de mês/ano vira dia/mês 1; abaixo de ano = nulo.
- Títulos: label `en` (fallback `mul`, depois sitelink enwiki); `pt-BR` ← `pt-br`/`pt`/`mul`; `es` ← `es`/`mul`; sem label usa o título da Wikipedia.
- `external_ids`: imdb P345, igdb P5794, steam P1733, tmdb P4947 (filme)/P4983 (série), isbn13 P212 (só dígitos), openlibrary P648, hltb P2816, goodreads P8383, metacritic P1712, rottentomatoes P1258, epic P6278, gog P2725. Valor de rank preferido primeiro.
- `meta`: genres P136, platforms P400 (jogos; fallback Steam/GOG), creators (filme P57, série P170, jogo P178, livro P50), runtimeMinutes P2047, pages P1104 (fallback Open Library), seasons P2437, episodes P1113, countries P495, trailers P1651 (máx. 3), coverUrl. Rótulos em inglês; chaves vazias são omitidas.
- `external_scores`: P444 com P447 = Metacritic (Q150248, só Metascore `NN/100`) → `metacritic`; Rotten Tomatoes (Q105584, só Tomatometer `NN%`) → `rottentomatoes`. Vale o mais recente (P585), rank preferido primeiro. `value` é o texto do Wikidata (`73/100`, `83%`); `url` montada do ID externo. Notas que sumiram do Wikidata são apagadas.
- `popularity` = número de sitelinks.

## Comandos

Ver [`webscrappings/README.md`](../webscrappings/README.md). Resumo (em `webscrappings/`):

```sh
uv sync
uv run python -m bootstrap --limit 20                     # carga local pequena
uv run python -m bootstrap --export-only --export snapshot.sql.gz
uv run python -m connectors.wikidata --since 36h          # --types, --limit, --refresh-known
uv run python -m connectors.wikipedia --since 8d          # --langs, --missing-only
uv run python -m normalize
uv run python -m covers
uv run ruff check && uv run ruff format --check && uv run pytest
```

Contra o sandbox, sem o GitHub Actions (o mesmo que o bootstrap com `skip_catalog=true`): os enriquecimentos rodam em paralelo, um processo por fonte/idioma, e depois `normalize` e `covers`. O código lê só `DATABASE_URL` e `BLOB_READ_WRITE_TOKEN` do ambiente:

```sh
E="--env-file ../apps/web/.env.sandbox --env-file .env.local"
uv run $E python -m connectors.wikipedia --missing-only --langs en &   # idem pt, es, openlibrary, steam, gog
wait
uv run $E python -m normalize && uv run $E python -m covers
```

O token do Blob precisa ser de um store **público**: num store privado todo upload de capa responde 400 ("Cannot use public access on a private store") e o item fica sem capa.

Snapshot: `.sql.gz` com `COPY ... FROM stdin` de `items`, `item_translations`, `external_scores` (sem a coluna gerada `search`), gerado por psycopg (não depende de `pg_dump`). Restaurar num banco migrado: `gunzip -c snapshot.sql.gz | psql "$DATABASE_URL"`.

## Agenda (GitHub Actions)

Secrets: `DATABASE_URL`, `APP_URL`, `CRON_SECRET`, `BLOB_READ_WRITE_TOKEN`. Os três workflows de ingestão compartilham `concurrency: ingest` (nunca dois normalizadores ao mesmo tempo).

- **bootstrap** (`ingest-bootstrap.yml`, manual, inputs `limit`/`types`): job Wikidata top-N por sitelinks (~35k filmes, 15k séries, 25k jogos, 25k livros) → 6 jobs paralelos de enriquecimento (Wikipedia en/pt/es, Open Library, Steam, GOG) com `--missing-only` → normalize, covers, snapshot publicado como GitHub Release pré-release `catalog-snapshot-<n>`. Cada job tem teto de 6h; se algum estourar, basta re-disparar (retoma do que falta).
- **incremental** (`ingest-incremental.yml`, diário 04:00 UTC): entidades editadas no Wikidata nas últimas 36h (com ≥ 15 sitelinks) → Wikipedia dos registros alterados → Steam dos itens alterados → normalize, covers.
- **semanal** (`ingest-weekly.yml`, domingo 05:00 UTC): relê todas as entidades do catálogo (notas P444 e o que o noturno perdeu) → sinopses alteradas/faltantes → Open Library, GOG → normalize, covers.
- **CI** (`python-ci.yml`, PR para `dev`/`main` tocando `webscrappings/**`): ruff check, ruff format --check, pytest (testes de banco contra Postgres 18.6 quando as migrations existem no branch).

Wikipedia: Action API (`action=query&prop=extracts|pageprops|info`, `exintro` + `explaintext`) com 20 títulos por requisição (`BATCH`, teto do TextExtracts para `exintro`) e 4 lotes simultâneos por idioma (`WORKERS`), sob o throttle por host do `HttpClient` compartilhado entre threads (5 req/s por wiki). O REST `page/summary` (1 página por requisição) recebia 429 contínuo como cliente anônimo e ficava em ~1 página/s; em lote, ~1.400 páginas/min por idioma. Redirecionamentos são seguidos (`redirects=1`); a página só vale se o `wikibase_item` bater com o QID do lote e não for desambiguação. `store_records` grava cada bloco de 200 registros com `executemany` (pipeline, 1 round trip), não 1 upsert por registro, o que limitava um banco remoto a ~6 linhas/s. Para retomar um bootstrap sem refazer o Wikidata, dispare o workflow com `skip_catalog=true`.

Limites conhecidos: sinopses só são re-buscadas quando a entidade Wikidata muda ou falta registro; `affiliate_links` (GOG/Steam) ainda não é populado; Commons responde 429 em rajadas (o cliente espera `Retry-After`).

Fora por ora: orquestrador (Airflow/Prefect), ORM Python, dump JSON do Wikidata (`wbgetentities` em lotes de 50 cobre os ~100k itens em ~1h).
