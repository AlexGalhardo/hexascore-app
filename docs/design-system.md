# HexaScore design system (current state)

This document describes the visual system as it exists in the code on `main` at commit `439a6f9` (2026-10-31). It is derived from the source; every claim cites a path. Nothing here is a proposal: where the code and [SDD 09](sdd/09-web.md) disagree, the code is described and the difference is listed in [section 11](#11-gaps-and-inconsistencies). Do not fix code from this document; the gaps are input for a redesign.

Screenshots of production (`https://app-production-785f.up.railway.app`, taken 2026-10-31, logged out, cookie banner visible) live in [design-system/](design-system/):

| Page | Desktop 1440x900 | Mobile 390x844 |
| --- | --- | --- |
| Home `/` | [home-desktop-1440.png](design-system/home-desktop-1440.png) | [home-mobile-390.png](design-system/home-mobile-390.png) |
| Games listing `/games` | [games-desktop-1440.png](design-system/games-desktop-1440.png) | [games-mobile-390.png](design-system/games-mobile-390.png) |
| Item `/games/diablo-v` | [diablo-v-desktop-1440.png](design-system/diablo-v-desktop-1440.png) | [diablo-v-mobile-390.png](design-system/diablo-v-mobile-390.png) |
| Login `/login` | [login-desktop-1440.png](design-system/login-desktop-1440.png) | [login-mobile-390.png](design-system/login-mobile-390.png) |

## Contents

1. [Stack and file map](#1-stack-and-file-map)
2. [Tokens](#2-tokens)
3. [Theming and CSP](#3-theming-and-csp)
4. [Layout](#4-layout)
5. [Components](#5-components)
6. [Page catalogue](#6-page-catalogue)
7. [Internationalisation](#7-internationalisation)
8. [Accessibility](#8-accessibility)
9. [Performance budgets](#9-performance-budgets)
10. [Admin panel](#10-admin-panel)
11. [Gaps and inconsistencies](#11-gaps-and-inconsistencies)

## 1. Stack and file map

| Concern | Technology | Source |
| --- | --- | --- |
| Pages | Inertia 3.9.1 + React 19.3.1, one component per route in `resources/js/pages/<Area>/<Name>.tsx` | `package.json`, `resources/js/app.tsx` |
| Styling | Tailwind CSS 4.3.3 utilities only, no component CSS | `resources/css/app.css` |
| Primitives | Base UI 1.9.0 (only `Dialog` is used) | `resources/js/components/Dialog.tsx` |
| Root document | Go template with theme, locale, CSP nonce, Vite tags | `resources/views/app.html` |
| Tokens | JSON source generating CSS and TS | `packages/tokens/` |
| Mobile theme | NativeWind reading the same generated TS | `apps/mobile/tailwind.config.js`, `apps/mobile/src/theme/` |
| Admin | Go `html/template` views, not Inertia | `app/admin/views/` |
| Build | Vite 8.3.5, SSR bundle by `bun run build:ssr` | `package.json`, SDD 09 §7.10 |

There is no icon library, no web font and no component CSS file. Icons are inline SVG paths (`Nav.tsx` `Icon`, `StarRating.tsx` `STAR_PATH`).

## 2. Tokens

### 2.1 Pipeline

1. Source of truth: `packages/tokens/tokens.json` (OKLCH strings `"L C h"` per mode, plus `textPairs`, `spacing`, `fontSize`, `radius`, `touchTarget`).
2. `packages/tokens/build.ts` (`bun run --cwd packages/tokens build`) writes two committed files:
   - `packages/tokens/tokens.css`: `:root` (light) and `.dark` blocks of `--hs-<name>: oklch(...)`, then an `@theme inline` block mapping `--color-<name>: var(--hs-<name>)` and `--radius-<sm|md|lg|full>`.
   - `packages/tokens/tokens.ts`: `colors.light` / `colors.dark` with `{ oklch, hex }`, plus `spacing`, `fontSize`, `radius`, `touchTarget`. The hex values come from `packages/tokens/color.ts` `oklchToHex` (OKLab to linear sRGB, clamped to gamut).
3. Web: `resources/css/app.css` runs `@import "tailwindcss"` then `@import "../../packages/tokens/tokens.css"` and declares `@custom-variant dark (&:where(.dark, .dark *))`. Tokens therefore become utilities such as `bg-bg`, `text-fg`, `text-muted`, `bg-surface`, `border-border`, `text-accent-game`, `text-danger`, `rounded-md`.
4. Mobile: `apps/mobile/src/theme/read-tokens.cjs` parses `packages/tokens/tokens.ts` with regular expressions and feeds `apps/mobile/tailwind.config.js` (colours as `{ DEFAULT: light, dark }`, spacing in 4 px steps, font sizes, radii; `fontFamily` is `serif` for `display` and `sans-serif` for `text`). `apps/mobile/src/theme/index.ts` re-exports the constants and `palette(mode)`.
5. Guards: `packages/tokens/tokens.test.ts` asserts WCAG AA (4.5:1) for every pair in `textPairs` in both modes, the hex conversion, and that the committed outputs equal the generator output.

### 2.2 Colours

Hex values are from `packages/tokens/tokens.ts`. Utility names drop the `--color-` prefix.

| Token | Light OKLCH | Light hex | Dark OKLCH | Dark hex | Use |
| --- | --- | --- | --- | --- | --- |
| `bg` | `0.99 0.003 90` | `#fcfcf9` | `0.17 0.01 260` | `#0d1014` | page background (`body` has `bg-bg text-fg`) |
| `fg` | `0.2 0.01 260` | `#13161b` | `0.94 0.005 90` | `#ecebe7` | text, primary button fill |
| `muted` | `0.5 0.01 260` | `#606369` | `0.7 0.01 260` | `#9b9fa5` | secondary text |
| `surface` | `0.96 0.005 90` | `#f3f2ee` | `0.22 0.01 260` | `#181b1f` | cards, cover placeholder, inputs, toast |
| `border` | `0.88 0.005 260` | `#d5d7db` | `0.32 0.01 260` | `#303338` | dividers, outlines |
| `accent-game` | `0.5 0.2 290` | `#6741ca` | `0.72 0.16 290` | `#a490ff` | Games; also the generic link colour in `AppLayout` and `ResultGroup` |
| `accent-movie` | `0.52 0.19 25` | `#be222a` | `0.72 0.16 25` | `#f97770` | Movies; notification badge fill |
| `accent-series` | `0.5 0.15 220` | `#00739d` | `0.72 0.13 220` | `#09b7dc` | TV series |
| `accent-book` | `0.5 0.13 150` | `#137738` | `0.72 0.12 150` | `#69ba7c` | Books |
| `score-low` | `0.66 0.19 25` | `#f05653` | same | `#f05653` | band below 5.0 (see 2.3) |
| `score-mid` | `0.78 0.16 85` | `#e6ad00` | same | `#e6ad00` | band 5.0 to 7.4 |
| `score-high` | `0.62 0.17 145` | `#2f9f3d` | same | `#2f9f3d` | band 7.5 and above |
| `score-none` | `0.7 0.01 260` | `#9b9fa5` | same | `#9b9fa5` | no score |
| `star` | `0.78 0.16 85` | `#e6ad00` | same | `#e6ad00` | filled star (equals `score-mid`) |
| `danger` | `0.5 0.19 25` | `#b71824` | `0.72 0.17 25` | `#fd736d` | errors, destructive |

Contrast pairs enforced by the test: `fg/bg`, `fg/surface`, `muted/bg`, `muted/surface`, the four `accent-*` on `bg`, and `danger/bg`. The `score-*` and `star` tokens are not in `textPairs` (they are meant for fills, not text).

Colours used in code that are not tokens (see gaps): `indigo-600/500/400` (primary buttons, links, nav progress bar `#4f46e5` in `resources/js/app.tsx`), `zinc-300/700/600/400/900/50`, `red-600/700/400/500`, `amber-100/900/500`, `gray-300`, `lime-500`, `green-600`, `black/50` (dialog backdrop), `white` (badge and avatar text).

### 2.3 Score bands

Defined in SDD 09 §7.5: `< 5.0` low, `5.0 to 7.4` mid, `>= 7.5` high, `null` none. The tokens exist, but no web component uses them (`rg "score-low|score-mid|score-high|score-none" resources/js` returns nothing). The score is shown as plain text everywhere (section 5.4).

### 2.4 Spacing, radius, touch target, type

| Token | Values | Source | In use |
| --- | --- | --- | --- |
| spacing steps | 0, 1, 2, 3, 4, 6, 8, 12, 16 (x 4 px) | `tokens.json` | Only the mobile config consumes it; the web uses the Tailwind default scale |
| radius | `sm` 4, `md` 8, `lg` 12, `full` 9999 px | `tokens.json`, `@theme inline` | `rounded-sm` cover, `rounded-md` buttons and inputs, `rounded-lg` dialogs and menus, `rounded-full` badge and avatar |
| touch target | 44 px | `tokens.json` | `min-h-11`, `size-11`, `min-w-11` on every interactive control |
| font sizes | 12, 14, 16, 18, 20, 24, 30, 36 px | `tokens.json` | Mobile config only. Web uses Tailwind sizes: `text-xs` (11 uses), `text-sm` (110), `text-lg` (10), `text-xl` (15), `text-2xl` (33), `text-3xl` (7), `text-4xl` (2) |

Typography:

- No `@font-face`, no font files, no `font-family` override: the web renders Tailwind's default stacks (`font-sans` by default, `font-mono` for numbers via `font-mono tabular-nums`). SDD 09 §7.5 and NFR-09-7 describe self-hosted WOFF2 fonts and a serif display face; neither exists in code. Mobile maps `display` to the generic `serif` and `text` to `sans-serif`.
- Page titles: `h1` is `text-2xl font-semibold` on most pages, `text-3xl font-bold` on the item, home (`sm:text-4xl`), Edition, Pricing and Error pages. Section headings are `text-xl font-bold` or `text-xl font-semibold` and `text-2xl font-bold` on the home sections.
- Small caps style: `text-xs font-semibold uppercase tracking-wide text-muted` for the home block labels.

### 2.5 Shadows, z-index, breakpoints, motion

- Shadows: `shadow-lg` on the nav menu panel, SuggestBox listbox, toast, dialog popup and cookie banner. Button tokens in `resources/css/app.css` (derived from the colour tokens): `shadow-ledge-primary` / `-secondary` (2 px ledge plus a soft drop) and their `-pressed` 1 px variants for `btnClass`, `shadow-pressed` (inset) for ghost buttons and `iconButton`.
- Z-index (literal values, no tokens): menu panel `z-40`, SuggestBox popups and cookie banner `z-50`, dialog backdrop `z-[55]`, dialog popup `z-[60]`, toast `z-[70]`, skip link `z-[80]`.
- Breakpoints: Tailwind defaults (no custom `--breakpoint-*`). Used: `sm` 640 (search box width, footer, hero size), `md` 768 (header section nav switches between the top row and the scroller row; home block grid; docs sidebar), `lg` 1024 (item page two columns, home recent grid). The e2e projects use 1280x800 desktop and 360x740 mobile (`qa/e2e/playwright.config.ts`).
- Motion: none by design. The only animations are the `submit-spin` keyframes of `.submit-spinner` (0.8 s, slowed to 3 s under `prefers-reduced-motion: reduce`, `resources/css/app.css`) and `animate-pulse` on the CSR placeholder in `resources/views/app.html`. Buttons: every `button` and `[role="button"]` moves 1 px down while pressed (a base rule in `app.css`, 100 ms; no movement under `prefers-reduced-motion: reduce`); `btnClass` and `iconButton` add the pressed ledge or inset. No motion tokens. The Inertia progress bar uses `#4f46e5`.

## 3. Theming and CSP

### 3.1 Light and dark

- Dark mode is the `dark` class on `<html>` (`@custom-variant dark`), switching the `.dark` block of `tokens.css`.
- Resolution order, in `resources/js/app.tsx`, before Inertia mounts: `data-theme` attribute on `<html>` (written server-side by `resources/views/app.html` from the shared `theme` prop) falling back to `localStorage.theme`, falling back to `prefers-color-scheme: dark`. If storage is blocked the default (light) stays.
- The server reads the `theme` cookie (`light` or `dark`; any other value means "system") in `app/http/middleware/share_props.go` and shares it as `theme`; `RenderCache` keys cached pages on the `theme`, `consent` and `locale` cookies (`app/http/middleware/render_cache.go`).
- The toggle is `toggleTheme()` in `resources/js/components/Nav.tsx`: flips the class, writes cookie `theme` (1 year, `SameSite=Lax`, path `/`) and `localStorage.theme`. It is a single icon button (half-filled circle, `aria-label` `nav.theme`); there is no "system" option in the UI once chosen.
- Contrast rule: text pairs listed in 2.2 must stay at or above 4.5:1 in both modes (`tokens.test.ts`); Lighthouse requires accessibility 100 (section 9).
- Admin panel has no dark mode and does not use the tokens (section 10).

### 3.2 Content Security Policy constraint

- Policy (SDD 09 §7.8, `app/http/middleware/security_headers.go`): vukem default CSP plus `frame-src https://www.youtube-nocookie.com; worker-src 'self'`, per-request nonce on scripts and styles, `'strict-dynamic'`, no `'unsafe-inline'`. The JSON API uses `default-src 'none'`.
- Design consequence: no `style=` attribute and no `style={{...}}` prop anywhere in `resources/js`; every dimension is a Tailwind class. Example: `CoverThumb.tsx` keeps a `BOX` class map (`w-8 h-12`, `w-11 h-16`, `w-36 h-52`) instead of a style prop. Dynamic values must go through classes or CSSOM after mount.
- New visual systems must therefore be class-driven: no per-element computed colours, no inline SVG `style`, no CSS-in-JS runtime injection without the nonce. Third-party embeds are limited to YouTube nocookie iframes (and Turnstile when its flag is on).
- Images: `img-src 'self' data:` plus the covers bucket host. Fonts: `font-src 'self'` (a new font must be self-hosted).

## 4. Layout

### 4.1 Document

`resources/views/app.html`: `<html lang data-theme>`, charset, viewport, favicon, CSP nonce meta, CSS then JS (CSS first for LCP), `<body class="bg-bg text-fg antialiased">`. For CSR-only pages (`Profile/Index`, `Developer/ApiKeys`, `Billing/Index`, `Account/Notifications`, `Library/Index`, `Lists/Index`) the template also emits a `<title>` and a pulsing three-bar placeholder (`#csr-skeleton`, `role="status" aria-busy`) that CSS hides once `#app` is filled. While the app hydrates the loader keeps `<html inert>`; submit buttons show a spinner in that state.

### 4.2 AppLayout

`resources/js/layouts/AppLayout.tsx` wraps every Inertia page. Order of children:

1. Skip link to `#main` (`sr-only`, visible on focus, `z-[80]`).
2. `Nav` (header).
3. Deletion banner: when `auth.user.deletion_scheduled_at` is set, a full-width `role="alert"` strip (`bg-amber-100 text-amber-900`) with the purge date (scheduled date plus 30 days) and a "cancel" button (`DELETE /profile/deletion`).
4. `<main id="main" class="mx-auto max-w-6xl p-4">`: a single 1152 px centred container with 16 px padding for all pages.
5. `Footer`.
6. Toasts: `Flash` for `flash.success` or `flash.status` (kind `status`) and `flash.error` (kind `error`).
7. Live region (`aria-live="polite" role="status" sr-only`) announcing `document.title` after each navigation.
8. `CookieBanner`.

Behaviours: sets `document.documentElement.lang` on client visits; on each Inertia `navigate` moves focus to the first `<h1>` (adds `tabindex="-1"`). The focused `h1` shows the browser focus ring, visible in every screenshot as a black box around the title.

### 4.3 Header (`resources/js/components/Nav.tsx`)

Outer: `border-b border-border bg-bg`, inner `mx-auto max-w-6xl px-4`.

Desktop (from `sm`/`md`):

- Row `min-h-14`: brand link (`appName`, `text-lg font-bold`), section nav (Games, Movies, TV Series, Books; `hidden md:flex`, plain text colour), then a right-aligned group: `SuggestBox` (`sm:w-64`), theme toggle, and either the logged-out controls or the logged-in controls.
- Logged out: "Log in" link and a "More" `details` menu (hamburger icon, label `nav.more`) containing the language links, "Sign up", "Pricing", "API docs".
- Logged in: notification bell with an unread badge (`bg-accent-movie`, caps at "99+"), and an account `details` menu (user icon) with the user name, Library, Lists, Public profile (if nickname), Profile, API keys, Admin (role admin only), the language links, and Log out (`POST /logout`).

Mobile (below `md`, see [home-mobile-390.png](design-system/home-mobile-390.png)):

- The first row holds the brand and the icon group; the `SuggestBox` moves to its own full-width row (`order-last basis-full`), so the header is three rows tall: brand and icons, search, then the section row.
- A second `nav` (`md:hidden`, horizontally scrollable) shows the four sections coloured by accent (`text-accent-game`, `-movie`, `-series`, `-book`). On desktop the sections are uncoloured.
- Every control is at least 44x44 px (`min-h-11`, `size-11`).

Language switcher: a `ul` of `EN`, `PT`, `ES` links (`hrefLang`, `lang`, `aria-current` on the active one, bold and underlined). It lives inside the "More" menu (logged out) or the account menu (logged in), never in the bar itself. Clicking sets cookie `locale` and navigates to the same page under the locale prefix (`unlocalized` then `localized`, `resources/js/lib/route.ts`). Locale prefixes: `en` none, `pt-BR` `/pt-br`, `es` `/es`.

Menus are native `<details>` elements styled by `menuPanel` (`absolute right-0 z-40 w-64 rounded-lg border bg-bg p-2 shadow-lg`). They are not closed by outside click or Escape (native behaviour only).

### 4.4 Footer

`resources/js/components/Footer.tsx`: `mt-12 border-t border-border text-sm text-muted`, one row inside `max-w-6xl`: copyright with the current year and `appName`, then Terms, Privacy, Contact links (44 px high). There is no link to Docs, Pricing, Help or the language switcher in the footer.

### 4.5 Cookie banner

`CookieBanner.tsx`: fixed bottom bar (`z-50`, `bg-surface`, `border-t`, `shadow-lg`, labelled region) with copy and two buttons: "Reject" (outlined) and "Accept" (solid `bg-fg text-bg`). Choice stored in cookie `consent` (`all` or `essential`, 1 year). A spacer (`h-36 sm:h-20`) keeps the page end reachable. Consent gates YouTube embeds (`Trailers.tsx`) and GA4 (`window.vukemAnalytics.consent`, ADR 0023). The footer's "Cookie settings" button reopens it and focuses "Reject".

### 4.6 Grid and containers

- One container: `max-w-6xl` (72 rem). Pages add inner grids; there is no shared `Page`, `Container` or `Section` component and no narrower container for forms.
- Common grids: home recent `sm:grid-cols-2 lg:grid-cols-3`; home block lists `md:grid-cols-3`; item page `lg:grid-cols-[minmax(0,1fr)_24rem]` with the rating panel sticky on the right at `lg` and first on mobile (`order-1 lg:order-2`); item details `md:grid-cols-[3fr_2fr]`; library cards `sm:grid-cols-2 lg:grid-cols-3`; docs `md:grid-cols-[14rem_1fr]` with prose capped at `max-w-3xl`; definition lists `grid-cols-[max-content_1fr]`.
- Vertical rhythm is ad hoc: `grid gap-12` on home, `gap-8` on the item page, `mt-4`/`mt-6` elsewhere.

## 5. Components

All files are under `resources/js/components/` unless stated. Props are the TypeScript props in the source.

### 5.1 Identity and media

**CoverThumb** (`CoverThumb.tsx`). Item cover image.

- Props: `cover: { thumb, full } | null | undefined`, `alt: string`, `size?: "sm" | "md" | "lg"` (default `sm`), `priority?: boolean`.
- Sizes: `sm` 32x48 (`w-8 h-12`, list rows, search results, home block rows), `md` 44x64 (`w-11 h-16`, item cards, shelf cards, home recent), `lg` 144x208 (`w-36 h-52`, item page header). Ratio is about 2:3.
- Image: `<img>` with `src` = `coversBase` + `thumb` (or `full` for `lg`), `srcSet` `thumb 96w, full 600w`, `sizes` the pixel width, explicit `width`/`height`, `decoding="async"`, `object-cover rounded-sm bg-surface`. `priority` gives `loading="eager"` and `fetchPriority="high"`; otherwise `lazy`.
- Placeholder: with no cover a `span` of the same box, `bg-surface rounded-sm`, `aria-hidden`, `data-testid="cover-placeholder"`. It is a flat tinted rectangle without a title initial or icon (the Diablo V item screenshot shows it).
- Decorative by default: lists pass `alt=""`; only the item page header passes the title.

**ItemCard** (`catalog/ItemCard.tsx`). A row `li` (not a card) for listings and entity pages.

- Props: `item: Highlight { id, type, slug, title, year, score, ratingCount, cover }`, `href`.
- Layout: `flex items-center gap-3 py-2`: `CoverThumb md`, title link (`flex-1 truncate font-medium`), year (`font-mono text-xs text-muted`), `catalog.ratingCount` text, score (`w-8 text-right font-mono text-sm`, `title` attribute "Score x out of 10" or "Not enough ratings yet"; an en dash when `score` is null). No score colour, no badge.

**ShelfCard** (`library/ShelfCard.tsx`). Card used by Library, list pages and public profile tabs.

- Props: `item: ItemRef`, `stars?: number | null`, `statusLabel?: string`, `children?`.
- Layout: `article` with `rounded-md border border-border bg-surface p-2`: `CoverThumb md`, an underlined title link with the year in muted text, a muted line with the status label and "x stars" (`library.score`, an em dash when null), then `children` (edition picker, note).
- Variants are driven by the caller: Library passes status and stars; `User/List` passes a note; profile ratings pass the type as the label.

**ItemLink** (`library/ItemLink.tsx`): inline cover plus title link used in list editing rows.

**SectionHighlights** (`catalog/SectionHighlights.tsx`). Home landing block per type: heading with a "See all" link, then three ranked columns "Top rated", "New releases", "Most rated" (`md:grid-cols-3`, stacked on mobile). Rows: rank number, `CoverThumb sm`, truncated title, year or rating count, score. Empty column: "Nothing here yet." in muted text. Exports `SECTION_PATHS` (`game: games`, `movie: movies`, `series: series`, `book: books`).

### 5.2 Rating

**StarRating** (`rating/StarRating.tsx`). One row per rating category; half-star slider.

- Props: `label`, `value: number | null`, `onChange(value | null)`, `disabled?`.
- Range 0.5 to 10 in 0.5 steps, `null` is unset. Ten 24 px star SVGs (`STAR_PATH`), each `full`, `half` (left half clipped) or `empty`; filled colour `text-amber-500`, empty `text-gray-300` (neither is the `star` token).
- Pointer: hover preview, click sets the value from the pointer position (two zones per star). Touch: separate `-` and `+` 44x44 buttons (`tabIndex=-1`, with `aria-label` the row label plus a minus or plus sign).
- Keyboard on the slider: Left/Down -0.5, Right/Up +0.5, PageDown/PageUp plus or minus 1, Home 0.5, End 10, Delete clears; focus ring is the browser `focus-visible:outline-2` (not in the accent colour).
- ARIA: `role="slider"`, `aria-valuemin=0`, `aria-valuemax=10`, `aria-valuenow` (0 when unset), `aria-valuetext` `rating.valuetext` ("{value} of 10 stars") or `rating.unset`, `aria-disabled`.
- Layout: label (`sm:w-28` fixed column), `-`, stars, `+`, `<output>` with the one-decimal value or a muted dash (`w-10 tabular-nums`). Wraps below `sm` with the label on its own line.
- States: unset, hovered preview, set, disabled (while saving or logged out, buttons disabled, `aria-disabled`).

**RatingPanel** (`item/RatingPanel.tsx`). Right column of the item page.

- Props: `itemId`, `type`, `summary: RatingSummary { score, scoreByCategory, ratingCount, confidence, unusualActivity, ratingsLimitedUntil }`, `mine`, `canWrite`.
- Blocks: heading "Community score"; either "Not enough ratings. The score appears after 5 ratings." (when `score === null`) or the score as `text-4xl font-bold` plus a two-column list of the six category averages (`dt` label, `dd` value or `-`); a "See users' scores" button that lazy-loads `RatingsDialog`; the heading "Your rating" then, when logged in, six `StarRating` rows with Save/Update and Delete buttons (outlined `rounded border`, `min-h-11`), or "Log in to rate".
- Save is enabled only when all six are set; the label changes to Saving, Update or Save. The local values are optimistic; the aggregate is reloaded from the server. `confidence`, `unusualActivity` and `ratingsLimitedUntil` are in the type but never rendered.
- Categories per type (`item/types.ts` `CATEGORY_KEYS`, labels in `resources/lang/en.json` `categories`): game gameplay, fun, sound, technical, visual, storytelling; movie screenplay, acting, direction, cinematography, soundtrack, entertainment; series screenplay, acting, characters, pacing, production, entertainment; book plot, characters, writing, pacing, originality, impact.

**RatingsDialog** (`item/RatingsDialog.tsx`). Modal with a per-user score table: columns user, overall, six categories; states loading ("Loading..."), error with retry (`role="alert"`), empty ("No public ratings"), paginated with a labelled page nav (fetches `/api/items/{id}/ratings`).

**ScoreBadge.** There is no such component. The nearest thing is an inline paragraph on `Catalog/Show.tsx` (`data-testid="score-badge"`, `inline-flex w-fit rounded-md bg-surface px-3 py-1 text-lg font-bold`) that reads "Score 7.8 out of 10" or "Not enough ratings". No band colour, no size variants. Elsewhere the score is raw monospace text (`ItemCard`, `SectionHighlights`, `Home/Index`).

**HexRadar / ScoreBar / CategoryLabel.** Listed as "kept from legacy" in SDD 09 §7.5 and the i18n catalogue still has `score.radar`, but none of these components exists.

### 5.3 Item page modules (`item/`)

| Component | Purpose | States |
| --- | --- | --- |
| `StatusPicker` | `select` of the viewer's shelf status (want, in progress, completed, dropped, none); logged in only | busy (disabled) |
| `TimeToBeat` | games: medians (main, extras, completionist hours) and the viewer's report form (hours inputs, save, delete) | "not enough reports (needs 3)", report form only when status is `completed` |
| `Trailers` | up to three YouTube nocookie embeds; each is a "Load trailer" button until clicked, with a consent notice | locked until the user loads it |
| `AddToList` | per-list add or remove buttons | "no lists" empty text |
| `Editions` | books: first 10 editions with a "+N more" note | hidden when the item has no editions |
| `types.ts` | shared prop types and `CATEGORY_KEYS` | |

### 5.4 Forms and feedback

**Field** (`Field.tsx`). Labelled text input with an error line.

- Props: `label: string`, `error?: string`, plus any `<input>` attribute. Markup: `label.block > span.font-medium + input + p.text-red-600`.
- Input class: `mt-1 block w-full rounded-md border border-zinc-300 bg-transparent px-3 py-2 focus:border-indigo-500 focus:outline-none dark:border-zinc-700`. Always full width of its container; no hint, no `aria-describedby`/`aria-invalid` wiring (unlike `PasswordField`), focus outline removed in favour of a border colour change.
- States: default, focus (indigo border), error text. No disabled or success style.

**PasswordField** (`PasswordField.tsx`). Password input bound with `htmlFor`, a show/hide toggle (`aria-pressed`, `aria-controls`; labels are hard-coded English), optional live checklist of the server password rules (`role="list"`) and a four-step strength meter (`zinc`, `red-500`, `amber-500`, `lime-500`, `green-600`).

**SubmitButton** (`SubmitButton.tsx`). Primary button.

- Props: `pending?` plus button attributes except `type`. Always `type="submit"`, `data-submit`, `data-pending`, `aria-busy`, disabled while pending.
- Look: `inline-flex gap-2 rounded-md bg-indigo-600 px-4 py-2 font-medium text-white hover:bg-indigo-500 disabled:opacity-50` (about 36 px high, below the 44 px target) with a CSS spinner (`.submit-spinner`) shown when pending or while `html[inert]` (before hydration, button at 50 % opacity and `cursor: progress`).
- Other buttons are hand-written per page: outlined (`rounded-md border border-border px-3 py-1`), solid indigo (`bg-indigo-600`), inverted (`bg-fg text-bg`, cookie banner and report dialog), text links with underline. There is no `Button` component.

**Flash / toast** (`Flash.tsx`). Props `message`, `kind: "status" | "error"`. Fixed bottom-right (`right-4 bottom-4 z-[70] max-w-sm`), `rounded-md border bg-surface shadow-lg`, border and text `text-fg` or `border-danger text-danger`. `role="status"` or `role="alert"`; auto-dismiss after 6 s, paused on hover and focus; a 44x44 close button (`nav.dismiss`). One toast at a time, no stacking, no icon, no animation.

**Dialog** (`Dialog.tsx`, Base UI). Props `open`, `onOpenChange`, `title`, `description?`, `size: "sm" | "md" | "lg"` (`sm:max-w-sm|md|2xl`), `children`. `dialogPopup(size)` returns the single allowed popup class (centred, `w-[calc(100%-2rem)]`, `rounded-lg border bg-bg p-6 shadow-lg`, `z-[60]`); backdrop `bg-black/50 z-[55]`. Focus trap and Escape come from Base UI. `ReportDialog.tsx` is the main user (reason select, details textarea, submit).

**Other form widgets:** `FieldEditor` (`community/FieldEditor.tsx`, one editable field of a suggestion with a locked state), `DataExport`, `DeleteAccount` (account sections), `EditionPicker` (select plus save, loads editions on demand with a retry error), `ChallengeStatus` (live "checking" text plus a `noscript` Google fallback for the proof-of-work challenge).

### 5.5 Search

**SuggestBox** (`search/SuggestBox.tsx`). Header search and item picker.

- Props: `onPick?(suggestion)` (turns the box into a picker, used by `Lists/Edit`), `type?` (restrict to one item type), `label?`.
- Behaviour: `type="search"` input with `role="combobox"`, `aria-autocomplete="list"`, `aria-expanded`, `aria-controls`, `aria-activedescendant`. Starts at 2 characters, 150 ms debounce, aborts stale requests, fetches `/search/suggest?q=` (locale-prefixed). Keys: Down/Up cycle, Enter picks or submits to `/search?q=`, Escape closes. Options show title, year and type; 44 px rows.
- Popups: listbox `absolute right-0 z-50 mt-1 w-72 rounded-md border bg-surface shadow-lg`; "no suggestions" paragraph when the request returns nothing; an sr-only status "Searching".
- Input: `min-h-11 w-full rounded-md border border-border bg-bg px-3 sm:w-64`. Placeholder equals the label; the placeholder is truncated in the header ("Search games, movies, series, b").
- Suggestions carry no cover thumbnail.

**ResultGroup** (`search/ResultGroup.tsx`). Ordered list of hits: `CoverThumb sm`, accent-coloured title link with the matched substring in `font-semibold` (`highlight()`), and an optional muted "matched title" (alternate title). Click reporting via `navigator.sendBeacon('/search/events')`. Exports `Hit`, `reportClick`, `highlight`.

**Filters** (`catalog/Filters.tsx`). GET form: genre select, platform select (games only), year number (1800 to 2100), minimum score (0 to 10, step 0.5), sort select (popularity, score, release date, rating count, ascending and descending), Apply (indigo) and Reset link. Fixed widths (`w-40`, `w-24`) avoid layout shift. Filter state is in the URL.

### 5.6 Infrastructure components

| Component | Purpose |
| --- | --- |
| `Seo.tsx` | Head tags from the shared `seo` prop plus page overrides: title (60 chars max), description (160), canonical, hreflang alternates, robots, Open Graph, JSON-LD (escaped `<`). Props `title`, `description`, `image`, `type`, `noindex`, `robots`, `jsonLd`. |
| `CookieBanner.tsx`, `Footer.tsx`, `Nav.tsx` | See section 4. |
| `Skeleton` | Does not exist (see gaps). The only loading placeholder is the CSR block in `resources/views/app.html`; `RatingsDialog` and `EditionPicker` use plain "Loading" text. |

### 5.7 Hooks and helpers (`resources/js/lib/`)

`useT` (translator for the page locale), `useLocalized`, `route.ts` (`localized`, `unlocalized`, `LOCALE_PREFIXES`), `i18n.ts` (`{name}` placeholders, CLDR plurals via `Intl.PluralRules`, explicit `zero` wins), `preferences.ts` (cookie helpers and the consent event), `password-policy.ts`, `webauthn.ts`, `challenge.ts`, `pageWindow.ts`.

## 6. Page catalogue

Every Inertia page component in `resources/js/pages/` (33 components; `ls resources/js/pages`) is listed below with its route from `routes/*.go` and the render calls in `app/`. Public paths also exist under `/pt-br` and `/es`. "SSR" or "CSR" follows SDD 09 §6.2 (SSR exclusions per SDD 09 §6.1 `SSRConfig.Except`: `/admin`, `/profile`, `/me`, `/billing`, `/suggest`, `/suggestions`, `/library`, `/lists`). All pages render inside `AppLayout` (section 4) in the single `max-w-6xl` container. Unless stated, there is no dedicated loading state (SSR pages arrive complete; CSR pages show the `app.html` placeholder) and errors come as a toast (`flash.error`) or per-field text.

### 6.1 Public catalogue

#### Home/Index

- Route: `GET /` (SSR). File: `pages/Home/Index.tsx`. Screenshots: [desktop](design-system/home-desktop-1440.png), [mobile](design-system/home-mobile-390.png).
- Purpose: landing and discovery.
- Blocks: hero `h1` ("Every game, movie, series and book, scored by people, category by category.", `text-3xl sm:text-4xl`) with an "Explore" primary button linking to `/search`; one `SectionHighlights` per type (Games, Movies, TV Series, Books) with Top rated, New releases and Most rated columns of six rows.
- Actions: Explore, section "See all" links, row links, header search.
- Data: `sections[{type, count, topRated, newest, mostRated}]`, `trending.<type>.items[{rank, basis, genre, ...Highlight}]` (`genre`: primary genre name or null, shown next to the year; every card reads "#N trending"). The "Recently rated by members" strip and its query were removed (owner round 2026-11-01 item 5).
- States: empty column "Nothing here yet."; items without a score show an en dash.

#### Catalog/Index

- Route: `GET /games`, `/movies`, `/series`, `/books` (SSR). File: `pages/Catalog/Index.tsx`. Screenshots: [desktop](design-system/games-desktop-1440.png), [mobile](design-system/games-mobile-390.png).
- Purpose: filterable, sortable listing of one type.
- Blocks: `h1` and intro, `Filters` form, result count ("24873 items found"), then an `InfiniteScroll` list of `ItemCard` rows divided by borders.
- Actions: apply or reset filters, open an item, scroll for more (Inertia infinite scroll).
- Data: `type`, `items.data[Highlight]`, `filters`, `facets {genres, platforms}`, optional `total`.
- States: empty "No items match..." (`catalog.empty`); "N items found" is hidden when `total` is undefined; fixed-width filter fields avoid shift. Filtered and cursor pages are `noindex`.

#### Catalog/Show

- Route: `GET /{section}/{slug}` (SSR). File: `pages/Catalog/Show.tsx`. Screenshots: [desktop](design-system/diablo-v-desktop-1440.png), [mobile](design-system/diablo-v-mobile-390.png).
- Purpose: item detail, community score and rating entry.
- Blocks: header (`CoverThumb lg` priority, `h1`, "Suggest an edit" link, "Game, 2029" meta line, inline score badge); right column `RatingPanel` (sticky at `lg`, first on mobile); left column Synopsis (text, licence, source link), Details (runtime, pages, role lists such as director, developer, publisher, platform, genre as links to entity pages), Trailers (consent gated), Status picker (logged in), Time to beat (games), Editions (books), Add to list (logged in); attribution footer line.
- Actions: rate (six categories), change status, report time to beat, add to a list, load trailers, open entity pages, suggest an edit, see users' scores.
- Data: `item`, `ratingSummary`, `entities` by role, `editions`, `editionCount`, `timeToBeat`, `me`, `myLists`, `trailers`, `cover`; JSON-LD per type (`VideoGame`, `Movie`, `TVSeries`, `Book`) with `aggregateRating` only when a score exists.
- States: no cover (flat placeholder), no synopsis ("No synopsis yet."), no score ("Not enough ratings" twice: badge and panel), logged out ("Log in to rate"), empty "Details" column when no fields exist (production Diablo V shows a bare heading), trailers locked until consent. `SDD 09` mentions a deferred "recent changes" block; it is not implemented.

#### Catalog/Edition

- Route: `GET /books/{slug}/editions/{id}` (SSR). File: `pages/Catalog/Edition.tsx`.
- Purpose: one book edition.
- Blocks: `h1` title, "Suggest an edit" link, subtitle, definition list (pages, ISBN, language, publisher, year, format; null rows hidden), back-to-work link.
- States: no cover shown at all; no empty state beyond hidden rows.

#### Entity/Show

- Route: `GET /people|companies|franchises|platforms|genres|game-modes/{slug}`, `/countries|languages/{code}` (SSR). File: `pages/Entity/Show.tsx`.
- Purpose: everything linked to a person, company, franchise, platform, genre, mode, country or language.
- Blocks: kind label, `h1` name, "Suggest an edit", life or founding facts (born, died, founded), licensed description with source, role chips (`Director (3)`), then one section per item type with "Showing x of y", average score and `ItemCard` rows, and a "See all" or "Next" link.
- States: empty `entity.empty`; entities with fewer than three items are `noindex,follow`.

#### Search/Index

- Route: `GET /search` (SSR, `noindex,follow`). File: `pages/Search/Index.tsx`.
- Purpose: full-page search.
- Blocks: `h1` ("Results for ..."), search form (input plus indigo submit), type filter nav (All, Game, Movie, TV series, Book; current has `aria-current` and a `surface` background), grouped results (`ResultGroup` per type with count and "more" link) or flat paged results with Previous and Next, then matched entity links.
- States: empty query prompt, "too short" notice, "no results for q" with a "did you mean" link; no loading state (full navigation).

### 6.2 Community and social

#### User/Profile

- Route: `GET /u/{nickname}` (SSR). File: `pages/User/Profile.tsx`.
- Purpose: public profile.
- Blocks: header (64 px round avatar with initial on `bg-accent-game`, nickname `h1`, summary of ratings and lists, Report dialog for other viewers), tab nav (Ratings, Status, Lists via `?tab=`), tab content: infinite list of `ShelfCard` ratings, per-type status counts, or the user's public lists.
- States: private profile shows only the nickname plus a notice (visitor) or an owner hint; empty rating, shelf and list messages; `noindex` unless indexable.

#### User/List

- Route: `GET /u/{nickname}/lists/{id}` (SSR). File: `pages/User/List.tsx`.
- Purpose: a public list.
- Blocks: `h1` title, byline link, item count, private flag, description, edit link (owner), report dialog (other logged-in users), ordered `ShelfCard` rows with optional notes.
- States: empty list message; `noindex` unless indexable.

#### Community/SuggestEdit

- Route: `GET /suggest/{target_type}/{id}`, saves via `POST|PATCH /suggestions` (CSR, `noindex`, verified users). File: `pages/Community/SuggestEdit.tsx`.
- Purpose: propose corrections to items, editions and entities.
- Blocks: `h1`, remaining quota ("x of y left today"), draft-restored note, error list (`role="alert"`), a `FieldEditor` per editable field (with locked fields), locale select, sources textarea, note textarea, Save draft and Submit buttons (`max-w-xl` form).
- States: "suggestions closed" view when the feature flag is off, locked fields, restored draft (stored in the browser), validation errors, quota exhausted.

### 6.3 Library and lists (CSR, logged in and verified)

#### Library/Index

- Route: `GET /library` (CSR, `noindex`). File: `pages/Library/Index.tsx`.
- Purpose: personal shelves.
- Blocks: `h1`, filter form (status, type, sort; counts in options), sections per status (Want, In progress, Completed, Dropped) each a 1 to 3 column grid of `ShelfCard`s with an edition picker or edition label for books, per-type "showing x of y" with "see all" or "next".
- States: `library.empty`; CSR placeholder before hydration.

#### Lists/Index

- Route: `GET /lists` (CSR, `noindex`). File: `pages/Lists/Index.tsx`.
- Purpose: manage own lists.
- Blocks: `h1`, per-type limit note, divided rows (title link to edit, type, count, public or private, red "hidden by moderator" note), "New list" form (type select, title `Field`, public checkbox, Create).
- States: `lists.noLists` empty text; processing disables Create; validation error under the title.

#### Lists/Edit

- Route: `GET /lists/{id}/edit` (CSR, `noindex`). File: `pages/Lists/Edit.tsx`.
- Purpose: edit list details and order.
- Blocks: back link, `h1`, details form (title, description textarea, public checkbox, Save, Delete via `window.confirm`), "Items" heading with a `SuggestBox` picker, ordered rows with position, `ItemLink`, Move up, Move down and Remove buttons, Save order.
- States: empty list text; Save order enabled only when the order changed (optimistic version number sent); focus returns to the moved row's button.

#### Account/Notifications

- Route: `GET /me/notifications` (CSR). File: `pages/Account/Notifications.tsx`.
- Purpose: in-app notifications.
- Blocks: `h1`, "Mark all read" (when any unread), infinite list; each row is a localised sentence (`notifications.kind.<kind>`) with a locale-formatted date, bold when unread, and a "Mark read" button.
- States: `notifications.empty`.

#### Dashboard

- Route: `GET /dashboard` (logged in, verified). File: `pages/Dashboard.tsx`.
- Purpose: post-login hub.
- Blocks: `h1` "Welcome, name", lead text, verify-email prompt when unverified, a card link grid (Library, Lists, public page, API keys) each with a help line, and a two-factor suggestion link.
- Note: not in the header or footer; reached after login and from `/two-factor`.

### 6.4 Account and billing

#### Profile/Index

- Route: `GET /profile` (CSR). File: `pages/Profile/Index.tsx`.
- Purpose: account settings, a single long page of anchored sections (`#details`, `#password`, `#passkeys`, `#two-factor`, `#sessions`, `#privacy`, `#export`, `#delete`).
- Blocks: Details (name, nickname, locale), Password (set or change), Passkeys (list, add, remove with last-passkey warning), Two-factor (status and manage link), Sessions (list, revoke, revoke all), Privacy (public profile, personalisation), Data export (`DataExport`), Delete account (`DeleteAccount`, 30-day grace).
- States: empty sessions and passkeys, failed passkey registration (`role="alert"`), pending form buttons; CSR placeholder before hydration.

#### Developer/ApiKeys

- Route: `GET /profile/api` (CSR); create `POST /profile/api/keys`, revoke `DELETE /profile/api/keys/{id}`. File: `pages/Developer/ApiKeys.tsx`.
- Purpose: manage public-API keys.
- Blocks: `h1`, intro with a docs link, one-time "new key" panel (alert, copy and dismiss buttons), create form (`Field` plus button, max keys), table (name, prefix, created, last used, status, actions) with an inline revoke confirmation.
- States: `api.profile.keys.empty`, copied state, never-used and revoked labels, create pending.

#### Billing/Index

- Route: `GET /billing` (CSR, verified), `POST /billing/checkout/{plan}`, `POST /billing/portal`. File: `pages/Billing/Index.tsx`.
- Purpose: subscribe to or manage a paid API plan (Stripe).
- Blocks: `h1`, success or cancelled notice after checkout, current plan with period end and "Manage billing", plan purchase links with a currency selector.
- States: subscribed vs not, checkout success, cancelled.

#### Pricing/Index

- Route: `GET /pricing` (SSR). File: `pages/Pricing/Index.tsx`.
- Purpose: public API plans.
- Blocks: `h1` and subtitle, currency switcher links, plan cards (name, price, quota, max page size, rate limit, CTA that depends on login and plan), coverage section with a link.
- States: free plan price 0; CTA text varies (sign up, current plan).

### 6.5 Authentication (all `noindex`; all share one unstyled full-width layout)

All Auth pages render `h1` (`text-2xl font-semibold`), a form with `space-y-4` and links in `text-indigo-600 dark:text-indigo-400`. Forms are not width-constrained: at 1440 px the inputs span the full 1120 px container ([login-desktop-1440.png](design-system/login-desktop-1440.png)); at 390 px they are naturally narrow ([login-mobile-390.png](design-system/login-mobile-390.png)).

| Component | Route | Purpose and blocks | States |
| --- | --- | --- | --- |
| `Auth/Login` | `GET /login` | Email or nickname `Field`, `PasswordField`, remember-me checkbox, `SubmitButton`, `ChallengeStatus`, "Forgot your password?" and "Create account" links, outlined "Sign in with a passkey" button | per-field errors, passkey failure `role="alert"`, pending and "checking" state |
| `Auth/Signup` | `GET /signup` | Name, email, password with rule checklist and strength meter, confirmation, honeypot form token, `SubmitButton`, "Have an account? Log in" | per-field errors, challenge check |
| `Auth/CheckEmail` | `GET /signup/check-email` | Notice after sign-up and a login link | static |
| `Auth/VerifyNotice` | `GET /verify-email` | "Sent to email" notice, resend link | static, flash toast on resend |
| `Auth/Onboarding` | `GET /onboarding` | Nickname `Field` with hint and submit | error codes `nickname-taken`, `invalid-nickname` mapped to catalogue text |
| `Auth/Recover` | `GET /recover-password` | Email field, submit, `ChallengeStatus` | generic success via flash |
| `Auth/Reset` | `GET /reset-password/{token}` | Email, new password, confirmation, submit | per-field errors |
| `Auth/TwoFactorChallenge` | `GET /two-factor-challenge` | Code or recovery-code `Field`, verify, switch between authenticator and backup code | per-field error |
| `Auth/TwoFactor` | `GET /user/two-factor` | Enable button; QR code (200 px) with setup key and confirm form; enabled view with recovery codes, regenerate and disable forms (each asks for the password); back to dashboard | three modes (off, pending, on); all copy is hard-coded English |

### 6.6 Content, help and system pages

#### Docs/Page

- Route: `GET /docs`, `/docs/{slug}` (SSR); slugs from `resources/docs/`: `index`, `authentication`, `errors`, `pagination`, `rate-limits`, `coverage`, `changelog`. The Scalar reference at `/docs/reference` is a separate Go page. File: `pages/Docs/Page.tsx`.
- Purpose: public API documentation.
- Blocks: left column nav (pages, "API reference", on-page table of contents of `h2`s), prose article (`max-w-3xl`, build-time sanitised Markdown HTML, headings and code styled with arbitrary-variant classes).
- States: none beyond 404.

#### Legal/Page

- Route: `GET /terms`, `/policy` (SSR). Purpose: legal text. Blocks: `h1`, "Last updated" date, sections with `h2`. Copy comes from the catalogue.

#### Help/Score

- Route: `GET /help/score` (SSR). Purpose: explains how the score is built. Blocks: `h1`, intro, categories list for the four types (`categoryKeys`), several explanatory sections.

#### Help/DeleteAccount

- Route: `GET /help/delete-account` (SSR, public; URL used for Google Play). Purpose: deletion policy. Blocks: `h1`, grace period, what is erased, what is kept, links to `/profile#delete` and `/login`.

#### Contact/Index

- Route: `GET /contact`, `POST /contact` (SSR). Purpose: contact form. Blocks: `h1`, intro, name, email, topic select, message textarea with hint, hidden honeypot, submit. States: `contact.sent` confirmation (`role="status"`), per-field errors.

#### Error

- Route: catch-all `GET /{path...}` and error renders (SSR, `noindex`). Purpose: 404 and 5xx. Blocks: status code (`text-accent-game`), `h1` heading, body copy for 404 and 5xx, "Back home" link; centred with `py-12`. Other statuses show only the heading.

### 6.7 Pages outside Inertia

- `/admin/*`: Go templates, see section 10.
- `/docs/reference`: Scalar API reference with its own nonce CSP.
- Error 500 fallback: localised `errors/500.html` (SDD 09 FR-09-9).

### 6.8 Routes in the SDD without a page

`Catalog/RecentChanges` (`/{section}/{slug}/changes`) is in SDD 09 §6.2 but there is no component; `curl -I /games/diablo-v/changes` on production returns 404.

## 7. Internationalisation

- Locales: `en` (default, unprefixed), `pt-BR` (`/pt-br`), `es` (`/es`). Catalogues `resources/lang/{en,pt-BR,es}.json`, 35 top-level namespaces in each (`flash`, `account`, `actions`, `api`, `contact`, `auth`, `billing`, `catalog`, `categories`, `itemTypes`, `score`, `meta`, `nav`, `home`, `footer`, `cookies`, `sandbox`, `docs`, `errors`, `item`, `updatedLabel`, `policy`, `terms`, `lists`, `ratings`, `search`, `social`, `help`, `rating`, `notifications`, and others).
- Delivery: `inertia.I18n(...)` shares `locale`, `locales` and flat `messages` (SDD 09 §7.1). `useT()` fills `{name}` placeholders and picks CLDR plural forms (`resources/js/lib/i18n.ts`). A missing key renders the key itself.
- Routing: prefix is authoritative; unprefixed GETs redirect (307) by cookie `locale` then `Accept-Language`. Each indexable page emits canonical plus `hreflang` `en`, `pt-BR`, `es`, `x-default` (`Seo.tsx`).
- Formatting: scores one decimal; dates through `toLocaleDateString` or `toLocaleString` with the locale (`AppLayout`, `Notifications`).
- Layout implications for a redesign: Portuguese and Spanish strings are longer than English; the header, filters and buttons must tolerate it (filter selects use fixed widths for CLS).
- Hard-coded English: `Auth/TwoFactor.tsx` (the whole page), `PasswordField.tsx` toggle labels ("Show password", "Hide password"), `CookieBanner` `aria-label="Cookies"`, the admin panel. See gaps.

## 8. Accessibility

- Targets: WCAG 2.2 AA, axe zero serious or critical (SDD 09 NFR-09-11), Lighthouse accessibility 100 (`lighthouserc.cjs`).
- Tooling: `@axe-core/playwright` 4.14.0 in `qa/e2e/helpers/axe.ts` (`expectNoSeriousA11yViolations`, tags `wcag2a`, `wcag2aa`, `wcag21aa`; moderate and minor are tolerated). Used by 20 of the 29 e2e specs. E2E projects: desktop 1280x800 and mobile 360x740, each in three locales.
- Structure: skip link, landmarks (`header`, `nav` with labels, `main#main`, `footer`), one `h1` per page, focus moved to the `h1` after each navigation, polite live region announcing the title.
- Keyboard: every control is reachable; the star slider has the full key map (section 5.2); SuggestBox follows the ARIA 1.2 combobox pattern; dialogs trap focus (Base UI); `Lists/Edit` restores focus after reordering.
- Targets: 44 px minimum on interactive controls (`min-h-11`, `size-11`); exceptions in gaps.
- Focus: `focus-visible:outline-2 focus-visible:outline-offset-2` on nav items, buttons and the slider, using the browser default outline colour rather than a token (SDD says "2 px ring in the type accent").
- Reduced motion: the spinner slows; there is no other motion to disable. The star widget has no animation.
- Forms: errors are text under the field; `PasswordField` wires `aria-invalid` and `aria-describedby`, `Field` does not.
- Screen readers: icons are `aria-hidden` with `aria-label` on their buttons; the score has an sr-only "x of 10 stars" suffix; `dl` pairs are used for details.
- JavaScript off: SSR pages render the `h1`, content and meta (e2e `ssr-nojs.spec.ts`); the filters and the search form are plain GET forms.
- CSP violations: zero expected, asserted by `qa/e2e/specs/csp.spec.ts`.

## 9. Performance budgets

From SDD 09 §4 and enforced where noted.

| Budget | Value | Enforced by |
| --- | --- | --- |
| LCP (mobile, Moto G Power class, slow 4G) | at most 2.0 s | `lighthouserc.cjs` (`largest-contentful-paint` 2000 ms, median of 3) |
| INP | 150 ms field p75; lab proxy TBT at most 200 ms | `total-blocking-time` 200 ms in `lighthouserc.cjs` |
| CLS | at most 0.05 | `lighthouserc.cjs` |
| TTFB p75 | 400 ms SSR (150 ms on props-cache hit) | `server-response-time` 400 ms in `lighthouserc.cjs`; k6 `qa/k6/pages-ttfb.js` |
| JS on item page | 150 kB total (153600 bytes), 90 kB before interaction | `resource-summary:script:size` 153600 |
| CSS | 30 kB (30720 bytes) | `resource-summary:stylesheet:size` |
| Fonts | at most 3 WOFF2 files, 100 kB total, `font-display: swap` | `resource-summary:font:count` 3 and `font:size` 102400 (the app ships none) |
| Lighthouse (mobile) | performance at least 90, accessibility 100, SEO 100, best practices at least 95 | `lighthouserc.cjs` categories |
| Shared i18n prop | at most 12 kB gzip per locale | unit test on catalogue size |
| SQL per page | no N+1; item page at most 12 cold queries | dev checklist |
| Viewport | usable at 360 px | e2e mobile project |

Lighthouse CI configuration: `lighthouserc.cjs` (`formFactor: mobile`, 412x823 at 1.75 DPR, simulated throttling, 3 runs, URLs `/`, `/games`, `/games/adipiscing-consectetur`, `/books/consectetur-2004`, `/people/ana-oliveira`); run with `bunx --bun @lhci/cli@0.15.1 autorun` and `E2E_BASE_URL`. Design levers in the current code: CSS is first in the document head, the item cover is eager with `fetchpriority=high`, list covers are lazy with fixed `width`/`height`, filter selects have fixed widths, the CSR placeholder prevents a blank document, only `/build` assets are gzip-encoded and pages are render-cached for 60 s (SDD 09 §7.2).

## 10. Admin panel

Not Inertia and not tokenised: `app/admin/views/layouts/admin.html` plus `ingestion.html`, `items_manage.html`, `show.html` and the vukem admin resource views, styled by the vukem admin CSS (`{{.Panel.CSS}}`, `{{.Panel.JS}}`).

- Layout: `topbar` (brand `Panel.Title`, "Back to site" link), `shell` with a left `nav.sidebar` ("Resources": Dashboard, one link per registered resource with `aria-current="page"`, Feature flags, Ingestion, Audit log) and `main#main`.
- Flash: `p.flash role="status"` (`admin.success`) and `p.flash.flash-warning role="alert"`.
- Content patterns: `toolbar` with `h1` and `buttons` (`a.button`, `.danger`), `dl.record` detail grid, `form.stack` with labels, selects and textareas (report resolution with outcome select and note), CSRF field via `csrfField`.
- Language: English only, `lang="en"`, `noindex, nofollow`; no theme toggle, no locale switcher, no shared components with the app.
- Resources are registered in `app/admin/registry.go` and `app/admin/*_resource*.go` (items, users, reports, scoring, etc.).

## 11. Gaps and inconsistencies

Listed only; none were changed.

### Tokens and theming

- G1: Component colours bypass the tokens: primary buttons and links use `indigo-*` (`SubmitButton`, `Filters`, `Home`, `Search`, `Lists`, `Auth/*`, `Billing`, `Pricing`), inputs use `zinc-*` (`Field`, `Search`, `Lists/Edit`), errors use `red-*`, the banner uses `amber-*`, the stars use `amber-500` and `gray-300`, the password meter uses five different Tailwind colours. The `danger` token is used in only a few places (`Flash`, `EditionPicker`).
- G2: The Inertia progress bar colour `#4f46e5` is hard-coded in `resources/js/app.tsx`.
- G3: `score-low`, `score-mid`, `score-high`, `score-none` and `star` tokens are unused on the web; scores have no band colour and no badge component. `ScoreBadge`, `Skeleton`, `HexRadar`, `ScoreBar`, `CategoryLabel`, a `Button`, `Notice` and `Input` component do not exist, although SDD 09 §7.5 lists several as kept from legacy and `P2.H.2.1` names `Skeleton.tsx`.
- G4: The spacing and font-size scales are only consumed by mobile; the web uses Tailwind's defaults, with sizes (`text-xs`, `text-4xl`) outside the 12 to 36 px token scale.
- G5: No font is loaded (SDD 09 NFR-09-7 describes self-hosted WOFF2, a serif display face and tabular mono). The web depends on the platform default stack, so rendering differs per OS; mobile declares `serif` for display which the web never uses.
- G6: No shadow, z-index, motion or breakpoint tokens; z-indexes are literals (`z-40` to `z-[80]`).
- G7: The theme toggle has no "system" state once a cookie is set, and the icon does not reflect the current mode.
- G8: `AppLayout` uses `text-accent-game` for the generic `link` constant while Auth pages use `text-indigo-600`; `ResultGroup` titles are `accent-game` regardless of type.

### Layout and responsive

- G9: A single `max-w-6xl` container is used for everything: forms (Login, Signup, Account, Contact) stretch to about 1120 px on desktop ([login-desktop-1440.png](design-system/login-desktop-1440.png)); `Lists/Index` and `Community/SuggestEdit` use their own `max-w-md` or `max-w-xl`; there is no page header, section or card primitive.
- G10: The mobile header takes three rows (brand and icons, search, section links); the search placeholder is truncated on desktop ("Search games, movies, series, b").
- G11: Section nav colours differ between mobile (accent colours) and desktop (plain).
- G12: The language switcher is hidden inside the "More" and account menus; the footer has no language links and no links to Docs, Pricing or Help (`/help/score` has no inbound link in the UI).
- G13: `details` menus do not close on outside click, Escape or navigation (not verified in all browsers).
- G14: The item page header shows an empty grey box when there is no cover and an empty "Details" section when no fields exist (both visible on production Diablo V).
- G15: Rows show "0" and a dash for items without ratings (`ItemCard`: `catalog.ratingCount` text, then `–`); the home highlights show a bare numeric `0` (see [home-desktop-1440.png](design-system/home-desktop-1440.png)).
- G16: The focused `h1` after navigation shows a heavy default focus rectangle on every page, because `tabindex=-1` focus is not styled with `:focus-visible` rules.
- G17: Filter form on `/games` is five stacked inputs on mobile before any result; no collapse.

### Components

- G18: `SubmitButton` is about 36 px high (below the 44 px target); the home "Explore" button and many `py-1` buttons (`Lists/*`, `Library`, `Lists/Edit` `button`) are also under 44 px; the `StarRating` `-`/`+` buttons are plain bordered squares.
- G19: `Field` lacks `aria-invalid`/`aria-describedby`, uses `focus:outline-none` with only a border colour change, and has no hint, disabled or success style; `Lists/Index`, `Library` and `Contact` selects use a different border and background than `Field`.
- G20: Only one toast at a time; no stacking, icon or animation; `Lists/Edit` deletes with `window.confirm` while API keys use an inline confirmation and `ReportDialog` uses a modal.
- G21: `RatingPanel` ignores `confidence`, `unusualActivity` and `ratingsLimitedUntil` (rate-limit and moderation states have no UI).
- G22: `PasswordField` toggle labels and `Auth/TwoFactor` copy are hard-coded English; the admin panel is English only.
- G23: Docs prose is styled by a long arbitrary-variant class string on the `article`; there is no shared typography style for Legal, Help and Docs.

### Pages and routes

- G24: `Developer/ApiKeys` links to `/docs/api`, which returns 404 on production; the docs slugs are `authentication`, `errors`, `pagination`, `rate-limits`, `coverage`, `changelog`.
- G25: SDD 09 §6.2 lists `Catalog/RecentChanges` (and `Defer`ed recent changes on the item page); no component exists and the route returns 404.
- G26: `Dashboard` is not linked from the header, account menu or footer; `/dashboard` is only reached after login and from the two-factor page.
- G27: `Lists/*` and `Library` have the CSR placeholder but no in-page skeletons; `RatingsDialog`, `EditionPicker` and `TimeToBeat` use plain "Loading" text.
- G28: `Search/Index` and `Notifications` use `<Head>` directly instead of `<Seo>` (no `noindex` through the shared helper), and `Search` hard-codes `/search` as the form action (not locale prefixed).
- G29: There is no user avatar upload (initial on `bg-accent-game` only) and no cover for editions or entities.

### Accessibility and performance

- G30: Focus rings use the browser default colour rather than the type accent specified by SDD 09 §7.4; the nav menu items use `hover:text-accent-game`, which is not a contrast-tested pair on `surface`.
- G31: Lighthouse and axe run only against DemoCatalog URLs in CI, not production; production screenshots were visual only. Axe tolerates moderate violations. Dark mode is not covered by the screenshots or Lighthouse configuration (dark contrast is covered only by the token test).
- G32: The JS/CSS budgets are transfer sizes of a cold page (Lighthouse `resource-summary`); no budget exists for images or for the CSR pages (library, lists, profile, billing, notifications).

### Not verified

- Computed focus colours, real dark-mode rendering, the logged-in header and every authenticated page were not captured: production was browsed read-only without logging in.
- Exact rendered fonts per OS; `font-mono` stack.
- Whether `/profile/api/keys` and `/billing` states match the code in production.
