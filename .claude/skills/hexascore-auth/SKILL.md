---
name: hexascore-auth
description: Work on Hexascore auth and account features (Better Auth 1.7 + BotID + Server Actions) and test them locally. Use when touching apps/web/lib/auth.ts, lib/session.ts, actions/account.ts, actions/contact.ts, the /login /signup /profile /contact pages, the purge cron, or when you need to sign up/verify/2FA a user on a local dev server.
---

# Hexascore auth & account

## Where things live

- `lib/auth.ts` — Better Auth config. `hooks.before` enforces `passwordSchema` on every endpoint that accepts a password (`PASSWORD_PATHS`) and runs BotID on `BOT_PROTECTED_AUTH_PATHS` (only for real HTTP requests: `ctx.request` is undefined for `auth.api.*` calls).
- `lib/session.ts` — `getSessionUser()` is the only identity source. Add fields to `SessionUser` there (DTO), never pass the raw Better Auth user to the client.
- `actions/*.ts` — the only `"use server"` files. Contract: zod-validate the input, `getSessionUser()` for identity, return `ActionResult` (`lib/action-result.ts`), never throw to the client. Add new error codes to `ActionError` **and** to `messages/<locale>/account.json` → `actions`.
- Client flows (sign-in, sign-up, 2FA, passkeys, change password, reset) use `authClient` and map errors with `authErrorKey()` (`lib/auth-errors.ts`) to `auth.errors.*`.

## Gotchas

- **Server-only endpoints**: `setPassword` (and other `createAuthEndpoint.serverOnly`) are not reachable over HTTP. Call `auth.api.setPassword({ body, headers: await headers() })` from a Server Action. `nextCookies()` makes cookie changes (e.g. `auth.api.signOut`) work inside actions.
- `twoFactor.enable` returns a union; pass `method: "totp"` and narrow on `data.method === "totp"` to read `totpURI`/`backupCodes`. `allowPasswordless: true` lets Google/passkey-only accounts enable 2FA without a password.
- Callback URLs must be localized paths: `getPathname({ href: "/verify-email", locale })`. A localized path redirected with next-intl's `redirect` gets a second prefix — use `next/navigation`'s `redirect(path as Route)` for already-localized strings.
- Email templates must be passed as `createElement(Template, props)` (not `Template(props)`), otherwise the dev log cannot read `url` from the element props.
- Long legal text lives in `messages/<locale>/legal.json` and is **not** in `NAMESPACES` (would ship to every page). Load it with `loadLegal()`.

## BotID

- Client: `instrumentation-client.ts` `initBotId({ protect })` must list every path the server checks (Better Auth endpoints + the page paths that Server Actions POST to, per locale).
- Server: `isBotRequest()` in `lib/bot.ts`. Off Vercel (`VERCEL !== "1"`, includes local `next start`) it passes `isDevelopment: true`, so BotID bypasses as human (it would otherwise throw for a missing OIDC token). On Vercel, an exception denies.
- `next.config.ts` is wrapped with `withBotId(...)` for the challenge rewrites.

## Local verification

```bash
cd apps/web
# .env.local: BETTER_AUTH_SECRET / CRON_SECRET = openssl rand -hex 32, DATABASE_URL="", BETTER_AUTH_URL and NEXT_PUBLIC_APP_URL = http://localhost:<port>
bun run db:migrate
bun --bun next dev -p 3123 > dev.log 2>&1
```

- Without `RESEND_API_KEY`, every email prints `[email:dev] to=<email> subject="..." link=<url>` in the server log: open that URL to verify/reset.
- Better Auth POSTs need `Origin: http://localhost:<port>` when called with curl.
- TOTP code from the `secret` query param of `totpURI` (base32 → HMAC-SHA1 of `floor(now/30s)`, RFC 6238, 6 digits). The profile page shows the secret in a `<code>` next to the QR.
- Rate limit: 4th+ rapid `/sign-in/email` within 10 s returns 429.
- Cron: `curl -H "Authorization: Bearer $CRON_SECRET" localhost:3123/api/cron/purge-accounts` → `{ "purged": n }`.
- Do not run `next build` while a dev server holds `.pglite`; build workers use in-memory PGlite (`NEXT_PHASE`), but a corrupted `.pglite` (WASM "Aborted()" or failing `CREATE SCHEMA`) is fixed by deleting the folder and re-running `db:migrate`.
