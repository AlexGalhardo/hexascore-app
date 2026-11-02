# Exportaciones

Descarga la lista completa de ids de ítems una vez al día y luego obtén solo lo que necesites.

## Qué se exporta

Cada día, antes de las 08:00 UTC, generamos un archivo por tipo de ítem: games, movies, series y books. Los archivos se conservan 90 días. Las exportaciones están disponibles en los planes Pro y Business; el plan Free recibe [`plan-required`](/docs/errors#plan-required).

Cada archivo es un NDJSON comprimido con gzip, con un objeto JSON por línea:

```json
{"id":"0192f3a4-7b1c-7e21-9c55-3f1e2d4a6b71","slug":"hades","wikidata":"Q62640075","updatedAt":"2026-11-02T06:12:44Z"}
```

| Campo | Significado |
| --- | --- |
| `id` | El id del ítem en HexaScore, el que se usa en `GET /v1/{type}/{id}`. |
| `slug` | El slug de la URL de la página del ítem. |
| `wikidata` | El id de la entidad en Wikidata. |
| `updatedAt` | Cuándo cambió el ítem por última vez. |

Los archivos solo contienen ids. Los datos del ítem, las puntuaciones y los campos con restricciones de licencia salen de los endpoints habituales, por lo que siempre se aplican las reglas de tu plan.

## Descarga

`GET /v1/exports` lista los archivos disponibles. `GET /v1/exports/{type}/{day}` (por ejemplo `/v1/exports/games/2026-11-02`) responde `302` a una URL de descarga válida por 5 minutos. Sigue la redirección de inmediato y no guardes la URL.

```sh
curl -L -H "Authorization: Bearer $HEXASCORE_API_KEY" \
  "https://api.hexascore.app/v1/exports/games/2026-11-02" -o games_ids.json.gz
gunzip -c games_ids.json.gz | head -n 3
```

Un día desconocido, o con más de 90 días de antigüedad, devuelve [`not-found`](/docs/errors#not-found).

## Una buena rutina

- Descarga los archivos una vez al día y compáralos con tu copia: los ids nuevos son creaciones y los ids que faltan son eliminaciones.
- Obtén los ids nuevos con `GET /v1/{type}/{id}`, respetando tus [límites de solicitudes](/docs/rate-limits).
- Entre una exportación y otra, sigue [Cambios](/docs/changes) para tener la vista minuto a minuto.
