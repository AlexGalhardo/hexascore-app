# Ingestão de dados


| Fonte | Como | Fornece |
| --- | --- | --- |
| Wikidata | dump JSON (bootstrap) + SPARQL (incremental) | catálogo base, títulos por idioma, gêneros, datas, IDs externos, notas MC/RT (P444), trailer YouTube (P1651) |
| Wikipedia | REST `page/summary` | sinopse por idioma (CC BY-SA) |
| Open Library | dump + API | ISBN, páginas, autores, capas |
| Steam | `appdetails` público | preço, plataformas, trailers, capa |
| GOG | API pública de catálogo | disponibilidade, link |
| Epic, IMDb, MC, RT, HLTB, Goodreads, Skoob | sem fetch | deep-link pelo ID do Wikidata |
| TMDB, IGDB | fora da v1 | após licença comercial |

```
webscrappings/
├─ pyproject.toml   uv, ruff, pytest
├─ common/          http (rate limit, retry, User-Agent identificado), db (psycopg), upsert source_records
├─ connectors/      wikidata.py, wikipedia.py, openlibrary.py, steam.py, gog.py
│                   fetch(since) -> Iterable[Record(source, external_id, payload)]
├─ normalize.py     source_records → items / item_translations / external_scores (match por QID)
├─ covers.py        copia capas para Vercel Blob
└─ tests/           pytest com fixtures gravadas, sem rede
```

Padrão: conector grava payload bruto → `hash` igual = no-op → normalizador processa só alterados e faz upsert por `wikidata_qid` (sem QID = descartado na v1) → `POST /api/cron/revalidate` (segredo) invalida tags.

Agenda: **bootstrap** manual (`workflow_dispatch`): top 100k por sitelinks (~35k filmes, 15k séries, 25k jogos, 25k livros) · **incremental** noturno (Wikidata + Steam) · **semanal** notas externas + sinopses. Bootstrap exporta snapshot `.sql.gz` (GitHub Release) usado pelo seed de dev/preview.

Fora por ora: orquestrador (Airflow/Prefect), ORM Python.
