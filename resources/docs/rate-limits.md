# Rate limits

Each plan has a request quota, a per-second rate limit and a maximum page size.

## Plans

| Plan | Quota | Requests / second | Max page size |
| --- | --- | --- | --- |
| Free | 100 in total | 1 | 10 |
| Pro | 50,000 / month | 10 | 100 |
| Business | 500,000 / month | 50 | 100 |

Prices and upgrades: [/pricing](/pricing).

## What counts

- Every authenticated request that reaches an endpoint counts as one, including 404s.
- Not counted: `401`, `400` for invalid parameters, `429` and any `5xx` (refunded).
- `GET /v1/me/usage` counts too, so read the response headers of your regular calls when you can.
- Paid quotas reset at 00:00 UTC on the 1st of each month. The Free quota never resets.

## Response headers

Successful responses tell you where you stand (IETF RateLimit header fields).

| Header | Meaning |
| --- | --- |
| `RateLimit-Limit` | Requests in the current period (your plan quota). |
| `RateLimit-Remaining` | Requests left before 402. |
| `RateLimit-Reset` | Seconds until the quota resets; 0 for the lifetime Free quota. |
| `Retry-After` | On 429 and 503: seconds to wait before retrying. |

## When you hit a limit

Too many requests per second returns [`429 rate-limited`](/docs/errors#rate-limited): wait `Retry-After` seconds. Quota exhausted returns [`402 quota-exhausted`](/docs/errors#quota-exhausted): upgrade your plan; retrying will not help until the reset.
