# Parity checklist: legacy app to vukem rewrite

Cutover gate (D-21). Every legacy route, Server Action and e2e test maps to a replacement, a test and a status. Rows are generated from `legacy/` by `scripts/parity-inventory.sh` (one row per script line: 34 routes, 33 Server Actions, 51 e2e tests = 118 rows).

Status values:

- `done`: the replacement exists and a named test covers it.
- `deferred D-06` / `deferred D-09`: out of Phase 1 by owner decision (`docs/brainstorming.md`).
- `open (P1.I.x.y)`: replacement or test still to be written by the named microtask.
- `TBD (senior)`: needs an architecture or owner decision (Q-H31 for billing).

The gate is green only when no row is `open` or `TBD`. Verify the count:

```sh
scripts/parity-inventory.sh | wc -l
rg -c '\| (done|deferred D-0[69]) \|' docs/parity-checklist.md
```

Test names are Go tests under `tests/` (or `tests/scoring/`) unless marked otherwise. Playwright specs in `qa/e2e/specs/` cover the ported flows. A `done` row names a Go test or that spec's title.

## Routes

| # | Legacy route | Replacement | Test | Status |
| --- | --- | --- | --- | --- |
| R01 | `/` | `GET /` (`controllers.Home`) | TestHome | done |
| R02 | `/[...rest]` | `GET /{path...}` (`controllers.NotFound`) | TestErrorPage | done |
| R03 | `/[section]` | `GET /{section}` (catalog listings) | TestInertiaPages_Catalog_Index | done |
| R04 | `/[section]/[slug]` | `GET /{section}/{slug}` (item pages) | TestInertiaPages_Catalog_Show | done |
| R05 | `/admin` | admin panel root (admin registry) | TestAdminPanelAccess | done |
| R06 | `/admin/contact` | `/admin/contact-messages` (`ContactMessagesResource`) | TestAdminMarksContactHandled | done |
| R07 | `/admin/ingestion` | `GET /admin/ingestion` | TestSecurity_A01_AdminPagesNotFoundForMembers | done |
| R08 | `/admin/items` | admin items resource | TestSecurity_A01_AdminPagesNotFoundForMembers | done |
| R09 | `/admin/items/[id]` | `GET /admin/items/{id}/manage` | TestAffiliateLinks | done |
| R10 | `/admin/reports` | `/admin/reports` resource; resolve is `POST /admin/reports/{id}/resolve` | TestReportResolveNotifies | done |
| R11 | `/contact` | `GET /contact`, `POST /contact` | TestContactHoneypotHTTP | done |
| R12 | `/docs` | `GET /docs` | TestDocsPages | done |
| R13 | `/docs/authentication` | `GET /docs/{slug}` (`resources/docs/authentication.md`) | TestDocsPages | done |
| R14 | `/docs/changelog` | `GET /docs/{slug}` (`resources/docs/changelog.md`) | TestDocsPages | done |
| R15 | `/docs/errors` | `GET /docs/{slug}` (`resources/docs/errors.md`) | TestDocsPages | done |
| R16 | `/docs/pagination` | `GET /docs/{slug}` (`resources/docs/pagination.md`) | TestDocsPages | done |
| R17 | `/docs/rate-limits` | `GET /docs/{slug}` (`resources/docs/rate-limits.md`) | TestDocsPages | done |
| R18 | `/docs/reference` | `GET /docs/reference` (Scalar) | TestAPIReference_PublicInProduction | done |
| R19 | `/feed` | none | none | deferred D-06 |
| R20 | `/login` | `GET /login`, `POST /login` | TestSecurity_A07_LoginThrottle | done |
| R21 | `/onboarding` | `GET /onboarding`, `POST /onboarding` | TestOnboardingNickname | done |
| R22 | `/policy` | `GET /policy` | TestLegalPages | done |
| R23 | `/pricing` | `GET /pricing` | TestPricingPage | done |
| R24 | `/profile` | `GET /profile` | TestProfileDetails | done |
| R25 | `/profile/api` | `GET /profile/api` | TestKeysWebFlow | done |
| R26 | `/profile/api/checkout` | redirect to `/billing` (`routes/billing.go`) | TestLegacyCheckoutRedirect | done |
| R27 | `/recover-password` | `GET /recover-password`, `POST /recover-password` | TestPasswordReset | done |
| R28 | `/reset-password` | `GET /reset-password/{token}`, `POST /reset-password` | TestPasswordResetAnswersAlike | done |
| R29 | `/search` | `GET /search` | TestSearchPage | done |
| R30 | `/signup` | `GET /signup`, `POST /signup` | TestRegisterAndVerifyEmail | done |
| R31 | `/terms` | `GET /terms` | TestLegalPages | done |
| R32 | `/u/[nickname]` | `GET /u/{nickname}` | TestPublicProfilePrivacy | done |
| R33 | `/u/[nickname]/lists/[id]` | `GET /u/{nickname}/lists/{id}` | TestInertiaPages_UserLibraryLists | done |
| R34 | `/verify-email` | `GET /verify-email`, confirm and resend routes | TestUnverifiedUsersSeeTheNotice | done |

