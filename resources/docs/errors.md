# Errors

Errors use RFC 9457 problem details with a stable machine-readable code.

## Format

Every error has `Content-Type: application/problem+json`. Branch on `code` (stable) rather than on `title` or `detail` (human text that may change). `type` links to the matching section below, and `request_id` equals the `X-Request-Id` header: include it when you contact support.

```json
{
  "type": "https://hexascore.app/problems/quota-exhausted",
  "title": "Quota exhausted",
  "status": 402,
  "detail": "Monthly quota is exhausted. Upgrade at https://hexascore.app/pricing.",
  "instance": "/api/v1/games",
  "request_id": "01JB7Q4V0Q8M3K2Y7T9H5N6C1D",
  "code": "quota-exhausted",
  "plan": "pro",
  "limit": 50000
}
```

## Codes

### invalid-query

HTTP 400. A query parameter is unknown or out of range, or `limit` is above your plan's page size (10 on Free). `errors` lists each problem with its `path`. Fix the listed parameters. Not counted against your quota.

### invalid-cursor

HTTP 400. The `cursor` was tampered with or belongs to another query. Restart from the first page.

### changes-window-exceeded

HTTP 400. The `since` window of `/v1/changes` is larger than the endpoint allows. Request a more recent `since`.

### unauthorized

HTTP 401. No `Authorization: Bearer` header, a malformed key, or a key that does not exist or was revoked. Check the header and the key at [/profile/api](/profile/api). Never retried automatically.

### quota-exhausted

HTTP 402. The 100 lifetime requests on Free, or the monthly quota on paid plans. The body includes `plan` and `limit`. Upgrade your plan at [/pricing](/pricing). Paid quotas reset on the 1st of each month (UTC).

### plan-required

HTTP 403. The endpoint needs a higher plan. Upgrade at [/pricing](/pricing).

### account-banned

HTTP 403. The account that owns the key is suspended. Contact us through [/contact](/contact) with the `request_id`.

### not-found

HTTP 404. Unknown item id, an id that belongs to another type (for example a game id under `/movies`), an unknown type path, or no match for `/lookup`. Lookups require an exact id.

### rate-limited

HTTP 429. More requests per second than your plan allows. The `Retry-After` header says how many seconds to wait. Not counted against your quota. Space out bursts or upgrade for a higher rate.

### internal-error

HTTP 500. Something failed on our side. The request is refunded to your quota. Retry with exponential backoff; if it persists, send us the `request_id`.

### community-submissions-closed

HTTP 403. Community suggestions are switched off right now. Try again later.

### suggestion-not-eligible

HTTP 403. Your account cannot submit suggestions (unverified email, too new, banned from submissions, or the suggestion is not yours).

### suggestion-duplicate

HTTP 409. An open suggestion for the same item and fields already exists. Edit or withdraw it instead.

### stale-before

HTTP 409. The field changed since you loaded it. Reload and re-apply your edit.

### invalid-transition

HTTP 409. The suggestion cannot move from its current status to the requested one.

### field-locked

HTTP 423. A moderator locked this field; community edits to it are not accepted.

### suggestion-quota

HTTP 429. You reached your daily suggestion quota. Try again after the reset.

### suggestion-cooldown

HTTP 429. Wait for the cooldown to pass; the `Retry-After` header gives the seconds.

### validation

HTTP 422. The submission failed validation.

### service-unavailable

HTTP 503. A dependency is temporarily down. Comes with `Retry-After`. Retry after the indicated delay.

## Retries

Retry only `429`, `500` and `503`, honouring `Retry-After` when present and backing off exponentially otherwise. Every endpoint is a read-only `GET`, so retries are always safe.
