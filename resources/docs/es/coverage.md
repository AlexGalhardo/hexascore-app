# Cobertura de datos

Cada campo que devuelve la API proviene de una fuente que tenemos permiso para republicar. Por eso, algunos campos son irregulares y otros se omiten siempre. Esta página indica qué esperar antes de que construyas algo sobre ellos.

## Campos y sus fuentes

| Campo | Fuente | Vacío esperado |
| --- | --- | --- |
| `id`, `type`, `slug`, `url`, campos de puntuación, `trending`, `timeToBeat` | HexaScore (términos de la API) | ninguno |
| `title`, `releaseDate`, `genres`, `people`, `organisations`, `franchises`, `platforms`, `countries`, duración, páginas, temporadas, episodios | Wikidata (CC0) y la comunidad | duración, plataformas y cantidad de episodios son irregulares |
| `synopsis` | Wikipedia (CC BY-SA 4.0, con `sourceUrl`) | ítems sin artículo |
| ediciones y portadas de libros | Open Library y Wikidata | no aplica a otros tipos |
| portadas de películas, series y videojuegos, y tráilers | ninguna | nunca en la API |
| `externalIds` | Wikidata, ISBN-13, Open Library | los ids de TMDB, IGDB y Steam se omiten |

Cada respuesta enumera los campos que dejó fuera en `meta.omittedFields`, y cada valor con fuente lleva su `provenance` (fuente y licencia).

## Porcentajes de cobertura

La tabla siguiente indica, por tipo, la proporción de ítems públicos para los que la API devuelve cada campo. Se mide con `qa/sql/api_coverage.sql` sobre una copia del catálogo de producción, con las mismas reglas de fuente que usa la API, y muestra la fecha de la medición.