## Server Actions

| # | Legacy action | Replacement | Test | Status |
| --- | --- | --- | --- | --- |
| A01 | `account.ts` `updateName` | `PATCH /profile/details` | TestProfileDetails | done |
| A02 | `account.ts` `setNickname` | `POST /onboarding` (nickname) | TestOnboardingNickname | done |
| A03 | `account.ts` `setPassword` | `PUT /profile/password` | TestProfilePassword_SetFirstTime | done |
| A04 | `account.ts` `setProfilePrivate` | `PATCH /profile/privacy` | TestProfilePrivacy | done |
| A05 | `account.ts` `scheduleDeletion` | `POST /profile/deletion` | TestDeletionScheduleAndCancel | done |
| A06 | `account.ts` `cancelDeletion` | `DELETE /profile/deletion` | TestDeletionScheduleAndCancel | done |
| A07 | `admin.ts` `setContactHandled` | `/admin/contact-messages` edit (`handled`) | TestAdminMarksContactHandled | done |
| A08 | `admin.ts` `setReportStatus` | `POST /admin/reports/{id}/resolve` | TestReportResolveNotifies | done |
| A09 | `admin.ts` `setListHidden` | `POST /admin/reports/{id}/resolve` outcome `list_hidden` | TestReportResolveNotifies | done |
| A10 | `admin.ts` `banUser` | `POST /admin/users/{id}/ban` | TestSecurity_A07_Ban | done |
| A11 | `admin.ts` `saveAffiliateLink` | `POST /admin/items/{id}/affiliate-links` | TestAffiliateLinks | done |
| A12 | `admin.ts` `deleteAffiliateLink` | `POST /admin/items/{id}/affiliate-links/{store}/delete` | TestAffiliateLinks | done |
| A13 | `admin.ts` `mergeItems` | `POST /admin/items/{id}/merge` | TestAdminMerge | done |
| A14 | `api-keys.ts` `createApiKeyAction` | `POST /profile/api/keys` | TestKeysWebFlow | done |
| A15 | `api-keys.ts` `revokeApiKeyAction` | `DELETE /profile/api/keys/{id}` | TestKeysWebFlow | done |
| A16 | `billing.ts` `startStripeCheckout` | `POST /billing/checkout/{plan}` | TestBillingCheckout | done |
| A17 | `billing.ts` `openBillingPortal` | `POST /billing/portal` | TestBillingPortal | done |
| A18 | `billing.ts` `createPixCharge` | none | none | deferred D-06 |
| A19 | `billing.ts` `simulatePixPaymentAction` | none | none | deferred D-06 |
| A20 | `contact.ts` `sendContactMessage` | `POST /contact` | TestContactHoneypotHTTP | done |
| A21 | `follows.ts` `toggleFollow` | none | none | deferred D-06 |
| A22 | `item-status.ts` `saveItemStatus` | `PUT /items/{id}/status` | TestStatus | done |
| A23 | `lists.ts` `createList` | `POST /lists` | TestListWriteIdempotencyKey | done |
| A24 | `lists.ts` `updateList` | `PATCH /lists/{id}` | TestListTypeIsImmutableAndEventsFire | done |
| A25 | `lists.ts` `deleteList` | `DELETE /lists/{id}` | e2e lists `create, fill, reorder, publish, hide and delete a list` | done |
| A26 | `lists.ts` `addToList` | `PUT /lists/{id}/items/{itemId}` | TestListItemsAddRemoveReorder | done |
| A27 | `lists.ts` `removeFromList` | `DELETE /lists/{id}/items/{itemId}` | TestListItemsAddRemoveReorder | done |
| A28 | `lists.ts` `moveInList` | `PUT /lists/{id}/order` | TestListItemsAddRemoveReorder | done |
| A29 | `ratings.ts` `saveRating` | `PUT /items/{id}/rating` | TestRatingWriteValidationEditDelete | done |
| A30 | `ratings.ts` `deleteRating` | `DELETE /items/{id}/rating` | TestRatingWriteValidationEditDelete | done |
| A31 | `reports.ts` `report` | `POST /reports` | TestSecurity_A01_ReportSelf | done |
| A32 | `time-to-beat.ts` `saveTimeToBeat` | `PUT /items/{id}/time-to-beat` | TestTimeToBeat | done |
| A33 | `time-to-beat.ts` `deleteTimeToBeat` | `DELETE /items/{id}/time-to-beat` | TestTimeToBeat | done |

