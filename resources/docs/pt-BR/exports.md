# Exportações

Baixe a lista completa de ids de itens uma vez por dia e busque depois apenas o que precisar.

## O que é exportado

Todo dia, antes das 08:00 UTC, geramos um arquivo por tipo de item: games, movies, series e books. Os arquivos são mantidos por 90 dias. As exportações estão disponíveis nos planos Hobby e PRO; o plano Free recebe [`plan-required`](/docs/errors#plan-required).

Cada arquivo é um NDJSON compactado com gzip, com um objeto JSON por linha:

```json
{"id":"0192f3a4-7b1c-7e21-9c55-3f1e2d4a6b71","slug":"hades","wikidata":"Q62640075","updatedAt":"2026-11-02T06:12:44Z"}
```

| Campo | Significado |
| --- | --- |
| `id` | O id do item no HexaScore, usado em `GET /v1/{type}/{id}`. |
| `slug` | O slug da URL da página do item. |
| `wikidata` | O id da entidade no Wikidata. |
| `updatedAt` | Quando o item mudou pela última vez. |

Os arquivos trazem apenas ids. Os dados do item, as notas e os campos com restrição de licença vêm dos endpoints normais, então as regras do plano sempre se aplicam.

## Download

`GET /v1/exports` lista os arquivos disponíveis. `GET /v1/exports/{type}/{day}` (por exemplo `/v1/exports/games/2026-11-02`) responde `302` para uma URL de download válida por 5 minutos. Siga o redirecionamento imediatamente e não guarde a URL.

```sh
curl -L -H "Authorization: Bearer $HEXASCORE_API_KEY" \
  "https://api.hexascore.app/v1/exports/games/2026-11-02" -o games_ids.json.gz
gunzip -c games_ids.json.gz | head -n 3
```

Um dia desconhecido, ou com mais de 90 dias, retorna [`not-found`](/docs/errors#not-found).

## Uma boa rotina

- Baixe os arquivos uma vez por dia e compare com a sua cópia: ids novos são criações, ids ausentes são exclusões.
- Busque os ids novos com `GET /v1/{type}/{id}`, respeitando os seus [limites de requisições](/docs/rate-limits).
- Entre uma exportação e outra, acompanhe as [Mudanças](/docs/changes) para ter a visão minuto a minuto.
