# Authentication

Every request is authenticated with a secret API key sent as a Bearer token.

## Bearer token

Send the key in the `Authorization` header. Keys look like `hx_live_` followed by 32 letters and digits. Query-string keys are not accepted, so keys never end up in access logs or browser history.

```sh
curl -H "Authorization: Bearer hx_live_..." https://api.hexascore.app/v1/games
```

A missing, malformed, unknown or revoked key returns `401` with the [`unauthorized`](/docs/errors#unauthorized) code and a `WWW-Authenticate: Bearer` header.

## Managing keys

- Create and revoke keys at [/profile/api](/profile/api) (verified email required, up to 5 active keys).
- The full key is shown once. Only a SHA-256 hash is stored; we cannot recover it for you.
- Revocation is immediate. Rotate by creating a new key, deploying it, then revoking the old one.
- All keys of an account share the same plan, quota and per-second rate limit.

## Keep keys secret

Call the API from your server. A key embedded in a web page or mobile app can be extracted and used to burn your quota. If a key leaks, revoke it right away.
