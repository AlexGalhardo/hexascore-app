# Paginación

Los endpoints de lista usan cursores opacos: estables ante inserciones y rápidos a cualquier profundidad.

## Cómo funciona

`GET /v1/games`, `/v1/movies`, `/v1/series` y `/v1/books` devuelven `{ "data": [...], "meta": { "next_cursor": "..." } }`. Para obtener la página siguiente, repite la misma solicitud con `cursor=<meta.next_cursor>`. Cuando `meta.next_cursor` no está presente, llegaste al final.

- Mantén idénticos todos los demás parámetros. Un cursor de otro `sort` devuelve `400`.
- Los cursores son opacos: no los interpretes ni los construyas, su formato puede cambiar.
- El orden es estable: los empates se resuelven por el id del ítem, y los ítems sin puntuación o sin fecha siempre van al final.

## Tamaño de página

`limit` es 10 por defecto. El máximo depende de tu plan: 10 en Free, 100 en Pro y Business. Un valor mayor devuelve `400 invalid-query`; la API nunca trunca en silencio. Cada página cuenta como una solicitud.

## Ejemplo: recorrer todas las páginas

```js
let cursor = null;
do {
  const url = new URL("https://api.hexascore.app/v1/games");
  url.searchParams.set("limit", "100");
  if (cursor) url.searchParams.set("cursor", cursor);
  const res = await fetch(url, { headers: { Authorization: `Bearer ${process.env.HEXASCORE_API_KEY}` } });
  const page = await res.json();
  for (const game of page.data) console.log(game.title);
  cursor = page.meta.next_cursor ?? null;
} while (cursor);
```

## Búsqueda

`GET /v1/search` no está paginado: devuelve las mejores coincidencias (hasta 50) para una consulta, de la mejor a la peor.
