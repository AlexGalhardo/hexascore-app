# Mudanças

Consulte um único endpoint para manter a sua cópia do catálogo sincronizada, em vez de rastreá-lo inteiro de novo.

## Como funciona

`GET /v1/changes` retorna o que mudou desde um ponto no tempo, do mais antigo para o mais recente. Cada entrada diz qual item mudou, o que aconteceu com ele e quais campos se alteraram, nunca os novos valores: busque o item de novo com `GET /v1/{type}/{id}` para lê-los. Assim, a filtragem por licença e os campos do seu plano sempre se aplicam aos dados que você lê.

| Parâmetro | Significado |
| --- | --- |
| `since` | Obrigatório. Timestamp RFC 3339, de no máximo 14 dias atrás. |
| `type` | Opcional. `games`, `movies`, `series` ou `books`. |
| `cursor` | Opcional. O `meta.next_cursor` da página anterior. |
| `limit` | Opcional. De 1 a 1000, padrão 100. Não segue o tamanho de página do seu plano. |

Cada entrada tem este formato:

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

- `action` é `created`, `updated` ou `deleted`.
- `fields` lista os nomes dos campos visíveis na API que mudaram. Mudanças em campos ocultos não são listadas, e uma mudança de nota só é listada quando a nota com uma casa decimal, a confiança ou `ratingCount` mudou.
- Um duplicado mesclado é informado como `deleted`, com `mergedInto` apontando para o id que o substitui.

## Mantenha uma cópia sincronizada

Comece com um `since` de no máximo 14 dias atrás, siga `meta.next_cursor` até que ele esteja vazio, depois guarde o `changedAt` da última entrada e use-o como o próximo `since`. As entradas são mantidas por 30 dias, mas só os últimos 14 dias podem ser consultados; um `since` mais antigo retorna [`changes-window-exceeded`](/docs/errors#changes-window-exceeded) e um no futuro retorna [`invalid-query`](/docs/errors#invalid-query). Se você ficou offline por mais de 14 dias, reconstrua a partir das [exportações diárias de IDs](/docs/exports).

Cada página conta como uma requisição.

```ts
import { createHexascoreClient, getItem, listChanges } from "@hexascore/sdk";

const client = createHexascoreClient({ apiKey: process.env.HEXASCORE_API_KEY! });

let since = new Date(Date.now() - 24 * 3600 * 1000).toISOString();
let cursor: string | undefined;
do {
	const { data, error } = await listChanges({ client, query: { since, type: "games", limit: "500", cursor } });
	if (error) throw new Error(`${error.code}: ${error.detail}`);
	for (const change of data.data ?? []) {
		if (change.action === "deleted") continue; // remova o item ou siga mergedInto
		const item = await getItem({ client, path: { type: change.itemType, id: change.itemId } });
		// guarde item.data
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
                continue  # remova o item ou siga merged_into
            item = get_item.sync(change.item_type, client=c, id=change.item_id)
            since = change.changed_at
        cursor = page.meta.next_cursor
        if not cursor:
            break
```

## Webhooks em vez de polling

Os planos Pro e Business podem receber os mesmos eventos em uma URL. Veja [Webhooks](/docs/webhooks). As entregas de webhook podem falhar, então use `/v1/changes` para recuperar qualquer lacuna.
