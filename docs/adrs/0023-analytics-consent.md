# ADR 0023: analytics with one cookie banner (LGPD)

Status: accepted (2026-11-01). Decisions: owner request in board row ADOPT-vukem-0.36.0; ADR 0022; SDD 09 §7.8, SDD 13 §6.2. Framework: <https://vukem.dev/docs/digging-deeper/analytics>.

## Context

vukem v0.36.0 renders Umami, Google Analytics 4 and the Search Console and Bing verification tags from environment variables through the `{{analytics}}` view function. Umami sets no cookie and stores no personal data. GA4 sets `_ga` cookies, so LGPD needs the visitor's consent first. The app already has a cookie banner (`resources/js/components/CookieBanner.tsx`) whose first-party `consent` cookie (`all` or `essential`) gates the YouTube embeds.

## Decision

- One banner, one cookie. "Accept" (`consent=all`) allows YouTube embeds and GA4; "Reject" (`consent=essential`) refuses both. No per-purpose toggles: two optional purposes do not justify a settings dialog.
- GA4 runs in Consent Mode v2 with every signal denied by default (`GA_CONSENT_DEFAULT_GRANTED` stays unset). Google receives no request and `gtag.js` is not loaded until the banner calls `window.vukemAnalytics.consent(true)`. The stored choice is replayed on every full page load, because the GA bootstrap starts denied each time.
- Umami runs for every visitor regardless of the choice (cookieless, `data-do-not-track` honoured by default) and is named in the privacy policy.
- The footer's "Cookie settings" button reopens the banner, so the choice can be changed or withdrawn at any time; a reject after an accept calls `consent(false)`.
- The cookie stays readable by JavaScript and by the server (`SameSite=Lax`, one year): the server shares it as the `consent` prop and `RenderCache` keys on it.

## Consequences

- Analytics needs no new dependency and no new cookie.
- The privacy policy and terms (en, pt-BR, es) describe GA4, Umami and the settings link.
- Visitors who never answer the banner are counted by Umami only.
- A third optional purpose (for example ads) would need per-purpose choices and a new ADR.
