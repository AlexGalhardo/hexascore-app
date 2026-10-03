---
name: hexascore-secure-endpoint
description: Checklist for any Hexascore Server Action, route handler, webhook, cron or data that reaches the browser. Use before adding or changing one.
---

# Secure endpoint checklist

- Identity only from `getSessionUser()` (`apps/web/lib/session.ts`); writes that need a member go through `lib/member.ts#asMember` (session, not banned, verified email, nickname). Never trust ids from the body.
- `"use server"` only in `apps/web/actions/`; those files export only async functions and return typed `ActionResult` (`lib/action-result.ts`), never throw to the client.
- Validate every input with zod (`lib/validation.ts` or a local schema); parse jsonb on read.
- Return DTOs, never whole `user` rows; money and quota computed on the server.
- Public API routes use `apiRoute()` (`lib/api/handler.ts`): strict query schema, response schema, RFC 9457 errors.
- Webhooks: verify signature/HMAC in constant time, idempotency via `processed_events`, 400 bad signature / 200 duplicate / 500 transient.
- Crons: `isAuthorizedCron(request)` (`lib/cron.ts`).
- Outbound calls: `fetchWithTimeout` (`lib/http.ts`).
- Rate limit abuse-prone actions (`lib/rate-limit.ts`); BotID on public forms (`lib/bot.ts`).
- Add a row to the surfaces table in `docs/security.md` and a test for the authorization path.
