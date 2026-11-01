# Privacy policy review

Status: draft for the owner. This file is not a legal opinion.

## What P2.D.4.1 changed

The public policy in `resources/lang/{en,pt-BR,es}.json` (`policy`, rendered by `Legal/Page`) now names:

- processors: Railway, Resend, Stripe, Google, Telegram
- cookies: session, two-factor, `locale`, `theme`, `consent`
- retention: session IP and user agent 30 days, opt-in client reports 7 days, export files 24 hours, account deletion grace 30 days
- minimum age 13
- encarregado contact: `privacy@hexascore.app`

It no longer names Vercel, Neon, Upstash, passkeys, or AbacatePay.

The daily job `privacy.retain` (`app/privacy/retention.go`) enforces the session, client-report, and export periods. `SESSION_RECORD_CLIENT=true` is set in `.env.example` and the compose app env so a logged-in session stores the truncated IP the job later clears.

## Human gate (still open)

D-21 and D-29: a lawyer has not reviewed this text. The owner reviews it before cutover.

D-39: the encarregado is the owner. The public contact is `privacy@hexascore.app`. No personal name is published.

LGPD article text was not fetched (SDD 13 issue 5). This policy does not quote articles.

## Not run here

An export link with a bad signature, or used after 24 hours, was not curled against staging. `TestLGPD_ExportDeleteOptOut` covers both cases on the local app.
