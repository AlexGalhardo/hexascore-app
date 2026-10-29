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

Test names are Go tests under `tests/` (or `tests/scoring/`) unless marked otherwise. The e2e specs of P1.I.2 do not exist yet, so e2e rows are covered only by the Go tests named in their replacement column.

## Routes

| # | Legacy route | Replacement | Test | Status |
| --- | --- | --- | --- | --- |
| R01 | `/` | `GET /` (`controllers.Home`) | TestHome | done |
| R02 | `/[...rest]` | `GET /{path...}` (`controllers.NotFound`) | TestErrorPage | done |
| R03 | `/[section]` | `GET /{section}` (catalog listings) | TestInertiaPages_Catalog_Index | done |
| R04 | `/[section]/[slug]` | `GET /{section}/{slug}` (item pages) | TestInertiaPages_Catalog_Show | done |
| R05 | `/admin` | admin panel root (admin registry) | TestAdminPanelAccess, e2e `admin.spec.ts` | done (guests are redirected to `/login`, legacy answered 404; accepted by the orchestrator 2026-10-29) |
| R06 | `/admin/contact` | `/admin/contact-messages` (`ContactMessagesResource`) | none | open (P1.I.2.3) |
| R07 | `/admin/ingestion` | `GET /admin/ingestion` | none | open (P1.I.2.3) |
| R08 | `/admin/items` | admin items resource | none | open (P1.I.2.3) |
| R09 | `/admin/items/[id]` | `GET /admin/items/{id}/manage` | TestAffiliateLinks | done |
| R10 | `/admin/reports` | `/admin/reports` resource; resolve is `POST /admin/reports/{id}/resolve` | TestReportResolveNotifies covers resolve only | open (P1.I.2.3) |
| R11 | `/contact` | `GET /contact`, `POST /contact` | TestContactHoneypotHTTP | done |
| R12 | `/docs` | `GET /docs` | TestDocsPages | done |
| R13 | `/docs/authentication` | `GET /docs/{slug}` (`resources/docs/authentication.md`) | TestDocsPages | done |
| R14 | `/docs/changelog` | `GET /docs/{slug}` (`resources/docs/changelog.md`) | TestDocsPages | done |
| R15 | `/docs/errors` | `GET /docs/{slug}` (`resources/docs/errors.md`) | TestDocsPages | done |
| R16 | `/docs/pagination` | `GET /docs/{slug}` (`resources/docs/pagination.md`) | TestDocsPages | done |
| R17 | `/docs/rate-limits` | `GET /docs/{slug}` (`resources/docs/rate-limits.md`) | TestDocsPages | done |
| R18 | `/docs/reference` | none found in `routes/`; Scalar hosting not decided | none | TBD (senior) |
| R19 | `/feed` | none | none | deferred D-06 |
| R20 | `/login` | `GET /login`, `POST /login` | TestSecurity_A07_LoginThrottle | done |
| R21 | `/onboarding` | `GET /onboarding`, `POST /onboarding` | TestOnboardingNickname | done |
| R22 | `/policy` | `GET /policy` | TestLegalPages | done |
| R23 | `/pricing` | `GET /billing` (`billing.show`) | none | TBD (senior) |
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
| A07 | `admin.ts` `setContactHandled` | `/admin/contact-messages` edit (`handled`) | none | open (P1.I.2.3) |
| A08 | `admin.ts` `setReportStatus` | `POST /admin/reports/{id}/resolve` | TestReportResolveNotifies | done |
| A09 | `admin.ts` `setListHidden` | no direct route; resolve outcome `list_hidden` exists | none | TBD (senior) |
| A10 | `admin.ts` `banUser` | `POST /admin/users/{id}/ban` | TestSecurity_A07_Ban | done |
| A11 | `admin.ts` `saveAffiliateLink` | `POST /admin/items/{id}/affiliate-links` | TestAffiliateLinks | done |
| A12 | `admin.ts` `deleteAffiliateLink` | `POST /admin/items/{id}/affiliate-links/{store}/delete` | TestAffiliateLinks | done |
| A13 | `admin.ts` `mergeItems` | `POST /admin/items/{id}/merge` | TestAdminMerge | done |
| A14 | `api-keys.ts` `createApiKeyAction` | `POST /profile/api/keys` | TestKeysWebFlow | done |
| A15 | `api-keys.ts` `revokeApiKeyAction` | `DELETE /profile/api/keys/{id}` | TestKeysWebFlow | done |
| A16 | `billing.ts` `startStripeCheckout` | `POST /billing/checkout/{plan}` | none | TBD (senior) |
| A17 | `billing.ts` `openBillingPortal` | `POST /billing/portal` | none | TBD (senior) |
| A18 | `billing.ts` `createPixCharge` | none | none | deferred D-06 |
| A19 | `billing.ts` `simulatePixPaymentAction` | none | none | deferred D-06 |
| A20 | `contact.ts` `sendContactMessage` | `POST /contact` | TestContactHoneypotHTTP | done |
| A21 | `follows.ts` `toggleFollow` | none | none | deferred D-06 |
| A22 | `item-status.ts` `saveItemStatus` | `PUT /items/{id}/status` | TestStatus | done |
| A23 | `lists.ts` `createList` | `POST /lists` | TestListWriteIdempotencyKey | done |
| A24 | `lists.ts` `updateList` | `PATCH /lists/{id}` | TestListTypeIsImmutableAndEventsFire | done |
| A25 | `lists.ts` `deleteList` | `DELETE /lists/{id}` | none | open (P1.I.2.2) |
| A26 | `lists.ts` `addToList` | `PUT /lists/{id}/items/{itemId}` | TestListItemsAddRemoveReorder | done |
| A27 | `lists.ts` `removeFromList` | `DELETE /lists/{id}/items/{itemId}` | TestListItemsAddRemoveReorder | done |
| A28 | `lists.ts` `moveInList` | `PUT /lists/{id}/order` | TestListItemsAddRemoveReorder | done |
| A29 | `ratings.ts` `saveRating` | `PUT /items/{id}/rating` | TestRatingWriteValidationEditDelete | done |
| A30 | `ratings.ts` `deleteRating` | `DELETE /items/{id}/rating` | TestRatingWriteValidationEditDelete | done |
| A31 | `reports.ts` `report` | `POST /reports` | TestSecurity_A01_ReportSelf | done |
| A32 | `time-to-beat.ts` `saveTimeToBeat` | `PUT /items/{id}/time-to-beat` | TestTimeToBeat | done |
| A33 | `time-to-beat.ts` `deleteTimeToBeat` | `DELETE /items/{id}/time-to-beat` | none | open (P1.I.2.2) |