## e2e tests

| # | Legacy spec and test | Replacement | Test | Status |
| --- | --- | --- | --- | --- |
| E01 | admin: `/admin answers 404 to visitors and regular members` | admin panel access rules | TestSecurity_A01_AdminPagesNotFoundForMembers | done |
| E02 | admin: `a member reports a user; an admin resolves it in /admin/reports` | `POST /reports`, `POST /admin/reports/{id}/resolve` | TestReportResolveNotifies | done |
| E03 | api-keys: `create a key once, use it, revoke it` | `/profile/api` keys flow | TestKeysWebFlow | done |
| E04 | api-keys: `rejects empty and over-long key names` | key name validation | TestKeysNameAndCap | done |
| E05 | auth: `signup → verification link → logged in` | `/signup`, `/verify-email` | TestRegisterAndVerifyEmail | done |
| E06 | auth: `signup validates nickname and password rules on the client` | `/signup` | e2e auth `signup refuses a weak password and stays on the form` | done |
| E07 | auth: `wrong password shows a generic error` | `POST /login` | e2e auth `wrong password shows a generic error` | done |
| E08 | auth: `login with TOTP two-factor` | `/two-factor-challenge` | TestTwoFactorLogin | done |
| E09 | catalog: `${section}: filters change the URL and the list` | `/{section}` filters | e2e catalog `filters change the URL and the list` | done |
| E10 | catalog: `keyset pagination goes forward and back` | catalog cursor | TestCursorRoundTrip | done |
| E11 | catalog: `filters with no match show the empty state` | `/{section}` empty state | e2e catalog `filters with no match show the empty state` | done |
| E12 | catalog: `a tampered cursor or filter never breaks the page` | catalog cursor validation | e2e catalog `a tampered cursor or filter never breaks the page` | done |
| E13 | contact: `counter follows the message and the form validates` | `/contact` form | e2e contact `the form is accessible and validates before sending` | done |
| E14 | contact: `message length is capped by the field` | `/contact` form | e2e contact `message length is capped by the field` | done |
| E15 | contact: `sends a message with identity from the session` | `POST /contact` | e2e contact `a valid message is sent and hostile markup stays inert` | done |
| E16 | docs: `navigation reaches every docs page` | `/docs/{slug}` | TestDocsPages | done |
| E17 | docs: `Scalar reference renders under the strict CSP` | `GET /docs/reference` | e2e docs `the Scalar reference is public and renders the OpenAPI document` | done |
| E18 | docs: `localized reference loads the translated spec` | `GET /docs/reference` (one OpenAPI document) | e2e docs `the Scalar reference is public and renders the OpenAPI document` | done |
| E19 | docs: `${path} guide is translated` | `/docs/{slug}` in `en`, `pt-BR`, `es` | e2e docs `the docs index and every guide page render` | done |
| E20 | docs: `${path} is served as JSON` | `GET /api/v1/openapi.json` | e2e docs `the Scalar reference is public and renders the OpenAPI document` | done |
| E21 | headers: `pages get a per-request nonce CSP that every script carries` | CSP middleware | TestSecurity_A02_HexaHeadersWeb | done |
| E22 | headers: `API responses are JSON-only and cannot be framed` | `/api/v1` headers | TestSecurity_A02_HexaHeadersAPI | done |
| E23 | headers: `Scalar allowances are scoped to the reference pages` | same CSP as the rest of the site | TestSecurity_A02_HexaCSPHasStrictDynamic | done |
| E24 | item: `shows score, categories and a centered, paginated users' scores modal` | `GET /api/items/{id}/ratings` and the scores dialog | e2e item `shows score, categories and a centered, paginated users' scores modal` | done |
| E25 | item: `trailers load only after consent, from youtube-nocookie` | click-to-load trailers | e2e item `trailers load only after consent, from youtube-nocookie` | done |
| E26 | item: `trailers load immediately` | trailers when `consent=all` | e2e item `trailers load immediately` | done |
| E27 | item: `renders with license attribution in ${code}` | item page attribution | e2e item `shows the community score with category averages and attribution` | done |
| E28 | not-found: `${prefix}${path} answers 404 with the localized page` | `GET /{path...}` with locales | e2e not-found `answers 404 with the localized page` | done |
| E29 | not-found: `API 404s are RFC 9457 problems` | `/api/v1` problem responses | e2e not-found `API 404s are RFC 9457 problems` | done |
| E30 | pricing: `lists the three plans and switches currency on the server` | `GET /pricing` | e2e pricing `lists the three plans, the coverage statement and switches currency on the server` | done |
| E31 | pricing: `a forged currency falls back to a valid one` | `GET /pricing` | e2e pricing `a forged currency falls back to a valid one` | done |
| E32 | pricing: `checkout requires login` | `POST /billing/checkout/{plan}` | TestBillingCheckout | done |
| E33 | pricing: `shows the current plan, the provider form or the not-configured notice, and return states` | `GET /billing` | TestBillingPage | done |
| E34 | profile: `profile requires login` | `GET /profile` | e2e profile `profile requires login` | done |
| E35 | profile: `change password, then log in with the new one` | `PUT /profile/password` | TestProfilePassword | done |
| E36 | profile: `account deletion: 10 s countdown, cancel, confirm, banner and undo` | `POST`/`DELETE /profile/deletion` | e2e profile `account deletion: 10 s countdown, cancel, confirm, banner and undo` | done |
| E37 | public-pages: `${prefix}${path} renders, is themed and accessible` | public pages | e2e public-pages `renders in the project locale and is accessible` | done |
| E38 | public-pages: `theme toggle persists across reloads` | theme cookie | e2e public-pages `theme toggle persists across reloads` | done |
| E39 | public-pages: `asks once and remembers the choice` | cookie banner | e2e public-pages `asks once and remembers the choice` | done |
| E40 | public-pages: `locale prefixes render translated navigation` | locale prefixes | e2e public-pages `the locale prefixes render translated navigation` | done |
| E41 | ratings: `rate in 6 categories, edit, community score appears at the 5th rating` | `PUT /items/{id}/rating` | TestRatingWriteValidationEditDelete | done |
| E42 | ratings: `item status cycles through the four values and shows on the profile shelves` | `PUT /items/{id}/status` | TestStatus | done |
| E43 | ratings: `curated list: create from the item page, then see it on the profile and the list page` | lists flow | e2e lists `create, fill, reorder, publish, hide and delete a list` | done |
| E44 | search: `autocomplete is keyboard operable and opens the highlighted item` | `GET /search/suggest` | e2e search `autocomplete is keyboard operable and opens the highlighted item` | done |
| E45 | search: `Enter without a highlighted suggestion runs a full search` | `GET /search` | e2e search `Enter without a highlighted suggestion runs a full search` | done |
| E46 | search: `short, empty and hostile queries are handled` | `GET /search` | e2e search `short, empty and hostile queries are handled` | done |
| E47 | search: `typo-tolerant search finds items (trigram)` | `GET /search` | e2e search `typo-tolerant suggestions answer JSON` | done |
| E48 | social: `public profile lists ratings; a private one hides them from everyone but the owner` | `GET /u/{nickname}` | TestPublicProfilePrivacy | done |
| E49 | social: `follow a member and see their ratings in /feed` | none (`/feed`, `toggleFollow`) | none | deferred D-06 |
| E50 | social: `/feed requires login` | none (`/feed`) | none | deferred D-06 |
| E51 | time-to-beat: `report hours on a completed game; the median shows at the 3rd report` | `PUT /items/{id}/time-to-beat` | e2e time-to-beat `report hours on a completed game, then lose the form when the status changes` | done |

## Notes

- Guests hitting `/admin` are redirected to `/login`. Legacy answered 404. The orchestrator accepted that on 2026-10-29. Members still get 404 (`TestSecurity_A01_AdminPagesNotFoundForMembers`).
- Q-H31 still blocks live Stripe keys on Railway. The local billing rows use `TestBillingCheckout`, `TestBillingPortal` and `TestPricingPage`.
- D-09 dropped critic scores. No inventory row is only a critic-score check.
- E24, E25, E26 and E39 are the users' scores dialog, the trailer player and the cookie banner. `TestPublicItemRatings` covers the JSON page; the Playwright titles cover the UI.
