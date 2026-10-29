# E-008 legacy runtime dependencies (source: legacy/apps/web/package.json, legacy/sdks/*)

Usage = count of .ts/.tsx files under apps/web importing the package (grep of the quoted name; node_modules excluded).

## Auth
| package | pinned | used by (examples) | files |
| --- | --- | --- | --- |
| better-auth | 1.7.7 | lib/auth.ts, lib/auth-client.ts, app/api/auth/[...all]/route.ts | 6 |
| @better-auth/passkey | 1.7.7 | lib/auth.ts, lib/auth-client.ts (passkeys; VF-001 gap in vukem) | 2 |
| botid | 1.5.11 | lib/bot.ts, instrumentation-client.ts, next.config.ts (bot protection; VF-003) | 3 |
| qrcode | 1.5.4 | app/[locale]/profile/two-factor-section.tsx (TOTP QR) | 1 |

## Payments
| package | pinned | used by (examples) | files |
| --- | --- | --- | --- |
| stripe | 23.0.0 | actions/billing.ts, app/api/webhooks/stripe/route.ts | 17 |

## Email
| package | pinned | used by (examples) | files |
| --- | --- | --- | --- |
| resend | 6.32.0 | lib/email.ts (send); components/auth/email-request-form.tsx | 2 |
| @react-email/components | 1.0.12 | emails/action-email.tsx, emails/contact-email.tsx (templates) | 2 |

## Search
| package | pinned | used by (examples) | files |
| --- | --- | --- | --- |
| postgres (driver) + drizzle-orm | 3.4.9 / 0.45.3 | lib/catalog/repository.ts L300-323: pg_trgm word_similarity + websearch_to_tsquery('simple'); extensions pg_trgm, unaccent in lib/db/migrations/0000_extensions.sql | 3 / 45 |
No external search engine: search is Postgres only.

## i18n
| package | pinned | used by (examples) | files |
| --- | --- | --- | --- |
| next-intl | 4.14.9 | actions/billing.ts, app/[locale]/contact/contact-form.tsx | 86 |

## Analytics
| package | pinned | used by (examples) | files |
| --- | --- | --- | --- |
| @vercel/analytics | 2.0.1 | app/[locale]/layout.tsx (`@vercel/analytics/next`) | 1 |
| @vercel/speed-insights | 2.0.0 | app/[locale]/layout.tsx (`@vercel/speed-insights/next`) | 1 |

## UI
| package | pinned | used by (examples) | files |
| --- | --- | --- | --- |
| @base-ui/react | 1.8.0 | profile/delete-account.tsx, components/billing/checkout-form.tsx (Dialog, Radio) | 7 |
| @scalar/api-reference-react | 0.9.77 | components/api/api-reference.tsx (API docs page) | 1 |
| @tanstack/react-query | 5.104.1 | components/layout/query-provider.tsx, components/item/ratings-dialog.tsx, checkout-form.tsx | 4 |

## Other (outside the requested groups)
| package | pinned | used by | files |
| --- | --- | --- | --- |
| @upstash/ratelimit / @upstash/redis | 2.2.0 / 1.39.0 | lib/api/quota.ts (API quota) | 1 each |
| flags | 4.3.1 | lib/flags.ts, app/.well-known/vercel/flags/route.ts | 3 |
| @electric-sql/pglite | 0.5.8 | lib/db/client.ts, lib/billing/test-db.ts (tests) | 6 |
| zod | 4.6.5 | actions/*.ts and others | 49 |
| next / react / react-dom | 16.3.8 / 19.3.0 / 19.3.0 | framework | - |

## SDKs (legacy/sdks)
| package | version | notes |
| --- | --- | --- |
| sdks/ts `@hey-api/openapi-ts` | 0.99.0 (dev) | generator; `typescript` 6.0.3 (dev) |
| sdks/python httpx | `>=0.23.1,<0.29.0` | range, not pinned; attrs `>=22.2.0`; generator openapi-python-client 0.29.1 (pyproject.toml comment) |
