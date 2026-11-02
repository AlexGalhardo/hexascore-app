# Cambios

Consulta un único endpoint para mantener tu copia del catálogo sincronizada, en lugar de rastrearlo completo otra vez.

## Cómo funciona

`GET /v1/changes` devuelve lo que cambió desde un momento dado, del más antiguo al más reciente. Cada entrada indica qué ítem cambió, qué le ocurrió y qué campos se modificaron, nunca los valores nuevos: vuelve a obtener el ítem con `GET /v1/{type}/{id}` para leerlos. Así, el filtrado por licencia y los campos de tu plan siempre se aplican a los datos que lees.

| Parámetro | Significado |
| --- | --- |
| `since` | Obligatorio. Marca de tiempo RFC 3339, de como máximo 14 días atrás. |
| `type` | Opcional. `games`, `movies`, `series` o `books`. |
| `cursor` | Opcional. El `meta.next_cursor` de la página anterior. |
| `limit` | Opcional. De 1 a 1000, 100 por defecto. No sigue el tamaño de página de tu plan. |

Cada entrada tiene este aspecto:

```json
{
  "id": "chg_1048576",
  "action": "updated",
  "itemType": "games",
  "itemId": "0192f3a4-7b1c-7e21-9c55-3f1e2d4a6b71",
  "fields": ["score", "ratingCount"],
  "changedAt": "2026-11-02T10:14:12Z"
}
```

- `action` es `created`, `updated` o `deleted`.
- `fields` enumera los nombres de los campos visibles en la API que cambiaron. Los cambios en campos ocultos no se listan, y un cambio de puntuación solo se lista cuando cambió la puntuación con un decimal, la confianza o `ratingCount`.
- Un duplicado fusionado se informa como `deleted`, con `mergedInto` apuntando al id que lo reemplaza.

## Mantén una copia sincronizada

Empieza con un `since` de como máximo 14 días atrás, sigue `meta.next_cursor` hasta que esté vacío, luego guarda el `changedAt` de la última entrada y úsalo como el siguiente `since`. Las entradas se conservan 30 días, pero solo se pueden consultar los últimos 14; un `since` más antiguo devuelve [`changes-window-exceeded`](/docs/errors#changes-window-exceeded) y uno en el futuro devuelve [`invalid-query`](/docs/errors#invalid-query). Si estuviste sin conexión más de 14 días, reconstruye a partir de las [exportaciones diarias de IDs](/docs/exports).

Cada página cuenta como una solicitud.

```ts
import { createHexascoreClient, getItem, listChanges } from "@hexascore/sdk";

const client = createHexascoreClient({ apiKey: process.env.HEXASCORE_API_KEY! });

let since = new Date(Date.now() - 24 * 3600 * 1000).toISOString();
let cursor: string | undefined;
do {
	const { data, error } = await listChanges({ client, query: { since, type: "games", limit: "500", cursor } });
	if (error) throw new Error(`${error.code}: ${error.detail}`);
	for (const change of data.data ?? []) {
		if (change.action === "deleted") continue; // elimina el ítem o sigue mergedInto
		const item = await getItem({ client, path: { type: change.itemType, id: change.itemId } });
		// guarda item.data
		since = change.changedAt;
	}
	cursor = data.meta?.next_cursor || undefined;
} while (cursor);
```

```python
import os
from datetime import datetime, timedelta, timezone

from hexascore import AuthenticatedClient
from hexascore.api.catalog import get_item, list_changes
from hexascore.models import Problem
from hexascore.types import UNSET

client = AuthenticatedClient(base_url="https://api.hexascore.app/v1", token=os.environ["HEXASCORE_API_KEY"])
since = (datetime.now(timezone.utc) - timedelta(days=1)).isoformat()

with client as c:
    cursor = UNSET
    while True:
        page = list_changes.sync(client=c, since=since, type_="games", limit="500", cursor=cursor)
        if isinstance(page, Problem) or page is None:
            raise RuntimeError(page.detail if page else "no response")
        for change in page.data:
            if change.action == "deleted":
                continue  # elimina el ítem o sigue merged_into
            item = get_item.sync(change.item_type, client=c, id=change.item_id)
            since = change.changed_at
        cursor = page.meta.next_cursor
        if not cursor:
            break
```

## Webhooks en lugar de polling

Los planes Hobby y PRO pueden recibir los mismos eventos en una URL. Consulta [Webhooks](/docs/webhooks). Las entregas de webhooks pueden perderse, así que usa `/v1/changes` para recuperar cualquier vacío.
