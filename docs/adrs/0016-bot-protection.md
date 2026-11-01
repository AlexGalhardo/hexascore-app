# ADR 0016: Bot and abuse protection without Vercel BotID

Status: accepted (2026-10-28), amended 2026-11-01 (Turnstile in production, see the end). Decisions: brainstorming D-05, D-08; SDD 03 open issue 4; `vukem-feedback.md` VF-003.

## Context

The legacy app relied on Vercel BotID (client script plus server check, failing closed only on Vercel) for signup, login and the contact form. On Railway there is no equivalent. vukem offers GCRA rate limiting by IP and token, login throttling keyed by account and IP, and secure defaults, but no challenge mechanism (`https://vukem.dev/docs/security/api-tokens` "Rate limiting", `https://vukem.dev/docs/security/rate-limiting` "Rate limiting"). The budget ceiling (D-08) and the "native tools first" rule (requirement 1) argue against a paid bot-detection vendor.

## Decision

Layered, server-side, in this order:

1. **Rate limits** on every unauthenticated write route (signup, login, password reset, contact, suggestion submission) by IP and, where it exists, by account: limits defined in SDD 13; failures answer 429 with `Retry-After`.
2. **Form integrity**: honeypot field, minimum fill time, signed form token bound to the session, `Sec-Fetch-Site` enforcement (vukem CSRF default).
3. **Email verification before any write** that reaches other users (ratings count only after verification; suggestions need a verified email and 7-day account age, SDD 04).
4. **Escalating challenge** only when a threshold trips (per IP or per ASN velocity): a stdlib proof-of-work challenge (hash-cash style, 100–300 ms of client CPU) served by the Go app; optional Cloudflare Turnstile adapter behind a config flag for when proof-of-work is insufficient (adds a third-party script only on challenged requests). Turnstile requires its own ADR update if enabled in production.
5. **Signals for the scoring and moderation layers**: account age, verification state, challenge history and rate-limit hits feed the weights in SDD 02 and the anti-abuse tiers in SDD 04.

## Alternatives considered

- **Cloudflare Turnstile everywhere.** Robust and free tier available, but a third-party script on every auth page and a dependency on Cloudflare; kept as the optional escalation step.
- **reCAPTCHA / hCaptcha.** Privacy and UX costs; rejected.
- **Paid bot-detection vendors.** Cost over the ceiling; rejected.
- **Nothing beyond rate limits.** Insufficient for signup spam that poisons scores; rejected.

## Consequences

- Positive: no new runtime dependency by default; works on any host; signals reused by scoring and moderation.
- Negative: proof-of-work does not stop well-funded abuse; the escalation to Turnstile must be easy to flip.
- Obligations: SDD 13 defines thresholds and tests (`TestSecurity_*` names); `vukem-feedback.md` VF-003 asks for a framework challenge hook.

## Amendment 2026-11-01: Cloudflare Turnstile in production

vukem v0.36.0 ships the hook VF-003 asked for (`challenge.TurnstileConfig`, `challenge.StepUp`; <https://vukem.dev/docs/security/bot-challenge>) and the owner issued production keys (board row ADOPT-vukem-0.36.0).

- Turnstile is on exactly when `TURNSTILE_SITE_KEY` and `TURNSTILE_SECRET_KEY` are both set (one without the other refuses to boot). Unset, every form and page behaves as before.
- When on, `POST /signup`, `/login` and `/contact` need a token solved for that form (actions `register`, `login`, `contact`), checked before the proof-of-work gate, which stays as the second layer. The `api.js` script loads only on those three pages (and from the widget after a client-side visit), not site-wide.
- The page guard (ADR 0022) adds a soft per-IP budget (`pages.soft`, 30/min) below the hard limits: past it an anonymous client solves Turnstile once at `POST /turnstile` and gets a 30 min clearance cookie; verified search crawlers are never asked.
- Siteverify outages fail closed (503); `TURNSTILE_INSECURE_FAIL_OPEN` stays unset.
- Tests and local e2e use Cloudflare's published test keys and a local fake siteverify (`TURNSTILE_VERIFY_URL`, which refuses to boot in production).
