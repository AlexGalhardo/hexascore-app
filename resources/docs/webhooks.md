# Webhooks

Get a signed HTTP request when items change, instead of polling.

## Set up

Webhooks are available on the Pro (2 endpoints) and Business (10 endpoints) plans. Create an endpoint with `POST /v1/webhooks`:

| Field | Meaning |
| --- | --- |
| `url` | An `https` URL that answers within a few seconds. Private and internal addresses are refused. |
| `events` | One or more of `item.created`, `item.updated`, `item.deleted`. |
| `itemTypes` | One or more of `games`, `movies`, `series`, `books`. |

The response contains the signing `secret`, which starts with `whsec_`. It is shown once: store it in your secrets manager. Use `GET /v1/webhooks` to list endpoints, `DELETE /v1/webhooks/{id}` to remove one and `POST /v1/webhooks/{id}/test` to send a test delivery. Managing webhooks does not count against your quota.

## What we send

We send a batch of up to 100 events as one `POST` with a JSON body. Payloads carry ids and field names only; refetch the item to read its data.

```json
{
  "type": "item.batch",
  "deliveredAt": "2026-11-02T10:15:00Z",
  "events": [
    {
      "id": "chg_1048576",
      "type": "item.updated",
      "itemType": "games",
      "itemId": "0192f3a4-7b1c-7e21-9c55-3f1e2d4a6b71",
      "fields": ["score", "ratingCount"],
      "changedAt": "2026-11-02T10:14:12Z"
    },
    {
      "id": "chg_1048577",
      "type": "item.deleted",
      "itemType": "movies",
      "itemId": "0192f3a4-7b1c-7e21-9c55-3f1e2d4a5521",
      "mergedInto": "0192f3a4-7b1c-7e21-9c55-3f1e2d4a5520",
      "changedAt": "2026-11-02T10:14:30Z"
    }
  ]
}
```

Events for the same item inside one delivery window are coalesced. Event ids are unique, so use them to ignore a repeat.

## Verify the signature

Deliveries follow the [Standard Webhooks](https://www.standardwebhooks.com) scheme. Every request has three headers: `webhook-id`, `webhook-timestamp` (Unix seconds) and `webhook-signature` (one or more space-separated `v1,<base64>` values).

To verify a delivery:

- Read the raw request body. Do not parse and re-serialise it.
- Reject the request when `webhook-timestamp` is more than 5 minutes from your clock.
- Compute the base64 HMAC-SHA256 of `<webhook-id>.<webhook-timestamp>.<raw body>`, using the secret without its `whsec_` prefix, base64-decoded, as the key.
- Compare it with each `v1` value in constant time. Accept the request when one matches.

```ts
import { createHmac, timingSafeEqual } from "node:crypto";

export function verifyWebhook(secret: string, headers: Headers, rawBody: string): boolean {
	const id = headers.get("webhook-id");
	const timestamp = headers.get("webhook-timestamp");
	const signatures = headers.get("webhook-signature");
	if (!id || !timestamp || !signatures) return false;
	if (Math.abs(Date.now() / 1000 - Number(timestamp)) > 300) return false;

	const key = Buffer.from(secret.replace(/^whsec_/, ""), "base64");
	const expected = createHmac("sha256", key).update(`${id}.${timestamp}.${rawBody}`).digest();
	return signatures.split(" ").some((entry) => {
		const [version, value] = entry.split(",");
		if (version !== "v1" || !value) return false;
		const got = Buffer.from(value, "base64");
		return got.length === expected.length && timingSafeEqual(got, expected);
	});
}
```

```python
import base64
import hashlib
import hmac
import time


def verify_webhook(secret: str, headers: dict[str, str], raw_body: bytes) -> bool:
    try:
        msg_id = headers["webhook-id"]
        timestamp = headers["webhook-timestamp"]
        signatures = headers["webhook-signature"]
        if abs(time.time() - int(timestamp)) > 300:
            return False
    except (KeyError, ValueError):
        return False

    key = base64.b64decode(secret.removeprefix("whsec_"))
    signed = f"{msg_id}.{timestamp}.".encode() + raw_body
    expected = hmac.new(key, signed, hashlib.sha256).digest()
    for entry in signatures.split(" "):
        version, _, value = entry.partition(",")
        if version == "v1" and hmac.compare_digest(base64.b64decode(value), expected):
            return True
    return False
```

## Retries and failures

Answer with any `2xx` status within a few seconds, and do the real work afterwards. We retry other answers and timeouts 8 times with growing delays, starting at 30 seconds. A `410 Gone` stops the delivery for good. After three deliveries in a row run out of retries, the endpoint is disabled and we email the account owner. Deliveries are logged for 30 days.

Delivery is at least once and not strictly ordered. Webhooks can still miss events, so call [`/v1/changes`](/docs/changes) with the `since` of your last processed event to fill any gap.
