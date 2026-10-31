# QA-i18n-parity Key parity and untranslated values (resources/lang vs en)
status: concluida (report only; no edits)
branch: none (main @ 246ced8 at read time)
files: .orchestration/reports/QA-i18n-parity.md
acceptance:
- command: `bun -e` flatten of `resources/lang/{en,pt-BR,es}.json` (string leaves), compare key sets and values
  - pt-BR: en leaves 1013, pt-BR leaves 1013, missing 0, extra 0, value equal to en: 39
  - es: en leaves 1013, es leaves 1013, missing 0, extra 0, value equal to en: 28
  - keys equal to en in both locales: 26
validation: key parity ok in both locales; no code touched
changelog: none
roadmap: no box
risks / follow-ups: classification below is mine, from the values only, not from a native review. Brand, abbreviation and cognate keys are likely fine.

## Equal-to-en values, classified

Likely real misses (translate in the text-editor pass):

| key | en | pt-BR | es |
| --- | --- | --- | --- |
| billing.plans.free | Free | Free | Free |
| billing.plans.pro | Pro | Pro | Pro |
| billing.plans.business | Business | Business | Business |
| categories.game.gameplay | Gameplay | Gameplay | (not equal) |
| categories.game.storytelling | Storytelling | Storytelling | (not equal) |
| auth.fields.nickname | Nickname | Nickname | Nickname |
| account.details.nickname | Nickname | Nickname | Nickname |
| nav.menu | Menu | Menu | (not equal) |
| social.report.reasons.spam | Spam | Spam | Spam |

Probably acceptable (brand, abbreviation, or word used as-is in the locale):

- Brands and sources: `item.stores.amazon/steam/gog`, `item.sources.*` (Metacritic, Rotten Tomatoes, IMDb, Steam, Goodreads, HowLongToBeat, Open Library), `item.stores.mercadolivre` (pt only).
- Abbreviations and formats: `item.runtimeValue` ({minutes} min), `item.timeToBeat.hours` ({hours} h), `item.editionFields.isbn` (ISBN-13), `item.trailers` (pt "Trailers").
- Words used as-is: `api.profile.title`, `api.profile.keys.status` / `social.status.label` / `library.status` / `social.profile.tabs.status` (Status), `nav.admin`, `nav.api`, `nav.apiDocs` (API Docs), `billing.checkout.total` (Total), `docs.quickstart.sdkHeading` (SDKs), `docs.limits.colPlan` (es Plan), `policy.sections.3.heading` / `terms.sections.7.heading` (4. / 8. Cookies), `lists.itemCount.one` / `userLists.itemCount.one` (pt "{count} item"), `social.profile.counts.series.other` (es series).

Other pt-BR/es-only equal values not listed above (`item.trailerTitle`, `item.sources.steam` in es) follow the same brand/abbreviation pattern.

## Suggested owner-or-text-editor pass (not done here)
- pt-BR: Free/Pro/Business, Gameplay, Storytelling, Menu, Nickname (auth and account), Spam.
- es: Free/Pro/Business, Nickname (auth and account), Spam.
- Confirm with the owner whether "Nickname" is a deliberate product term in pt-BR and es (the app uses "nickname" elsewhere).

## Review (especialista, 2026-10-31)
verdict: APROVADO (report only), with a correction
checked: counts sampled against FIX-i18n-untranslated (pt-BR 39, es 28 equal-to-en leaves, 0 missing/extra); parity test green
correction: the "likely real misses" Free/Pro/Business, Gameplay and Storytelling are the legacy wording (legacy/apps/web/messages/{pt-BR,es}/{billing,categories}.json), so they are intentional (FIX-i18n-untranslated r1 REJEITADO). Future parity runs: a value equal to en and equal to the legacy locale value is not a miss. Nickname/Menu/Spam stay an owner copy call.
