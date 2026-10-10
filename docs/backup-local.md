# Backup local do catálogo em JSON: quanto pesa

Estimativa do espaço em disco para guardar na máquina local, em JSON, as informações das 4 categorias (filmes, séries, jogos, livros), como cópia de segurança do banco de produção (Neon, ligado ao projeto na Vercel).

Medido em produção em **2026-10-10**, somente leitura, com `octet_length(to_jsonb(linha)::text)` por linha (bytes exatos do JSON em UTF-8, sem a coluna gerada `search`). A taxa de gzip vem de uma amostra de 1.500 itens por tipo (nível 6). A rodada do TMDB ainda estava em andamento na medição; a seção [O que ainda vai crescer](#o-que-ainda-vai-crescer) projeta o valor final.

## Resumo

| O que guardar | JSON puro | JSON + gzip |
| --- | --- | --- |
| Catálogo (itens + traduções/sinopses + notas externas) | **201 MB** | **~55 MB** |
| + registros brutos das fontes (`source_records`) | +230 MB (≈ 256 MB com o TMDB completo) | +~58 MB (≈ 68 MB) |
| + imagens de capa (WebP no Vercel Blob) | +~1,5 GB (78.904 capas após a rodada TMDB/IGDB) | não comprime |
| **Total com tudo** | **~2,0 GB** | **~1,7 GB** |

Só o texto do catálogo cabe em menos de 60 MB compactado. O que pesa são as capas.

## Catálogo por categoria

Um documento por item, com as traduções e as notas externas embutidas.

| Categoria | Itens | `items` | `item_translations` | `external_scores` | Total JSON | Média por item | gzip (taxa) | Total gzip |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Filmes | 34.968 | 25,1 MB | 49,8 MB (84.232 linhas, 72.561 sinopses) | 5,2 MB (26.726 linhas) | **80,2 MB** | 2,3 KB | 0,271 | ~21,7 MB |
| Séries | 14.911 | 8,9 MB | 19,6 MB (31.616 linhas, 28.217 sinopses) | 0,02 MB (127 linhas) | **28,6 MB** | 1,9 KB | 0,286 | ~8,2 MB |
| Jogos | 24.873 | 16,9 MB | 33,5 MB (74.244 linhas, 37.173 sinopses) | 0,06 MB (394 linhas) | **50,5 MB** | 2,0 KB | 0,250 | ~12,6 MB |
| Livros | 24.089 | 16,3 MB | 25,9 MB (41.605 linhas, 33.328 sinopses) | 0,001 MB (4 linhas) | **42,1 MB** | 1,7 KB | 0,287 | ~12,1 MB |
| **Total** | **98.841** | 67,2 MB | 128,9 MB | 5,3 MB | **201,4 MB** | 2,0 KB | | **~54,6 MB** |

As sinopses (em `item_translations`) são quase dois terços do peso.

## Registros brutos das fontes (`source_records`)

Opcional. Com eles dá para reconstruir o catálogo do zero com `normalize`, sem depender das APIs de origem. Tamanho do texto JSON estimado por amostra de 300 registros por fonte.

| Fonte | Registros | JSON | gzip (taxa) |
| --- | --- | --- | --- |
| `wikidata` | 99.960 | 101,8 MB | 0,135 |
| `wikipedia:en` | 90.569 | 55,1 MB | 0,370 |
| `wikipedia:es` | 45.837 | 27,3 MB | 0,358 |
| `igdb` | 22.848 | 18,6 MB | 0,322 |
| `wikipedia:pt` | 35.107 | 18,5 MB | 0,354 |
| `tmdb` (parcial: 9.000 de 49.239) | 9.000 | 5,8 MB | 0,332 |
| `steam` | 6.049 | 1,3 MB | 0,188 |
| `openlibrary` | 10.187 | 0,9 MB | 0,228 |
| `gog` | 1.974 | 0,6 MB | 0,261 |
| **Total** | 321.531 | **229,9 MB** | **~58 MB** |

## Capas

As imagens não estão no banco: ficam no Vercel Blob (`hexascore-covers`) e o banco guarda só a URL. Um backup completo precisa baixá-las.

- Média medida em `docs/ingestion.md` (12 capas reais): 18,2 KB a capa + 1,8 KB a miniatura = **20 KB por item**.
- Antes da rodada TMDB/IGDB: 27.469 itens com capa × 20 KB ≈ 0,55 GB.
- Depois da rodada de 2026-10-10: **78.904 itens com capa** (34.657 filmes, 13.300 séries, 22.799 jogos, 8.148 livros) × 20 KB ≈ **1,5 GB**. Os blobs das capas antigas substituídas ou removidas continuam no store (órfãos), então o store em si ocupa mais que isso. Pôsteres verticais podem ficar acima da média de 18,2 KB, que foi medida em capas de livro e headers da Steam.

WebP já é comprimido: gzip não reduz.

## O que ainda vai crescer

- **TMDB completo**: ~49 mil registros × 672 B ≈ 31,6 MB em `source_records` (+25,8 MB sobre a medição).
- **Campos preenchidos por TMDB/IGDB** (gêneros, criadores, duração, plataformas): alguns bytes por item, bem abaixo de 5% do catálogo.
- **Sinopses próprias** para itens sem Wikipedia (passo ainda não escrito): na ordem de 600 B por sinopse, o mesmo tamanho médio de um extrato da Wikipedia.

## Projeção por volume

Média atual de 2.137 B de JSON por item (texto do catálogo, sem `source_records` e sem capas):

| Itens | JSON puro | JSON + gzip | Capas (20 KB cada, se todos tiverem) |
| --- | --- | --- | --- |
| 100 mil (hoje) | 0,2 GB | 0,05 GB | 2 GB |
| 500 mil | 1,0 GB | 0,27 GB | 10 GB |
| 1 milhão | 2,0 GB | 0,55 GB | 20 GB |

## O que este cálculo não cobre

- **Dados de usuários**: `ratings` (591 MB), `item_status` (356 MB), `list_items` (86 MB), `time_to_beat`, `follows`, `lists` e `user`. Hoje são quase todos do seed de demonstração. O banco inteiro tem 1.923 MB; o catálogo (`items`, `item_translations`, `external_scores`, `source_records`) ocupa ~823 MB disso, contando índices e a coluna de busca, que não entram no JSON.
- **Um exportador JSON**: ainda não existe. O que existe é o snapshot SQL (`uv run python -m bootstrap --export-only --export snapshot.sql.gz`, ver `docs/ingestion.md`), que cobre `items`, `item_translations` e `external_scores`, mas não `source_records` nem as capas.

## Formato sugerido

NDJSON (um documento JSON por linha), um arquivo por categoria, compactado com gzip:

```text
backup/2026-10-10/
├─ movies.ndjson.gz           ~22 MB
├─ series.ndjson.gz           ~8 MB
├─ games.ndjson.gz            ~13 MB
├─ books.ndjson.gz            ~12 MB
├─ source_records.ndjson.gz   ~58 a 68 MB (opcional)
└─ covers/                    ~1,5 GB (opcional)
```

Cada linha precisa manter `id` (uuid) e `slug` do item: as avaliações e as URLs públicas dependem deles, então uma restauração não pode gerar novos.
