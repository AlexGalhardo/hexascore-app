# E-013 legacy i18n inventory (table)

Source: legacy/apps/web/messages/{en,pt-BR,es}/*.json (read-only). Leaf keys = jq paths(scalars). Rules: D-06 (docs/brainstorming.md line 47) and D-14/D-12 where cited.

| namespace | file(s) | leaf keys en/pt-BR/es | missing in pt-BR | missing in es | keep/drop | reason (D-xx cite) |
| --- | --- | --- | --- | --- | --- | --- |
| account | account.json (en, pt-BR, es) | 70/70/70 | 0 | 0 | keep | keep; drop account.passkeys.* (9 keys) per D-06 passkeys |
| api | api.json (en, pt-BR, es) | 69/69/69 | 0 | 0 | keep | keep; drop api.profile.plan.credits (1) per D-06 credit packs |
| auth | auth.json (en, pt-BR, es) | 88/88/88 | 0 | 0 | keep | keep; drop auth.passkey, auth.errors.passkeyFailed (2) per D-06 passkeys |
| billing | billing.json (en, pt-BR, es) | 89/89/89 | 0 | 0 | keep | keep Stripe checkout/plans/pricing; drop 18 PIX/AbacatePay + credit-pack keys per D-06 and D-14 (PIX second release) |
| catalog | catalog.json (en, pt-BR, es) | 42/42/42 | 0 | 0 | keep | keep (core catalogue, no dropped feature) |
| categories | categories.json (en, pt-BR, es) | 32/32/32 | 0 | 0 | keep | keep (core catalogue) |
| common | common.json (en, pt-BR, es) | 30/30/30 | 0 | 0 | keep | keep; drop nav.feed (1) per D-06 feed |
| contact | contact.json (en, pt-BR, es) | 14/14/14 | 0 | 0 | keep | keep (contact.subjects.feedback is a contact subject, not the feed) |
| docs | docs.json (en, pt-BR, es) | 120/120/120 | 0 | 0 | keep | keep API docs; drop limits.creditPacks, limits.colCredits (D-06 credit packs) and 3 changelog keys about per-locale OpenAPI (D-06) |
| errors | errors.json (en, pt-BR, es) | 7/7/7 | 0 | 0 | keep | keep (shared errors) |
| item | item.json (en, pt-BR, es) | 80/80/80 | 0 | 0 | keep | keep (item.coverCredit is image credit, not a credit pack) |
| legal | legal.json (en, pt-BR, es) | 70/70/70 | 0 | 0 | keep | keep (legal pages) |
| lists | lists.json (en, pt-BR, es) | 33/33/33 | 0 | 0 | keep | keep (lists are in scope, mobile v1 per D-12) |
| openapi | openapi.json (en, pt-BR, es) | 111/111/111 | 0 | 0 | drop (keep? en-only) | drop whole namespace: per-locale OpenAPI documents deferred per D-06; if the en-only OpenAPI doc is still needed for vukem openapi:export, owner to confirm (keep?) |
| ratings | ratings.json (en, pt-BR, es) | 13/13/13 | 0 | 0 | keep | keep (star rating in scope) |
| search | search.json (en, pt-BR, es) | 11/11/11 | 0 | 0 | keep | keep (search in scope, D-16) |
| social | social.json (en, pt-BR, es) | 72/72/72 | 0 | 0 | keep | keep; drop feed.* (7) and profile follow/unfollow/followers/following (4) per D-06 feed and follows |

Totals: kept 16 namespaces, 793 leaf keys (per locale, after dropping 47 sub-keys inside kept namespaces); dropped whole namespace: openapi (111 keys).
