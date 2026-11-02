# Cobertura de dados

Todo campo que a API retorna vem de uma fonte que temos permissão para republicar. Por isso, alguns campos são irregulares e outros são sempre omitidos. Esta página diz o que esperar antes de você construir algo em cima deles.

## Campos e suas fontes

| Campo | Fonte | Lacuna esperada |
| --- | --- | --- |
| `id`, `type`, `slug`, `url`, campos de nota, `trending`, `timeToBeat` | HexaScore (termos da API) | nenhuma |
| `title`, `releaseDate`, `genres`, `people`, `organisations`, `franchises`, `platforms`, `countries`, duração, páginas, temporadas, episódios | Wikidata (CC0) e a comunidade | duração, plataformas e contagem de episódios são irregulares |
| `synopsis` | Wikipedia (CC BY-SA 4.0, com `sourceUrl`) | itens sem artigo |
| edições e capas de livros | Open Library e Wikidata | não se aplica a outros tipos |
| capas de filmes, séries e jogos, e trailers | nenhuma | nunca na API |
| `externalIds` | Wikidata, ISBN-13, Open Library | os ids do TMDB, IGDB e Steam são omitidos |

Toda resposta lista os campos que deixou de fora em `meta.omittedFields`, e cada valor com fonte traz a sua `provenance` (fonte e licença).

## Percentuais de cobertura

A tabela abaixo mostra, por tipo, a fatia de itens públicos para os quais a API retorna cada campo. Ela é medida com `qa/sql/api_coverage.sql` em uma cópia do catálogo de produção, usando as mesmas regras de fonte da API, e exibe a data da medição.