## e2e tests

| # | Legacy spec and test | Replacement | Test | Status |
| --- | --- | --- | --- | --- |
| E01 | admin: `/admin answers 404 to visitors and regular members` | admin panel access rules | TestSecurity_A01_AdminPagesNotFoundForMembers | done |
| E02 | admin: `a member reports a user; an admin resolves it in /admin/reports` | `POST /reports`, `POST /admin/reports/{id}/resolve` | TestReportResolveNotifies | done |
| E03 | api-keys: `create a key once, use it, revoke it` | `/profile/api` keys flow | TestKeysWebFlow | done |
| E04 | api-keys: `rejects empty and over-long key names` | key name validation | TestKeysNameAndCap | done |
| E05 | auth: `signup → verification link → logged in` | `/signup`, `/verify-email` | TestRegisterAndVerifyEmail | done |
| E06 | auth: `signup validates nickname and password rules on the client` | client-side validation (UI) | none | open (P1.I.2.1) |
| E07 | auth: `wrong password shows a generic error` | `POST /login` | none | open (P1.I.2.1) |
| E08 | auth: `login with TOTP two-factor` | `/two-factor-challenge` | TestTwoFactorLogin | done |
| E09 | catalog: `${section}: filters change the URL and the list` | `/{section}` filters | none | open (P1.I.2.2) |
| E10 | catalog: `keyset pagination goes forward and back` | catalog cursor pagination | none | open (P1.I.2.2) |
| E11 | catalog: `filters with no match show the empty state` | `/{section}` empty state | none | open (P1.I.2.2) |
| E12 | catalog: `a tampered cursor or filter never breaks the page` | catalog cursor validation | none | open (P1.I.2.2) |
| E13 | contact: `counter follows the message and the form validates` | `/contact` form (UI) | none | open (P1.I.2.3) |
| E14 | contact: `message length is capped by the field` | `/contact` form (UI) | none | open (P1.I.2.3) |
| E15 | contact: `sends a message with identity from the session` | `POST /contact` | none | open (P1.I.2.3) |
| E16 | docs: `navigation reaches every docs page` | `/docs/{slug}` | TestDocsPages | done |
| E17 | docs: `Scalar reference renders under the strict CSP` | depends on R18 | none | TBD (senior) |
| E18 | docs: `localized reference loads the translated spec` | depends on R18 | none | TBD (senior) |
| E19 | docs: `${path} guide is translated` | `/docs/{slug}` in `en`, `pt-BR`, `es` | none | open (P1.I.2.3) |
| E20 | docs: `${path} is served as JSON` | `/docs/{slug}` JSON | none | open (P1.I.2.3) |
| E21 | headers: `pages get a per-request nonce CSP that every script carries` | CSP middleware | none | open (P1.I.2.3) |
| E22 | headers: `API responses are JSON-only and cannot be framed` | `/api/v1`, `/app/v1` headers | none | open (P1.I.2.3) |
| E23 | headers: `Scalar allowances are scoped to the reference pages` | depends on R18 | none | TBD (senior) |
| E24 | item: `shows score, categories and a centered, paginated users' scores modal` | `/{section}/{slug}` (UI) | none | open (P1.I.2.2) |
| E25 | item: `trailers load only after consent, from youtube-nocookie` | item page (UI) | none | open (P1.I.2.2) |
| E26 | item: `trailers load immediately` | item page (UI) | none | open (P1.I.2.2) |
| E27 | item: `renders with license attribution in ${code}` | item page (UI) | none | open (P1.I.2.2) |
| E28 | not-found: `${prefix}${path} answers 404 with the localized page` | `GET /{path...}` with locales | none | open (P1.I.2.3) |
| E29 | not-found: `API 404s are RFC 9457 problems` | `/api/v1` problem responses | none | open (P1.I.2.3) |
| E30 | pricing: `lists the three plans and switches currency on the server` | `GET /billing` | none | TBD (senior) |
| E31 | pricing: `a forged currency falls back to a valid one` | `GET /billing` | none | TBD (senior) |
| E32 | pricing: `checkout requires login` | `POST /billing/checkout/{plan}` | none | TBD (senior) |
| E33 | pricing: `shows the current plan, the provider form or the not-configured notice, and return states` | `GET /billing` | none | TBD (senior) |
| E34 | profile: `profile requires login` | `GET /profile` | none | open (P1.I.2.1) |
| E35 | profile: `change password, then log in with the new one` | `PUT /profile/password` | TestProfilePassword | done |
| E36 | profile: `account deletion: 10 s countdown, cancel, confirm, banner and undo` | server flow in `POST`/`DELETE /profile/deletion`; countdown is UI | TestDeletionScheduleAndCancel (server only) | open (P1.I.2.1) |
| E37 | public-pages: `${prefix}${path} renders, is themed and accessible` | public pages (UI, axe) | none | open (P1.I.2.3) |
| E38 | public-pages: `theme toggle persists across reloads` | theme (UI) | none | open (P1.I.2.3) |
| E39 | public-pages: `asks once and remembers the choice` | consent banner (UI) | none | open (P1.I.2.3) |
| E40 | public-pages: `locale prefixes render translated navigation` | locale prefixes | TestLocalePrefix_Routing (routing only) | open (P1.I.2.3) |
| E41 | ratings: `rate in 6 categories, edit, community score appears at the 5th rating` | `PUT /items/{id}/rating` | TestRatingWriteValidationEditDelete (partial) | open (P1.I.2.2) |
| E42 | ratings: `item status cycles through the four values and shows on the profile shelves` | `PUT /items/{id}/status` | TestStatus (partial) | open (P1.I.2.2) |
| E43 | ratings: `curated list: create from the item page, then see it on the profile and the list page` | lists flow | none | open (P1.I.2.2) |
| E44 | search: `autocomplete is keyboard operable and opens the highlighted item` | `GET /search/suggest` (UI) | none | open (P1.I.2.2) |
| E45 | search: `Enter without a highlighted suggestion runs a full search` | `GET /search` | none | open (P1.I.2.2) |
| E46 | search: `short, empty and hostile queries are handled` | `GET /search` | none | open (P1.I.2.2) |
| E47 | search: `typo-tolerant search finds items (trigram)` | `GET /search` | none | open (P1.I.2.2) |
| E48 | social: `public profile lists ratings; a private one hides them from everyone but the owner` | `GET /u/{nickname}` | TestPublicProfilePrivacy | done |
| E49 | social: `follow a member and see their ratings in /feed` | none (`/feed`, `toggleFollow`) | none | deferred D-06 |
| E50 | social: `/feed requires login` | none (`/feed`) | none | deferred D-06 |
| E51 | time-to-beat: `report hours on a completed game; the median shows at the 3rd report` | `PUT /items/{id}/time-to-beat` | none | open (P1.I.2.2) |

## Notes

- Owner question Q-H31 blocks billing rows (R23, A16, A17, E30-E33). Their Go tests exist (TestBillingCheckout, TestBillingPortal, TestBillingCheckoutCurrency) but are not counted as replacement evidence until the owner answers.
- D-09 (critic scores dropped): no legacy route, action or test in this inventory is a critic-score row.
- Rows marked `open` or `TBD (senior)` block the release (P1.I.7.1) per the microtask notes.
