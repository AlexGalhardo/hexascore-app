# Changes

Poll one endpoint to keep your copy of the catalogue in sync, instead of crawling it again.

## How it works

`GET /v1/changes` returns what changed since a point in time, oldest first. Each entry says which item changed, what happened to it and which fields moved, never the new values: refetch the item with `GET /v1/{type}/{id}` to read them. That way licence filtering and your plan's fields always apply to the data you read.

| Parameter | Meaning |
| --- | --- |
| `since` | Required. RFC 3339 timestamp, at most 14 days ago. |
| `type` | Optional. `games`, `movies`, `series` or `books`. |
| `cursor` | Optional. The `meta.next_cursor` of the previous page. |
| `limit` | Optional. 1 to 1000, default 100. This does not follow your plan's page size. |

Each entry looks like this:

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

- `action` is `created`, `updated` or `deleted`.
- `fields` lists the API-visible field names that changed. Changes to hidden fields are not listed, and a score change is listed only when the one-decimal score, the confidence or `ratingCount` moved.
- A merged duplicate is reported as `deleted` with `mergedInto` set to the id that replaces it.

## Keep a copy in sync

Start with a `since` no older than 14 days, follow `meta.next_cursor` until it is empty, then store the `changedAt` of the last entry and use it as the next `since`. Entries are kept for 30 days but only the last 14 days can be queried; an older `since` returns [`changes-window-exceeded`](/docs/errors#changes-window-exceeded) and a future one returns [`invalid-query`](/docs/errors#invalid-query). If you were offline for longer than 14 days, rebuild from the [daily ID exports](/docs/exports).

Every page counts as one request.

```ts
import { createHexascoreClient, getItem, listChanges } from "@hexascore/sdk";

const client = createHexascoreClient({ apiKey: process.env.HEXASCORE_API_KEY! });

let since = new Date(Date.now() - 24 * 3600 * 1000).toISOString();
let cursor: string | undefined;
do {
	const { data, error } = await listChanges({ client, query: { since, type: "games", limit: "500", cursor } });
	if (error) throw new Error(`${error.code}: ${error.detail}`);
	for (const change of data.data ?? []) {
		if (change.action === "deleted") continue; // drop the item, or follow mergedInto
		const item = await getItem({ client, path: { type: change.itemType, id: change.itemId } });
		// store item.data
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
                continue  # drop the item, or follow merged_into
            item = get_item.sync(change.item_type, client=c, id=change.item_id)
            since = change.changed_at
        cursor = page.meta.next_cursor
        if not cursor:
            break
```

## Webhooks instead of polling

Pro and Business plans can have the same events pushed to a URL. See [Webhooks](/docs/webhooks). Webhook deliveries can be missed, so use `/v1/changes` to recover any gap.
