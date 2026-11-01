# Concept-art prompts for the HexaScore site redesign

Prompts for ChatGPT image generation (or any image model), one set per page of the current site. The owner picks a style, generates the images, and drops them into the repo as visual references for a full redesign. Page content comes from the code inventory in [design-system.md](design-system.md); screenshots of the current production site are in [design-system/](design-system/).

## How to use this file

1. Pick one style from the [glossary](#style-glossary) and keep it for every page so the set is coherent. Mixing styles per page is fine for exploration, but a redesign reference should come from one style.
2. For each page, copy the whole prompt of that page and style (everything inside its code fence) into a new ChatGPT image conversation. Each prompt is self-contained; do not paste several at once.
3. Generate the desktop image first. Then ask for the mobile variant with the last paragraph of the prompt ("Mobile variant"), or send that paragraph as a follow-up in the same conversation so the model keeps the look.
4. If text in the image is garbled, ask for "same layout, fix the typography and keep the exact copy". Small copy errors do not matter; layout, hierarchy, colour and component shape do.
5. Save the PNGs in `docs/design-concepts/` using the [naming convention](#naming-convention) and commit them in a separate change.

## Style glossary

Three art directions. Their names, palettes and type pairings are repeated verbatim inside every prompt, so a prompt works on its own.

### Editorial Press (`editorial`)

A printed culture magazine or newspaper review section: calm, authoritative, text-forward. Palette: paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Type pairing: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius).

### Neon Noir Arcade (`noir`)

A dark cinematic gaming dashboard: moody, high contrast, energetic but readable. Palette: void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Type pairing: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent.

### Soft Paper Play (`paper`)

A soft, minimal, playful product: approachable, light, a little whimsical, never childish. Palette: cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Type pairing: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders.

All three share the product facts that must not change: four categories with their own accent in each style, six rating categories per type (half-star scale, 0.5 to 10), an out-of-10 score with four bands (below 5.0, 5.0 to 7.4, 7.5 and above, no score), the cover aspect ratio of about 2:3, a header with search, theme toggle and menu, and a content width of about 1152px.

## Naming convention

```text
docs/design-concepts/<page-id>-<style>.png      desktop 1440x900
docs/design-concepts/<page-id>-<style>-m.png    mobile 390x844
```

- `<page-id>` is the id in the heading of each page below (for example `catalog-index`, `item`, `login`).
- `<style>` is `editorial`, `noir` or `paper`.
- Example: `docs/design-concepts/item-noir.png` and `docs/design-concepts/item-noir-m.png`.
- If several attempts exist, add `-v2`, `-v3` before the extension.

## Constraints the redesign must keep

- [ ] No inline `style` attributes: the CSP allows nonce scripts and styles only, so every colour, size and spacing must be a class (Tailwind utilities driven by CSS variables). Avoid designs that need per-element computed styles, runtime CSS injection or external fonts and CDNs; fonts must be self-hosted WOFF2.
- [ ] Text contrast of at least 4.5:1 (WCAG 2.2 AA) in light and dark; the token test enforces it for the pairs in `packages/tokens/tokens.json`. Interactive targets at least 44x44px, visible focus rings, a skip link, one `h1` per page, nothing that depends on motion (respect reduced motion).
- [ ] Three locales: English, Portuguese (pt-BR) and Spanish. Strings are longer in pt-BR and es, so no fixed-width labels, no text baked into images, and filter bars and menus must wrap.
- [ ] Light and dark themes from one token set (`packages/tokens`), which also feeds the Android app.
- [ ] Performance budgets (SDD 09 section 4): LCP at most 2.0s on mobile, CLS at most 0.05 (fixed cover boxes and filter widths), JS at most 150kB, CSS at most 30kB, at most 3 font files and 100kB of fonts, Lighthouse accessibility 100.
- [ ] Layouts usable at 360px, no horizontal page scroll, a single content column on mobile.
- [ ] Public pages work without JavaScript (server-rendered content, plain GET forms for search and filters).
- [ ] No real brand logos or third-party artwork in the design; covers come from the catalogue at runtime.

## Prompts

Pages follow the page catalogue in [design-system.md](design-system.md#6-page-catalogue), plus the global overlays and the admin panel; the id in each heading is the file-name prefix. Every page has one prompt per style.

### 1. Home (`home`)

Route: `/`. Landing page that invites people to explore and shows what is trending across the four categories.

#### Home - Editorial Press

File: `docs/design-concepts/home-editorial.png` (and `-m.png`)

```text
Design a concept screenshot of the Home page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Editorial Press" - a printed culture magazine or newspaper review section: calm, authoritative, text-forward.

Page purpose: Landing page that invites people to explore and shows what is trending across the four categories. Route: /.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. A hero with the headline 'Every game, movie, series and book, scored by people, category by category.' and an 'Explore' button; a 'Recently rated' grid of 6 entries (cover, title, 'rated by maya', score); then four stacked category sections (Games, Movies, TV Series, Books), each with a 'See all' link and three ranked columns of six rows: Top rated, New releases, Most rated. Each row: rank number, small 32x48 cover, title, year or rating count, score.

States to show: Show one column with the empty text 'Nothing here yet.' and one row with no score (grey dash).

Visual language: colour palette paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Typography: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius). All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Home - Neon Noir Arcade

File: `docs/design-concepts/home-noir.png` (and `-m.png`)

```text
Design a concept screenshot of the Home page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Neon Noir Arcade" - a dark cinematic gaming dashboard: moody, high contrast, energetic but readable.

Page purpose: Landing page that invites people to explore and shows what is trending across the four categories. Route: /.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. A hero with the headline 'Every game, movie, series and book, scored by people, category by category.' and an 'Explore' button; a 'Recently rated' grid of 6 entries (cover, title, 'rated by maya', score); then four stacked category sections (Games, Movies, TV Series, Books), each with a 'See all' link and three ranked columns of six rows: Top rated, New releases, Most rated. Each row: rank number, small 32x48 cover, title, year or rating count, score.

States to show: Show one column with the empty text 'Nothing here yet.' and one row with no score (grey dash).

Visual language: colour palette void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Typography: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the dark theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Home - Soft Paper Play

File: `docs/design-concepts/home-paper.png` (and `-m.png`)

```text
Design a concept screenshot of the Home page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Soft Paper Play" - a soft, minimal, playful product: approachable, light, a little whimsical, never childish.

Page purpose: Landing page that invites people to explore and shows what is trending across the four categories. Route: /.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. A hero with the headline 'Every game, movie, series and book, scored by people, category by category.' and an 'Explore' button; a 'Recently rated' grid of 6 entries (cover, title, 'rated by maya', score); then four stacked category sections (Games, Movies, TV Series, Books), each with a 'See all' link and three ranked columns of six rows: Top rated, New releases, Most rated. Each row: rank number, small 32x48 cover, title, year or rating count, score.

States to show: Show one column with the empty text 'Nothing here yet.' and one row with no score (grey dash).

Visual language: colour palette cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Typography: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

### 2. Catalog listing (Games) (`catalog-index`)

Route: `/games`. Browse, filter and sort every game; the same layout serves Movies, TV Series and Books.

#### Catalog listing (Games) - Editorial Press

File: `docs/design-concepts/catalog-index-editorial.png` (and `-m.png`)

```text
Design a concept screenshot of the Catalog listing (Games) page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Editorial Press" - a printed culture magazine or newspaper review section: calm, authoritative, text-forward.

Page purpose: Browse, filter and sort every game; the same layout serves Movies, TV Series and Books. Route: /games.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Page title 'Games' with a one-line intro; a filter bar (Genre select, Platform select, Year, Min. score, Sort by, 'Apply' button, 'Reset' link); a result count '24,873 items found'; an infinite list of rows (cover 44x64, title, year, rating count, score badge) with a 'Load more' link at the end.

States to show: Include rows with and without scores; the last row shows the 'Not enough ratings' dash.

Visual language: colour palette paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Typography: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius). All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Catalog listing (Games) - Neon Noir Arcade

File: `docs/design-concepts/catalog-index-noir.png` (and `-m.png`)

```text
Design a concept screenshot of the Catalog listing (Games) page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Neon Noir Arcade" - a dark cinematic gaming dashboard: moody, high contrast, energetic but readable.

Page purpose: Browse, filter and sort every game; the same layout serves Movies, TV Series and Books. Route: /games.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Page title 'Games' with a one-line intro; a filter bar (Genre select, Platform select, Year, Min. score, Sort by, 'Apply' button, 'Reset' link); a result count '24,873 items found'; an infinite list of rows (cover 44x64, title, year, rating count, score badge) with a 'Load more' link at the end.

States to show: Include rows with and without scores; the last row shows the 'Not enough ratings' dash.

Visual language: colour palette void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Typography: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the dark theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Catalog listing (Games) - Soft Paper Play

File: `docs/design-concepts/catalog-index-paper.png` (and `-m.png`)

```text
Design a concept screenshot of the Catalog listing (Games) page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Soft Paper Play" - a soft, minimal, playful product: approachable, light, a little whimsical, never childish.

Page purpose: Browse, filter and sort every game; the same layout serves Movies, TV Series and Books. Route: /games.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Page title 'Games' with a one-line intro; a filter bar (Genre select, Platform select, Year, Min. score, Sort by, 'Apply' button, 'Reset' link); a result count '24,873 items found'; an infinite list of rows (cover 44x64, title, year, rating count, score badge) with a 'Load more' link at the end.

States to show: Include rows with and without scores; the last row shows the 'Not enough ratings' dash.

Visual language: colour palette cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Typography: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

### 3. Item detail (`item`)

Route: `/games/hollow-meridian`. Everything about one title with its community score, six category scores and the rating widget.

#### Item detail - Editorial Press

File: `docs/design-concepts/item-editorial.png` (and `-m.png`)

```text
Design a concept screenshot of the Item detail page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Editorial Press" - a printed culture magazine or newspaper review section: calm, authoritative, text-forward.

Page purpose: Everything about one title with its community score, six category scores and the rating widget. Route: /games/hollow-meridian.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. A header with a large 144x208 cover, the title 'Hollow Meridian', a small 'Suggest an edit' link, the line 'Game, 2027' and a large score badge 8.4 out of 10. Left column: Synopsis (short paragraph, licence line), Details (developer, publisher, platforms, genres as links), Trailers (two 'Load trailer' placeholders with a consent note), Time to beat (Main 12 h, Extras 18 h, Completionist 31 h), Add to list. Right column (sticky): 'Community score' card with 8.4, the six category averages (Gameplay 8.9, Fun 8.6, Sound 8.2, Technical 7.4, Visuals 8.8, Storytelling 8.5), a 'See users' scores' button, and 'Your rating' with six half-star sliders (10 stars each, minus and plus buttons, value on the right) and a Save button disabled until all six are set.

States to show: Show the rating widget with three categories set (7.0, 9.0, 8.0) and three unset.

Visual language: colour palette paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Typography: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius). All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Item detail - Neon Noir Arcade

File: `docs/design-concepts/item-noir.png` (and `-m.png`)

```text
Design a concept screenshot of the Item detail page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Neon Noir Arcade" - a dark cinematic gaming dashboard: moody, high contrast, energetic but readable.

Page purpose: Everything about one title with its community score, six category scores and the rating widget. Route: /games/hollow-meridian.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. A header with a large 144x208 cover, the title 'Hollow Meridian', a small 'Suggest an edit' link, the line 'Game, 2027' and a large score badge 8.4 out of 10. Left column: Synopsis (short paragraph, licence line), Details (developer, publisher, platforms, genres as links), Trailers (two 'Load trailer' placeholders with a consent note), Time to beat (Main 12 h, Extras 18 h, Completionist 31 h), Add to list. Right column (sticky): 'Community score' card with 8.4, the six category averages (Gameplay 8.9, Fun 8.6, Sound 8.2, Technical 7.4, Visuals 8.8, Storytelling 8.5), a 'See users' scores' button, and 'Your rating' with six half-star sliders (10 stars each, minus and plus buttons, value on the right) and a Save button disabled until all six are set.

States to show: Show the rating widget with three categories set (7.0, 9.0, 8.0) and three unset.

Visual language: colour palette void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Typography: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the dark theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Item detail - Soft Paper Play

File: `docs/design-concepts/item-paper.png` (and `-m.png`)

```text
Design a concept screenshot of the Item detail page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Soft Paper Play" - a soft, minimal, playful product: approachable, light, a little whimsical, never childish.

Page purpose: Everything about one title with its community score, six category scores and the rating widget. Route: /games/hollow-meridian.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. A header with a large 144x208 cover, the title 'Hollow Meridian', a small 'Suggest an edit' link, the line 'Game, 2027' and a large score badge 8.4 out of 10. Left column: Synopsis (short paragraph, licence line), Details (developer, publisher, platforms, genres as links), Trailers (two 'Load trailer' placeholders with a consent note), Time to beat (Main 12 h, Extras 18 h, Completionist 31 h), Add to list. Right column (sticky): 'Community score' card with 8.4, the six category averages (Gameplay 8.9, Fun 8.6, Sound 8.2, Technical 7.4, Visuals 8.8, Storytelling 8.5), a 'See users' scores' button, and 'Your rating' with six half-star sliders (10 stars each, minus and plus buttons, value on the right) and a Save button disabled until all six are set.

States to show: Show the rating widget with three categories set (7.0, 9.0, 8.0) and three unset.

Visual language: colour palette cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Typography: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

### 4. Item detail without score (`item-no-score`)

Route: `/games/starfall-protocol`. The empty-data variant of the item page: a new title with fewer than 5 ratings and no cover.

#### Item detail without score - Editorial Press

File: `docs/design-concepts/item-no-score-editorial.png` (and `-m.png`)

```text
Design a concept screenshot of the Item detail without score page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Editorial Press" - a printed culture magazine or newspaper review section: calm, authoritative, text-forward.

Page purpose: The empty-data variant of the item page: a new title with fewer than 5 ratings and no cover. Route: /games/starfall-protocol.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Same structure as the item page but with a designed cover placeholder (the title initial on a tinted tile, not a blank grey box), a grey 'Not enough ratings' badge, 'The score appears after 5 ratings. 3 so far.', an empty Synopsis ('No synopsis yet.'), a short Details list, a 'Log in to rate' prompt in place of the sliders and a disabled 'See users' scores' button.

States to show: Make the empty states look intentional, friendly and consistent.

Visual language: colour palette paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Typography: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius). All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Item detail without score - Neon Noir Arcade

File: `docs/design-concepts/item-no-score-noir.png` (and `-m.png`)

```text
Design a concept screenshot of the Item detail without score page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Neon Noir Arcade" - a dark cinematic gaming dashboard: moody, high contrast, energetic but readable.

Page purpose: The empty-data variant of the item page: a new title with fewer than 5 ratings and no cover. Route: /games/starfall-protocol.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Same structure as the item page but with a designed cover placeholder (the title initial on a tinted tile, not a blank grey box), a grey 'Not enough ratings' badge, 'The score appears after 5 ratings. 3 so far.', an empty Synopsis ('No synopsis yet.'), a short Details list, a 'Log in to rate' prompt in place of the sliders and a disabled 'See users' scores' button.

States to show: Make the empty states look intentional, friendly and consistent.

Visual language: colour palette void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Typography: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the dark theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Item detail without score - Soft Paper Play

File: `docs/design-concepts/item-no-score-paper.png` (and `-m.png`)

```text
Design a concept screenshot of the Item detail without score page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Soft Paper Play" - a soft, minimal, playful product: approachable, light, a little whimsical, never childish.

Page purpose: The empty-data variant of the item page: a new title with fewer than 5 ratings and no cover. Route: /games/starfall-protocol.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Same structure as the item page but with a designed cover placeholder (the title initial on a tinted tile, not a blank grey box), a grey 'Not enough ratings' badge, 'The score appears after 5 ratings. 3 so far.', an empty Synopsis ('No synopsis yet.'), a short Details list, a 'Log in to rate' prompt in place of the sliders and a disabled 'See users' scores' button.

States to show: Make the empty states look intentional, friendly and consistent.

Visual language: colour palette cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Typography: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

### 5. Book edition (`edition`)

Route: `/books/a-quiet-engine/editions/4`. One edition of a book (ISBN, publisher, format) linking back to the work.

#### Book edition - Editorial Press

File: `docs/design-concepts/edition-editorial.png` (and `-m.png`)

```text
Design a concept screenshot of the Book edition page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Editorial Press" - a printed culture magazine or newspaper review section: calm, authoritative, text-forward.

Page purpose: One edition of a book (ISBN, publisher, format) linking back to the work. Route: /books/a-quiet-engine/editions/4.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'A Quiet Engine (Paperback, 2nd edition)', subtitle, 'Suggest an edit' link, a definition list (Pages 352, ISBN 978-0-00-000000-0, Language English, Publisher Lantern House, Year 2024, Format Paperback) and a link 'Back to the book'.

States to show: Add an edition cover thumbnail area (the current page has none) so the page is less sparse.

Visual language: colour palette paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Typography: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius). All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Book edition - Neon Noir Arcade

File: `docs/design-concepts/edition-noir.png` (and `-m.png`)

```text
Design a concept screenshot of the Book edition page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Neon Noir Arcade" - a dark cinematic gaming dashboard: moody, high contrast, energetic but readable.

Page purpose: One edition of a book (ISBN, publisher, format) linking back to the work. Route: /books/a-quiet-engine/editions/4.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'A Quiet Engine (Paperback, 2nd edition)', subtitle, 'Suggest an edit' link, a definition list (Pages 352, ISBN 978-0-00-000000-0, Language English, Publisher Lantern House, Year 2024, Format Paperback) and a link 'Back to the book'.

States to show: Add an edition cover thumbnail area (the current page has none) so the page is less sparse.

Visual language: colour palette void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Typography: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the dark theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Book edition - Soft Paper Play

File: `docs/design-concepts/edition-paper.png` (and `-m.png`)

```text
Design a concept screenshot of the Book edition page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Soft Paper Play" - a soft, minimal, playful product: approachable, light, a little whimsical, never childish.

Page purpose: One edition of a book (ISBN, publisher, format) linking back to the work. Route: /books/a-quiet-engine/editions/4.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'A Quiet Engine (Paperback, 2nd edition)', subtitle, 'Suggest an edit' link, a definition list (Pages 352, ISBN 978-0-00-000000-0, Language English, Publisher Lantern House, Year 2024, Format Paperback) and a link 'Back to the book'.

States to show: Add an edition cover thumbnail area (the current page has none) so the page is less sparse.

Visual language: colour palette cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Typography: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

### 6. Entity page (person, company, genre) (`entity`)

Route: `/people/ana-oliveira`. All titles linked to a person, company, franchise, platform, genre, game mode, country or language.

#### Entity page (person, company, genre) - Editorial Press

File: `docs/design-concepts/entity-editorial.png` (and `-m.png`)

```text
Design a concept screenshot of the Entity page (person, company, genre) page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Editorial Press" - a printed culture magazine or newspaper review section: calm, authoritative, text-forward.

Page purpose: All titles linked to a person, company, franchise, platform, genre, game mode, country or language. Route: /people/ana-oliveira.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Kind label 'Person', name 'Ana Oliveira', 'Suggest an edit' link, facts (Born 1981), a licensed description with a source link, role chips ('Director (3)', 'Writer (5)'), then sections per category (Movies, Series, Books) each titled with 'Showing 6 of 14' and 'average score 7.6' above rows (cover, title, year, rating count, score) and a 'See all' link.

States to show: One section has a 'Next' link; another shows only two rows.

Visual language: colour palette paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Typography: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius). All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Entity page (person, company, genre) - Neon Noir Arcade

File: `docs/design-concepts/entity-noir.png` (and `-m.png`)

```text
Design a concept screenshot of the Entity page (person, company, genre) page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Neon Noir Arcade" - a dark cinematic gaming dashboard: moody, high contrast, energetic but readable.

Page purpose: All titles linked to a person, company, franchise, platform, genre, game mode, country or language. Route: /people/ana-oliveira.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Kind label 'Person', name 'Ana Oliveira', 'Suggest an edit' link, facts (Born 1981), a licensed description with a source link, role chips ('Director (3)', 'Writer (5)'), then sections per category (Movies, Series, Books) each titled with 'Showing 6 of 14' and 'average score 7.6' above rows (cover, title, year, rating count, score) and a 'See all' link.

States to show: One section has a 'Next' link; another shows only two rows.

Visual language: colour palette void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Typography: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the dark theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Entity page (person, company, genre) - Soft Paper Play

File: `docs/design-concepts/entity-paper.png` (and `-m.png`)

```text
Design a concept screenshot of the Entity page (person, company, genre) page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Soft Paper Play" - a soft, minimal, playful product: approachable, light, a little whimsical, never childish.

Page purpose: All titles linked to a person, company, franchise, platform, genre, game mode, country or language. Route: /people/ana-oliveira.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Kind label 'Person', name 'Ana Oliveira', 'Suggest an edit' link, facts (Born 1981), a licensed description with a source link, role chips ('Director (3)', 'Writer (5)'), then sections per category (Movies, Series, Books) each titled with 'Showing 6 of 14' and 'average score 7.6' above rows (cover, title, year, rating count, score) and a 'See all' link.

States to show: One section has a 'Next' link; another shows only two rows.

Visual language: colour palette cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Typography: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

### 7. Search results (`search`)

Route: `/search?q=meridian`. Full-page search with type filter, grouped results and a did-you-mean hint.

#### Search results - Editorial Press

File: `docs/design-concepts/search-editorial.png` (and `-m.png`)

```text
Design a concept screenshot of the Search results page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Editorial Press" - a printed culture magazine or newspaper review section: calm, authoritative, text-forward.

Page purpose: Full-page search with type filter, grouped results and a did-you-mean hint. Route: /search?q=meridian.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell-free logged-out bar ('Log in' link and menu icon); the header search holds 'meridian' with its suggestion dropdown OPEN showing 5 suggestions (title, year, type). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'Results for meridian'; a search form; a type filter row (All, Game, Movie, TV series, Book; Game selected); grouped result lists with a cover thumbnail, the title with the matched word in bold and an alternate title in muted text; pagination Previous and Next; below, matching entity links (people, companies).

States to show: Show the suggestion popup open over the page, with keyboard focus on the second option.

Visual language: colour palette paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Typography: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius). All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Search results - Neon Noir Arcade

File: `docs/design-concepts/search-noir.png` (and `-m.png`)

```text
Design a concept screenshot of the Search results page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Neon Noir Arcade" - a dark cinematic gaming dashboard: moody, high contrast, energetic but readable.

Page purpose: Full-page search with type filter, grouped results and a did-you-mean hint. Route: /search?q=meridian.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell-free logged-out bar ('Log in' link and menu icon); the header search holds 'meridian' with its suggestion dropdown OPEN showing 5 suggestions (title, year, type). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'Results for meridian'; a search form; a type filter row (All, Game, Movie, TV series, Book; Game selected); grouped result lists with a cover thumbnail, the title with the matched word in bold and an alternate title in muted text; pagination Previous and Next; below, matching entity links (people, companies).

States to show: Show the suggestion popup open over the page, with keyboard focus on the second option.

Visual language: colour palette void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Typography: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the dark theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Search results - Soft Paper Play

File: `docs/design-concepts/search-paper.png` (and `-m.png`)

```text
Design a concept screenshot of the Search results page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Soft Paper Play" - a soft, minimal, playful product: approachable, light, a little whimsical, never childish.

Page purpose: Full-page search with type filter, grouped results and a did-you-mean hint. Route: /search?q=meridian.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell-free logged-out bar ('Log in' link and menu icon); the header search holds 'meridian' with its suggestion dropdown OPEN showing 5 suggestions (title, year, type). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'Results for meridian'; a search form; a type filter row (All, Game, Movie, TV series, Book; Game selected); grouped result lists with a cover thumbnail, the title with the matched word in bold and an alternate title in muted text; pagination Previous and Next; below, matching entity links (people, companies).

States to show: Show the suggestion popup open over the page, with keyboard focus on the second option.

Visual language: colour palette cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Typography: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

### 8. Search with no results (`search-empty`)

Route: `/search?q=zzqx`. The no-results state with a 'did you mean' suggestion.

#### Search with no results - Editorial Press

File: `docs/design-concepts/search-empty-editorial.png` (and `-m.png`)

```text
Design a concept screenshot of the Search with no results page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Editorial Press" - a printed culture magazine or newspaper review section: calm, authoritative, text-forward.

Page purpose: The no-results state with a 'did you mean' suggestion. Route: /search?q=zzqx.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'Results for zzqx'; search form; message 'No results for zzqx.'; a link 'Did you mean: zenith?'; an invitation to browse the four categories with four category tiles.

States to show: Design the empty state as a helpful moment built from layout and type, not from a drawing.

Visual language: colour palette paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Typography: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius). All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Search with no results - Neon Noir Arcade

File: `docs/design-concepts/search-empty-noir.png` (and `-m.png`)

```text
Design a concept screenshot of the Search with no results page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Neon Noir Arcade" - a dark cinematic gaming dashboard: moody, high contrast, energetic but readable.

Page purpose: The no-results state with a 'did you mean' suggestion. Route: /search?q=zzqx.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'Results for zzqx'; search form; message 'No results for zzqx.'; a link 'Did you mean: zenith?'; an invitation to browse the four categories with four category tiles.

States to show: Design the empty state as a helpful moment built from layout and type, not from a drawing.

Visual language: colour palette void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Typography: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the dark theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Search with no results - Soft Paper Play

File: `docs/design-concepts/search-empty-paper.png` (and `-m.png`)

```text
Design a concept screenshot of the Search with no results page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Soft Paper Play" - a soft, minimal, playful product: approachable, light, a little whimsical, never childish.

Page purpose: The no-results state with a 'did you mean' suggestion. Route: /search?q=zzqx.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'Results for zzqx'; search form; message 'No results for zzqx.'; a link 'Did you mean: zenith?'; an invitation to browse the four categories with four category tiles.

States to show: Design the empty state as a helpful moment built from layout and type, not from a drawing.

Visual language: colour palette cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Typography: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

### 9. Public profile (`user-profile`)

Route: `/u/jonas`. A person's public profile with their ratings, shelf counts and public lists.

#### Public profile - Editorial Press

File: `docs/design-concepts/user-profile-editorial.png` (and `-m.png`)

```text
Design a concept screenshot of the Public profile page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Editorial Press" - a printed culture magazine or newspaper review section: calm, authoritative, text-forward.

Page purpose: A person's public profile with their ratings, shelf counts and public lists. Route: /u/jonas.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon (logged in as maya, viewing another person's profile, with a 'Report' link). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Header with a round avatar (initial 'J'), nickname 'jonas', a summary line '128 ratings, 4 lists'; tabs Ratings, Status, Lists (Ratings active); a list of shelf cards (cover, title with year, 'Game', '8.5 stars of 10') as an infinite list.

States to show: Add a small second panel showing the Status tab content (per type counts: Want 12, In progress 3, Completed 41, Dropped 2).

Visual language: colour palette paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Typography: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius). All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Public profile - Neon Noir Arcade

File: `docs/design-concepts/user-profile-noir.png` (and `-m.png`)

```text
Design a concept screenshot of the Public profile page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Neon Noir Arcade" - a dark cinematic gaming dashboard: moody, high contrast, energetic but readable.

Page purpose: A person's public profile with their ratings, shelf counts and public lists. Route: /u/jonas.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon (logged in as maya, viewing another person's profile, with a 'Report' link). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Header with a round avatar (initial 'J'), nickname 'jonas', a summary line '128 ratings, 4 lists'; tabs Ratings, Status, Lists (Ratings active); a list of shelf cards (cover, title with year, 'Game', '8.5 stars of 10') as an infinite list.

States to show: Add a small second panel showing the Status tab content (per type counts: Want 12, In progress 3, Completed 41, Dropped 2).

Visual language: colour palette void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Typography: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the dark theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Public profile - Soft Paper Play

File: `docs/design-concepts/user-profile-paper.png` (and `-m.png`)

```text
Design a concept screenshot of the Public profile page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Soft Paper Play" - a soft, minimal, playful product: approachable, light, a little whimsical, never childish.

Page purpose: A person's public profile with their ratings, shelf counts and public lists. Route: /u/jonas.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon (logged in as maya, viewing another person's profile, with a 'Report' link). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Header with a round avatar (initial 'J'), nickname 'jonas', a summary line '128 ratings, 4 lists'; tabs Ratings, Status, Lists (Ratings active); a list of shelf cards (cover, title with year, 'Game', '8.5 stars of 10') as an infinite list.

States to show: Add a small second panel showing the Status tab content (per type counts: Want 12, In progress 3, Completed 41, Dropped 2).

Visual language: colour palette cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Typography: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

### 10. Public list (`user-list`)

Route: `/u/maya/lists/12`. A curated, ordered public list.

#### Public list - Editorial Press

File: `docs/design-concepts/user-list-editorial.png` (and `-m.png`)

```text
Design a concept screenshot of the Public list page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Editorial Press" - a printed culture magazine or newspaper review section: calm, authoritative, text-forward.

Page purpose: A curated, ordered public list. Route: /u/maya/lists/12.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon (logged in as maya, viewing her own list, with an 'Edit list' link). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'Cosy games for rainy days', byline 'by maya', '8 items', description, 'Edit list' link, then an ordered list of shelf cards (position number, cover, title, year, optional note in muted text).

States to show: Two entries carry notes.

Visual language: colour palette paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Typography: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius). All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Public list - Neon Noir Arcade

File: `docs/design-concepts/user-list-noir.png` (and `-m.png`)

```text
Design a concept screenshot of the Public list page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Neon Noir Arcade" - a dark cinematic gaming dashboard: moody, high contrast, energetic but readable.

Page purpose: A curated, ordered public list. Route: /u/maya/lists/12.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon (logged in as maya, viewing her own list, with an 'Edit list' link). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'Cosy games for rainy days', byline 'by maya', '8 items', description, 'Edit list' link, then an ordered list of shelf cards (position number, cover, title, year, optional note in muted text).

States to show: Two entries carry notes.

Visual language: colour palette void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Typography: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the dark theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Public list - Soft Paper Play

File: `docs/design-concepts/user-list-paper.png` (and `-m.png`)

```text
Design a concept screenshot of the Public list page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Soft Paper Play" - a soft, minimal, playful product: approachable, light, a little whimsical, never childish.

Page purpose: A curated, ordered public list. Route: /u/maya/lists/12.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon (logged in as maya, viewing her own list, with an 'Edit list' link). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'Cosy games for rainy days', byline 'by maya', '8 items', description, 'Edit list' link, then an ordered list of shelf cards (position number, cover, title, year, optional note in muted text).

States to show: Two entries carry notes.

Visual language: colour palette cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Typography: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

### 11. Library (`library`)

Route: `/library`. Personal shelves of everything the user wants, plays, finished or dropped.

#### Library - Editorial Press

File: `docs/design-concepts/library-editorial.png` (and `-m.png`)

```text
Design a concept screenshot of the Library page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Editorial Press" - a printed culture magazine or newspaper review section: calm, authoritative, text-forward.

Page purpose: Personal shelves of everything the user wants, plays, finished or dropped. Route: /library.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon (logged in as maya). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'My library'; a filter form (Status, Type, Sort and 'Apply'); sections Want to play (12), In progress (3), Completed (41), Dropped (2), each a three-column grid of shelf cards (cover, title, year, status, 'You: 8.5'), with one book card showing an edition picker and another showing 'Edition: Paperback 2nd ed. - Clear'; per-type 'Showing 6 of 20 - See all'.

States to show: Include the book edition picker open.

Visual language: colour palette paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Typography: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius). All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Library - Neon Noir Arcade

File: `docs/design-concepts/library-noir.png` (and `-m.png`)

```text
Design a concept screenshot of the Library page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Neon Noir Arcade" - a dark cinematic gaming dashboard: moody, high contrast, energetic but readable.

Page purpose: Personal shelves of everything the user wants, plays, finished or dropped. Route: /library.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon (logged in as maya). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'My library'; a filter form (Status, Type, Sort and 'Apply'); sections Want to play (12), In progress (3), Completed (41), Dropped (2), each a three-column grid of shelf cards (cover, title, year, status, 'You: 8.5'), with one book card showing an edition picker and another showing 'Edition: Paperback 2nd ed. - Clear'; per-type 'Showing 6 of 20 - See all'.

States to show: Include the book edition picker open.

Visual language: colour palette void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Typography: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the dark theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Library - Soft Paper Play

File: `docs/design-concepts/library-paper.png` (and `-m.png`)

```text
Design a concept screenshot of the Library page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Soft Paper Play" - a soft, minimal, playful product: approachable, light, a little whimsical, never childish.

Page purpose: Personal shelves of everything the user wants, plays, finished or dropped. Route: /library.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon (logged in as maya). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'My library'; a filter form (Status, Type, Sort and 'Apply'); sections Want to play (12), In progress (3), Completed (41), Dropped (2), each a three-column grid of shelf cards (cover, title, year, status, 'You: 8.5'), with one book card showing an edition picker and another showing 'Edition: Paperback 2nd ed. - Clear'; per-type 'Showing 6 of 20 - See all'.

States to show: Include the book edition picker open.

Visual language: colour palette cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Typography: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

### 12. My lists (`lists-index`)

Route: `/lists`. Manage the user's own lists and create a new one.

#### My lists - Editorial Press

File: `docs/design-concepts/lists-index-editorial.png` (and `-m.png`)

```text
Design a concept screenshot of the My lists page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Editorial Press" - a printed culture magazine or newspaper review section: calm, authoritative, text-forward.

Page purpose: Manage the user's own lists and create a new one. Route: /lists.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon (logged in as maya). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'My lists', a limit note 'Up to 10 lists per category', divided rows (title link, 'Game - 8 items - Public', and a red 'Hidden by a moderator' note on one row), and a 'New list' form (Type select, Title field, Public checkbox, Create button).

States to show: Show the form with a validation error under Title.

Visual language: colour palette paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Typography: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius). All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### My lists - Neon Noir Arcade

File: `docs/design-concepts/lists-index-noir.png` (and `-m.png`)

```text
Design a concept screenshot of the My lists page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Neon Noir Arcade" - a dark cinematic gaming dashboard: moody, high contrast, energetic but readable.

Page purpose: Manage the user's own lists and create a new one. Route: /lists.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon (logged in as maya). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'My lists', a limit note 'Up to 10 lists per category', divided rows (title link, 'Game - 8 items - Public', and a red 'Hidden by a moderator' note on one row), and a 'New list' form (Type select, Title field, Public checkbox, Create button).

States to show: Show the form with a validation error under Title.

Visual language: colour palette void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Typography: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the dark theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### My lists - Soft Paper Play

File: `docs/design-concepts/lists-index-paper.png` (and `-m.png`)

```text
Design a concept screenshot of the My lists page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Soft Paper Play" - a soft, minimal, playful product: approachable, light, a little whimsical, never childish.

Page purpose: Manage the user's own lists and create a new one. Route: /lists.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon (logged in as maya). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'My lists', a limit note 'Up to 10 lists per category', divided rows (title link, 'Game - 8 items - Public', and a red 'Hidden by a moderator' note on one row), and a 'New list' form (Type select, Title field, Public checkbox, Create button).

States to show: Show the form with a validation error under Title.

Visual language: colour palette cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Typography: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

### 13. Edit list (`lists-edit`)

Route: `/lists/12/edit`. Edit list details and reorder items.

#### Edit list - Editorial Press

File: `docs/design-concepts/lists-edit-editorial.png` (and `-m.png`)

```text
Design a concept screenshot of the Edit list page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Editorial Press" - a printed culture magazine or newspaper review section: calm, authoritative, text-forward.

Page purpose: Edit list details and reorder items. Route: /lists/12/edit.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon (logged in as maya). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. 'Back to my lists' link, title 'Edit list', details form (Title, Description textarea, Public checkbox, Save, Delete), 'Items' heading with an 'Add an item' search box, an ordered list where each row has position, small cover, title, 'Move up', 'Move down', 'Remove' buttons, and an enabled 'Save order' button.

States to show: Show one row mid-move with a visible keyboard focus ring on its 'Move down' button.

Visual language: colour palette paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Typography: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius). All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Edit list - Neon Noir Arcade

File: `docs/design-concepts/lists-edit-noir.png` (and `-m.png`)

```text
Design a concept screenshot of the Edit list page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Neon Noir Arcade" - a dark cinematic gaming dashboard: moody, high contrast, energetic but readable.

Page purpose: Edit list details and reorder items. Route: /lists/12/edit.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon (logged in as maya). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. 'Back to my lists' link, title 'Edit list', details form (Title, Description textarea, Public checkbox, Save, Delete), 'Items' heading with an 'Add an item' search box, an ordered list where each row has position, small cover, title, 'Move up', 'Move down', 'Remove' buttons, and an enabled 'Save order' button.

States to show: Show one row mid-move with a visible keyboard focus ring on its 'Move down' button.

Visual language: colour palette void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Typography: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the dark theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Edit list - Soft Paper Play

File: `docs/design-concepts/lists-edit-paper.png` (and `-m.png`)

```text
Design a concept screenshot of the Edit list page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Soft Paper Play" - a soft, minimal, playful product: approachable, light, a little whimsical, never childish.

Page purpose: Edit list details and reorder items. Route: /lists/12/edit.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon (logged in as maya). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. 'Back to my lists' link, title 'Edit list', details form (Title, Description textarea, Public checkbox, Save, Delete), 'Items' heading with an 'Add an item' search box, an ordered list where each row has position, small cover, title, 'Move up', 'Move down', 'Remove' buttons, and an enabled 'Save order' button.

States to show: Show one row mid-move with a visible keyboard focus ring on its 'Move down' button.

Visual language: colour palette cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Typography: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

### 14. Notifications (`notifications`)

Route: `/me/notifications`. In-app notifications such as moderation outcomes and replies.

#### Notifications - Editorial Press

File: `docs/design-concepts/notifications-editorial.png` (and `-m.png`)

```text
Design a concept screenshot of the Notifications page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Editorial Press" - a printed culture magazine or newspaper review section: calm, authoritative, text-forward.

Page purpose: In-app notifications such as moderation outcomes and replies. Route: /me/notifications.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon with a red badge '3' and an account icon (logged in as maya). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'Notifications' with a 'Mark all read' link; a list of 7 rows each with a sentence, a timestamp and a 'Mark read' button; unread rows are bold.

States to show: Add an 'All caught up' empty-state strip at the bottom as a secondary frame.

Visual language: colour palette paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Typography: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius). All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Notifications - Neon Noir Arcade

File: `docs/design-concepts/notifications-noir.png` (and `-m.png`)

```text
Design a concept screenshot of the Notifications page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Neon Noir Arcade" - a dark cinematic gaming dashboard: moody, high contrast, energetic but readable.

Page purpose: In-app notifications such as moderation outcomes and replies. Route: /me/notifications.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon with a red badge '3' and an account icon (logged in as maya). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'Notifications' with a 'Mark all read' link; a list of 7 rows each with a sentence, a timestamp and a 'Mark read' button; unread rows are bold.

States to show: Add an 'All caught up' empty-state strip at the bottom as a secondary frame.

Visual language: colour palette void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Typography: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the dark theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Notifications - Soft Paper Play

File: `docs/design-concepts/notifications-paper.png` (and `-m.png`)

```text
Design a concept screenshot of the Notifications page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Soft Paper Play" - a soft, minimal, playful product: approachable, light, a little whimsical, never childish.

Page purpose: In-app notifications such as moderation outcomes and replies. Route: /me/notifications.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon with a red badge '3' and an account icon (logged in as maya). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'Notifications' with a 'Mark all read' link; a list of 7 rows each with a sentence, a timestamp and a 'Mark read' button; unread rows are bold.

States to show: Add an 'All caught up' empty-state strip at the bottom as a secondary frame.

Visual language: colour palette cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Typography: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

### 15. Dashboard (`dashboard`)

Route: `/dashboard`. Hub right after login with shortcuts.

#### Dashboard - Editorial Press

File: `docs/design-concepts/dashboard-editorial.png` (and `-m.png`)

```text
Design a concept screenshot of the Dashboard page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Editorial Press" - a printed culture magazine or newspaper review section: calm, authoritative, text-forward.

Page purpose: Hub right after login with shortcuts. Route: /dashboard.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon (logged in as maya). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. 'Welcome, Maya', a lead sentence, a yellow 'Verify your email' prompt, a 2x2 grid of link cards (Library, Lists, My public page, API keys) each with a one-line help, and a link 'Set up two-factor authentication'.

States to show: Show the verify prompt.

Visual language: colour palette paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Typography: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius). All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Dashboard - Neon Noir Arcade

File: `docs/design-concepts/dashboard-noir.png` (and `-m.png`)

```text
Design a concept screenshot of the Dashboard page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Neon Noir Arcade" - a dark cinematic gaming dashboard: moody, high contrast, energetic but readable.

Page purpose: Hub right after login with shortcuts. Route: /dashboard.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon (logged in as maya). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. 'Welcome, Maya', a lead sentence, a yellow 'Verify your email' prompt, a 2x2 grid of link cards (Library, Lists, My public page, API keys) each with a one-line help, and a link 'Set up two-factor authentication'.

States to show: Show the verify prompt.

Visual language: colour palette void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Typography: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the dark theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Dashboard - Soft Paper Play

File: `docs/design-concepts/dashboard-paper.png` (and `-m.png`)

```text
Design a concept screenshot of the Dashboard page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Soft Paper Play" - a soft, minimal, playful product: approachable, light, a little whimsical, never childish.

Page purpose: Hub right after login with shortcuts. Route: /dashboard.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon (logged in as maya). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. 'Welcome, Maya', a lead sentence, a yellow 'Verify your email' prompt, a 2x2 grid of link cards (Library, Lists, My public page, API keys) each with a one-line help, and a link 'Set up two-factor authentication'.

States to show: Show the verify prompt.

Visual language: colour palette cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Typography: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

### 16. Account settings (`profile`)

Route: `/profile`. One long page of account sections.

#### Account settings - Editorial Press

File: `docs/design-concepts/profile-editorial.png` (and `-m.png`)

```text
Design a concept screenshot of the Account settings page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Editorial Press" - a printed culture magazine or newspaper review section: calm, authoritative, text-forward.

Page purpose: One long page of account sections. Route: /profile.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon with its menu OPEN (Library, Lists, Public profile, Profile, API keys, language links EN PT ES, Log out) (logged in as maya). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'Account', a sticky section index on the left on desktop (Details, Password, Passkeys, Two-factor, Sessions, Privacy, Export, Delete) and the sections on the right: Details form (name, nickname, language), Password form with a rule checklist and strength meter, Passkeys list (two passkeys with 'Remove'), Two-factor status, Sessions table (current session highlighted, 'Revoke'), Privacy toggles, 'Export my data' button, and a 'Delete account' danger zone with the 30-day grace note.

States to show: Show the password strength meter at 3 of 4 and the danger zone styled distinctly.

Visual language: colour palette paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Typography: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius). All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Account settings - Neon Noir Arcade

File: `docs/design-concepts/profile-noir.png` (and `-m.png`)

```text
Design a concept screenshot of the Account settings page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Neon Noir Arcade" - a dark cinematic gaming dashboard: moody, high contrast, energetic but readable.

Page purpose: One long page of account sections. Route: /profile.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon with its menu OPEN (Library, Lists, Public profile, Profile, API keys, language links EN PT ES, Log out) (logged in as maya). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'Account', a sticky section index on the left on desktop (Details, Password, Passkeys, Two-factor, Sessions, Privacy, Export, Delete) and the sections on the right: Details form (name, nickname, language), Password form with a rule checklist and strength meter, Passkeys list (two passkeys with 'Remove'), Two-factor status, Sessions table (current session highlighted, 'Revoke'), Privacy toggles, 'Export my data' button, and a 'Delete account' danger zone with the 30-day grace note.

States to show: Show the password strength meter at 3 of 4 and the danger zone styled distinctly.

Visual language: colour palette void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Typography: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the dark theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Account settings - Soft Paper Play

File: `docs/design-concepts/profile-paper.png` (and `-m.png`)

```text
Design a concept screenshot of the Account settings page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Soft Paper Play" - a soft, minimal, playful product: approachable, light, a little whimsical, never childish.

Page purpose: One long page of account sections. Route: /profile.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon with its menu OPEN (Library, Lists, Public profile, Profile, API keys, language links EN PT ES, Log out) (logged in as maya). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'Account', a sticky section index on the left on desktop (Details, Password, Passkeys, Two-factor, Sessions, Privacy, Export, Delete) and the sections on the right: Details form (name, nickname, language), Password form with a rule checklist and strength meter, Passkeys list (two passkeys with 'Remove'), Two-factor status, Sessions table (current session highlighted, 'Revoke'), Privacy toggles, 'Export my data' button, and a 'Delete account' danger zone with the 30-day grace note.

States to show: Show the password strength meter at 3 of 4 and the danger zone styled distinctly.

Visual language: colour palette cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Typography: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

### 17. API keys (`api-keys`)

Route: `/profile/api`. Create and revoke public API keys.

#### API keys - Editorial Press

File: `docs/design-concepts/api-keys-editorial.png` (and `-m.png`)

```text
Design a concept screenshot of the API keys page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Editorial Press" - a printed culture magazine or newspaper review section: calm, authoritative, text-forward.

Page purpose: Create and revoke public API keys. Route: /profile/api.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon (logged in as maya). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'API keys', intro with a docs link; a one-time 'Your new key' panel with the secret 'hsk_live_9f3a...' and Copy and Dismiss buttons plus a warning that it is shown once; a create form (Name field, Create key button, 'up to 5 keys'); a table (Name, Prefix, Created, Last used, Status, Actions) with three rows, one revoked and one with an inline 'Revoke this key? Cancel / Revoke' confirmation.

States to show: Both the new-key panel and the inline confirmation are visible.

Visual language: colour palette paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Typography: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius). All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### API keys - Neon Noir Arcade

File: `docs/design-concepts/api-keys-noir.png` (and `-m.png`)

```text
Design a concept screenshot of the API keys page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Neon Noir Arcade" - a dark cinematic gaming dashboard: moody, high contrast, energetic but readable.

Page purpose: Create and revoke public API keys. Route: /profile/api.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon (logged in as maya). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'API keys', intro with a docs link; a one-time 'Your new key' panel with the secret 'hsk_live_9f3a...' and Copy and Dismiss buttons plus a warning that it is shown once; a create form (Name field, Create key button, 'up to 5 keys'); a table (Name, Prefix, Created, Last used, Status, Actions) with three rows, one revoked and one with an inline 'Revoke this key? Cancel / Revoke' confirmation.

States to show: Both the new-key panel and the inline confirmation are visible.

Visual language: colour palette void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Typography: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the dark theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### API keys - Soft Paper Play

File: `docs/design-concepts/api-keys-paper.png` (and `-m.png`)

```text
Design a concept screenshot of the API keys page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Soft Paper Play" - a soft, minimal, playful product: approachable, light, a little whimsical, never childish.

Page purpose: Create and revoke public API keys. Route: /profile/api.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon (logged in as maya). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'API keys', intro with a docs link; a one-time 'Your new key' panel with the secret 'hsk_live_9f3a...' and Copy and Dismiss buttons plus a warning that it is shown once; a create form (Name field, Create key button, 'up to 5 keys'); a table (Name, Prefix, Created, Last used, Status, Actions) with three rows, one revoked and one with an inline 'Revoke this key? Cancel / Revoke' confirmation.

States to show: Both the new-key panel and the inline confirmation are visible.

Visual language: colour palette cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Typography: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

### 18. Billing (`billing`)

Route: `/billing`. Subscribe to or manage a paid API plan.

#### Billing - Editorial Press

File: `docs/design-concepts/billing-editorial.png` (and `-m.png`)

```text
Design a concept screenshot of the Billing page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Editorial Press" - a printed culture magazine or newspaper review section: calm, authoritative, text-forward.

Page purpose: Subscribe to or manage a paid API plan. Route: /billing.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon (logged in as maya). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'Billing', a success notice 'Thanks, your subscription is active', the current plan card (Plan: Pro, renews 2026-12-01, 'Manage billing' button), a currency selector (USD, EUR, BRL).

States to show: Show the subscribed state in the main frame and the unsubscribed state (plan purchase buttons) as a smaller second frame.

Visual language: colour palette paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Typography: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius). All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Billing - Neon Noir Arcade

File: `docs/design-concepts/billing-noir.png` (and `-m.png`)

```text
Design a concept screenshot of the Billing page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Neon Noir Arcade" - a dark cinematic gaming dashboard: moody, high contrast, energetic but readable.

Page purpose: Subscribe to or manage a paid API plan. Route: /billing.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon (logged in as maya). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'Billing', a success notice 'Thanks, your subscription is active', the current plan card (Plan: Pro, renews 2026-12-01, 'Manage billing' button), a currency selector (USD, EUR, BRL).

States to show: Show the subscribed state in the main frame and the unsubscribed state (plan purchase buttons) as a smaller second frame.

Visual language: colour palette void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Typography: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the dark theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Billing - Soft Paper Play

File: `docs/design-concepts/billing-paper.png` (and `-m.png`)

```text
Design a concept screenshot of the Billing page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Soft Paper Play" - a soft, minimal, playful product: approachable, light, a little whimsical, never childish.

Page purpose: Subscribe to or manage a paid API plan. Route: /billing.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon (logged in as maya). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'Billing', a success notice 'Thanks, your subscription is active', the current plan card (Plan: Pro, renews 2026-12-01, 'Manage billing' button), a currency selector (USD, EUR, BRL).

States to show: Show the subscribed state in the main frame and the unsubscribed state (plan purchase buttons) as a smaller second frame.

Visual language: colour palette cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Typography: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

### 19. Pricing (`pricing`)

Route: `/pricing`. Public pricing of the read-only API.

#### Pricing - Editorial Press

File: `docs/design-concepts/pricing-editorial.png` (and `-m.png`)

```text
Design a concept screenshot of the Pricing page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Editorial Press" - a printed culture magazine or newspaper review section: calm, authoritative, text-forward.

Page purpose: Public pricing of the read-only API. Route: /pricing.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'API pricing' with a subtitle, a currency switch (USD, EUR, BRL), three plan cards (Free $0, Pro $19/month, Scale $79/month) each with monthly quota, max page size, rate limit per second and a CTA ('Sign up', 'Subscribe'), the Pro card highlighted; a 'Coverage' section with a short paragraph and a link.

States to show: Pro is the recommended plan.

Visual language: colour palette paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Typography: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius). All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Pricing - Neon Noir Arcade

File: `docs/design-concepts/pricing-noir.png` (and `-m.png`)

```text
Design a concept screenshot of the Pricing page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Neon Noir Arcade" - a dark cinematic gaming dashboard: moody, high contrast, energetic but readable.

Page purpose: Public pricing of the read-only API. Route: /pricing.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'API pricing' with a subtitle, a currency switch (USD, EUR, BRL), three plan cards (Free $0, Pro $19/month, Scale $79/month) each with monthly quota, max page size, rate limit per second and a CTA ('Sign up', 'Subscribe'), the Pro card highlighted; a 'Coverage' section with a short paragraph and a link.

States to show: Pro is the recommended plan.

Visual language: colour palette void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Typography: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the dark theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Pricing - Soft Paper Play

File: `docs/design-concepts/pricing-paper.png` (and `-m.png`)

```text
Design a concept screenshot of the Pricing page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Soft Paper Play" - a soft, minimal, playful product: approachable, light, a little whimsical, never childish.

Page purpose: Public pricing of the read-only API. Route: /pricing.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'API pricing' with a subtitle, a currency switch (USD, EUR, BRL), three plan cards (Free $0, Pro $19/month, Scale $79/month) each with monthly quota, max page size, rate limit per second and a CTA ('Sign up', 'Subscribe'), the Pro card highlighted; a 'Coverage' section with a short paragraph and a link.

States to show: Pro is the recommended plan.

Visual language: colour palette cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Typography: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

### 20. Log in (`login`)

Route: `/login`. Sign in with email or nickname and password, or a passkey.

#### Log in - Editorial Press

File: `docs/design-concepts/login-editorial.png` (and `-m.png`)

```text
Design a concept screenshot of the Log in page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Editorial Press" - a printed culture magazine or newspaper review section: calm, authoritative, text-forward.

Page purpose: Sign in with email or nickname and password, or a passkey. Route: /login.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. A centred, width-constrained card with the title 'Log in', fields 'Email or nickname' and 'Password' (with show/hide eye), a 'Remember me' checkbox, a primary 'Log in' button, links 'Forgot your password?' and 'Create account', and an outlined 'Sign in with a passkey' button.

States to show: Show a field error under Password ('Incorrect password') and a 'Checking your browser' status line.

Visual language: colour palette paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Typography: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius). All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Log in - Neon Noir Arcade

File: `docs/design-concepts/login-noir.png` (and `-m.png`)

```text
Design a concept screenshot of the Log in page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Neon Noir Arcade" - a dark cinematic gaming dashboard: moody, high contrast, energetic but readable.

Page purpose: Sign in with email or nickname and password, or a passkey. Route: /login.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. A centred, width-constrained card with the title 'Log in', fields 'Email or nickname' and 'Password' (with show/hide eye), a 'Remember me' checkbox, a primary 'Log in' button, links 'Forgot your password?' and 'Create account', and an outlined 'Sign in with a passkey' button.

States to show: Show a field error under Password ('Incorrect password') and a 'Checking your browser' status line.

Visual language: colour palette void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Typography: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the dark theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Log in - Soft Paper Play

File: `docs/design-concepts/login-paper.png` (and `-m.png`)

```text
Design a concept screenshot of the Log in page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Soft Paper Play" - a soft, minimal, playful product: approachable, light, a little whimsical, never childish.

Page purpose: Sign in with email or nickname and password, or a passkey. Route: /login.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. A centred, width-constrained card with the title 'Log in', fields 'Email or nickname' and 'Password' (with show/hide eye), a 'Remember me' checkbox, a primary 'Log in' button, links 'Forgot your password?' and 'Create account', and an outlined 'Sign in with a passkey' button.

States to show: Show a field error under Password ('Incorrect password') and a 'Checking your browser' status line.

Visual language: colour palette cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Typography: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

### 21. Sign up (`signup`)

Route: `/signup`. Create an account.

#### Sign up - Editorial Press

File: `docs/design-concepts/signup-editorial.png` (and `-m.png`)

```text
Design a concept screenshot of the Sign up page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Editorial Press" - a printed culture magazine or newspaper review section: calm, authoritative, text-forward.

Page purpose: Create an account. Route: /signup.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Card with title 'Create your account', fields Name, Email, Password (with a live checklist: 12+ characters, a number, not a common password, and a 4-step strength meter), Confirm password, a primary button, and 'Have an account? Log in'.

States to show: Checklist shows two items satisfied and one pending.

Visual language: colour palette paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Typography: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius). All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Sign up - Neon Noir Arcade

File: `docs/design-concepts/signup-noir.png` (and `-m.png`)

```text
Design a concept screenshot of the Sign up page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Neon Noir Arcade" - a dark cinematic gaming dashboard: moody, high contrast, energetic but readable.

Page purpose: Create an account. Route: /signup.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Card with title 'Create your account', fields Name, Email, Password (with a live checklist: 12+ characters, a number, not a common password, and a 4-step strength meter), Confirm password, a primary button, and 'Have an account? Log in'.

States to show: Checklist shows two items satisfied and one pending.

Visual language: colour palette void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Typography: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the dark theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Sign up - Soft Paper Play

File: `docs/design-concepts/signup-paper.png` (and `-m.png`)

```text
Design a concept screenshot of the Sign up page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Soft Paper Play" - a soft, minimal, playful product: approachable, light, a little whimsical, never childish.

Page purpose: Create an account. Route: /signup.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Card with title 'Create your account', fields Name, Email, Password (with a live checklist: 12+ characters, a number, not a common password, and a 4-step strength meter), Confirm password, a primary button, and 'Have an account? Log in'.

States to show: Checklist shows two items satisfied and one pending.

Visual language: colour palette cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Typography: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

### 22. Check email and verify notices (`auth-notices`)

Route: `/signup/check-email and /verify-email`. Two small notice pages after sign-up: check your inbox, and verify your email.

#### Check email and verify notices - Editorial Press

File: `docs/design-concepts/auth-notices-editorial.png` (and `-m.png`)

```text
Design a concept screenshot of the Check email and verify notices page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Editorial Press" - a printed culture magazine or newspaper review section: calm, authoritative, text-forward.

Page purpose: Two small notice pages after sign-up: check your inbox, and verify your email. Route: /signup/check-email and /verify-email.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Two frames side by side on desktop (stacked on mobile): 'Check your email' with 'We sent a confirmation link to maya@example.test' and a 'Back to log in' link; 'Verify your email' with a 'Resend verification email' button and an envelope motif built only from UI shapes.

States to show: Both frames in the same screenshot.

Visual language: colour palette paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Typography: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius). All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Check email and verify notices - Neon Noir Arcade

File: `docs/design-concepts/auth-notices-noir.png` (and `-m.png`)

```text
Design a concept screenshot of the Check email and verify notices page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Neon Noir Arcade" - a dark cinematic gaming dashboard: moody, high contrast, energetic but readable.

Page purpose: Two small notice pages after sign-up: check your inbox, and verify your email. Route: /signup/check-email and /verify-email.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Two frames side by side on desktop (stacked on mobile): 'Check your email' with 'We sent a confirmation link to maya@example.test' and a 'Back to log in' link; 'Verify your email' with a 'Resend verification email' button and an envelope motif built only from UI shapes.

States to show: Both frames in the same screenshot.

Visual language: colour palette void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Typography: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the dark theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Check email and verify notices - Soft Paper Play

File: `docs/design-concepts/auth-notices-paper.png` (and `-m.png`)

```text
Design a concept screenshot of the Check email and verify notices page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Soft Paper Play" - a soft, minimal, playful product: approachable, light, a little whimsical, never childish.

Page purpose: Two small notice pages after sign-up: check your inbox, and verify your email. Route: /signup/check-email and /verify-email.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Two frames side by side on desktop (stacked on mobile): 'Check your email' with 'We sent a confirmation link to maya@example.test' and a 'Back to log in' link; 'Verify your email' with a 'Resend verification email' button and an envelope motif built only from UI shapes.

States to show: Both frames in the same screenshot.

Visual language: colour palette cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Typography: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

### 23. Onboarding (choose a nickname) (`onboarding`)

Route: `/onboarding`. Pick a public nickname after the first login.

#### Onboarding (choose a nickname) - Editorial Press

File: `docs/design-concepts/onboarding-editorial.png` (and `-m.png`)

```text
Design a concept screenshot of the Onboarding (choose a nickname) page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Editorial Press" - a printed culture magazine or newspaper review section: calm, authoritative, text-forward.

Page purpose: Pick a public nickname after the first login. Route: /onboarding.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon (logged in as maya). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Card with title 'Pick a nickname', a body line, a Nickname field with a hint (3 to 20 characters, letters, numbers, underscore), an error 'That nickname is taken', and a primary 'Continue' button.

States to show: Show the taken-nickname error.

Visual language: colour palette paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Typography: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius). All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Onboarding (choose a nickname) - Neon Noir Arcade

File: `docs/design-concepts/onboarding-noir.png` (and `-m.png`)

```text
Design a concept screenshot of the Onboarding (choose a nickname) page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Neon Noir Arcade" - a dark cinematic gaming dashboard: moody, high contrast, energetic but readable.

Page purpose: Pick a public nickname after the first login. Route: /onboarding.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon (logged in as maya). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Card with title 'Pick a nickname', a body line, a Nickname field with a hint (3 to 20 characters, letters, numbers, underscore), an error 'That nickname is taken', and a primary 'Continue' button.

States to show: Show the taken-nickname error.

Visual language: colour palette void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Typography: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the dark theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Onboarding (choose a nickname) - Soft Paper Play

File: `docs/design-concepts/onboarding-paper.png` (and `-m.png`)

```text
Design a concept screenshot of the Onboarding (choose a nickname) page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Soft Paper Play" - a soft, minimal, playful product: approachable, light, a little whimsical, never childish.

Page purpose: Pick a public nickname after the first login. Route: /onboarding.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon (logged in as maya). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Card with title 'Pick a nickname', a body line, a Nickname field with a hint (3 to 20 characters, letters, numbers, underscore), an error 'That nickname is taken', and a primary 'Continue' button.

States to show: Show the taken-nickname error.

Visual language: colour palette cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Typography: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

### 24. Recover and reset password (`recover`)

Route: `/recover-password and /reset-password/{token}`. Request a reset link and set a new password.

#### Recover and reset password - Editorial Press

File: `docs/design-concepts/recover-editorial.png` (and `-m.png`)

```text
Design a concept screenshot of the Recover and reset password page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Editorial Press" - a printed culture magazine or newspaper review section: calm, authoritative, text-forward.

Page purpose: Request a reset link and set a new password. Route: /recover-password and /reset-password/{token}.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Two frames: 'Forgot your password?' with an Email field and 'Send reset link'; 'Choose a new password' with Email (prefilled), New password with rules and meter, Confirm password and 'Reset password'.

States to show: Both frames in one screenshot.

Visual language: colour palette paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Typography: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius). All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Recover and reset password - Neon Noir Arcade

File: `docs/design-concepts/recover-noir.png` (and `-m.png`)

```text
Design a concept screenshot of the Recover and reset password page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Neon Noir Arcade" - a dark cinematic gaming dashboard: moody, high contrast, energetic but readable.

Page purpose: Request a reset link and set a new password. Route: /recover-password and /reset-password/{token}.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Two frames: 'Forgot your password?' with an Email field and 'Send reset link'; 'Choose a new password' with Email (prefilled), New password with rules and meter, Confirm password and 'Reset password'.

States to show: Both frames in one screenshot.

Visual language: colour palette void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Typography: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the dark theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Recover and reset password - Soft Paper Play

File: `docs/design-concepts/recover-paper.png` (and `-m.png`)

```text
Design a concept screenshot of the Recover and reset password page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Soft Paper Play" - a soft, minimal, playful product: approachable, light, a little whimsical, never childish.

Page purpose: Request a reset link and set a new password. Route: /recover-password and /reset-password/{token}.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Two frames: 'Forgot your password?' with an Email field and 'Send reset link'; 'Choose a new password' with Email (prefilled), New password with rules and meter, Confirm password and 'Reset password'.

States to show: Both frames in one screenshot.

Visual language: colour palette cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Typography: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

### 25. Two-factor challenge (`two-factor-challenge`)

Route: `/two-factor-challenge`. Enter the 6-digit authenticator code during login, or a backup code.

#### Two-factor challenge - Editorial Press

File: `docs/design-concepts/two-factor-challenge-editorial.png` (and `-m.png`)

```text
Design a concept screenshot of the Two-factor challenge page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Editorial Press" - a printed culture magazine or newspaper review section: calm, authoritative, text-forward.

Page purpose: Enter the 6-digit authenticator code during login, or a backup code. Route: /two-factor-challenge.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Card with title 'Two-factor authentication', help text, a 6-digit code field, a 'Verify' button, and a link 'Use a backup code instead'.

States to show: Show the 6-digit input as segmented boxes.

Visual language: colour palette paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Typography: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius). All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Two-factor challenge - Neon Noir Arcade

File: `docs/design-concepts/two-factor-challenge-noir.png` (and `-m.png`)

```text
Design a concept screenshot of the Two-factor challenge page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Neon Noir Arcade" - a dark cinematic gaming dashboard: moody, high contrast, energetic but readable.

Page purpose: Enter the 6-digit authenticator code during login, or a backup code. Route: /two-factor-challenge.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Card with title 'Two-factor authentication', help text, a 6-digit code field, a 'Verify' button, and a link 'Use a backup code instead'.

States to show: Show the 6-digit input as segmented boxes.

Visual language: colour palette void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Typography: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the dark theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Two-factor challenge - Soft Paper Play

File: `docs/design-concepts/two-factor-challenge-paper.png` (and `-m.png`)

```text
Design a concept screenshot of the Two-factor challenge page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Soft Paper Play" - a soft, minimal, playful product: approachable, light, a little whimsical, never childish.

Page purpose: Enter the 6-digit authenticator code during login, or a backup code. Route: /two-factor-challenge.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Card with title 'Two-factor authentication', help text, a 6-digit code field, a 'Verify' button, and a link 'Use a backup code instead'.

States to show: Show the 6-digit input as segmented boxes.

Visual language: colour palette cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Typography: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

### 26. Two-factor setup (`two-factor`)

Route: `/user/two-factor`. Enable, confirm and manage two-factor authentication.

#### Two-factor setup - Editorial Press

File: `docs/design-concepts/two-factor-editorial.png` (and `-m.png`)

```text
Design a concept screenshot of the Two-factor setup page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Editorial Press" - a printed culture magazine or newspaper review section: calm, authoritative, text-forward.

Page purpose: Enable, confirm and manage two-factor authentication. Route: /user/two-factor.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon (logged in as maya). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Three stacked states in one tall screenshot: off (explanation and 'Enable two-factor authentication'); pending (200px QR code placeholder, setup key 'JBSW Y3DP EHPK 3PXP', Code field, 'Confirm'); on (grid of 8 recovery codes, 'You have 8 recovery codes left', password fields to regenerate or disable, and a 'Turn off' danger section).

States to show: All three states in one screenshot.

Visual language: colour palette paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Typography: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius). All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Two-factor setup - Neon Noir Arcade

File: `docs/design-concepts/two-factor-noir.png` (and `-m.png`)

```text
Design a concept screenshot of the Two-factor setup page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Neon Noir Arcade" - a dark cinematic gaming dashboard: moody, high contrast, energetic but readable.

Page purpose: Enable, confirm and manage two-factor authentication. Route: /user/two-factor.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon (logged in as maya). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Three stacked states in one tall screenshot: off (explanation and 'Enable two-factor authentication'); pending (200px QR code placeholder, setup key 'JBSW Y3DP EHPK 3PXP', Code field, 'Confirm'); on (grid of 8 recovery codes, 'You have 8 recovery codes left', password fields to regenerate or disable, and a 'Turn off' danger section).

States to show: All three states in one screenshot.

Visual language: colour palette void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Typography: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the dark theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Two-factor setup - Soft Paper Play

File: `docs/design-concepts/two-factor-paper.png` (and `-m.png`)

```text
Design a concept screenshot of the Two-factor setup page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Soft Paper Play" - a soft, minimal, playful product: approachable, light, a little whimsical, never childish.

Page purpose: Enable, confirm and manage two-factor authentication. Route: /user/two-factor.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon (logged in as maya). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Three stacked states in one tall screenshot: off (explanation and 'Enable two-factor authentication'); pending (200px QR code placeholder, setup key 'JBSW Y3DP EHPK 3PXP', Code field, 'Confirm'); on (grid of 8 recovery codes, 'You have 8 recovery codes left', password fields to regenerate or disable, and a 'Turn off' danger section).

States to show: All three states in one screenshot.

Visual language: colour palette cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Typography: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

### 27. Suggest an edit (`suggest-edit`)

Route: `/suggest/item/1042`. Propose corrections to a title; drafts are kept in the browser.

#### Suggest an edit - Editorial Press

File: `docs/design-concepts/suggest-edit-editorial.png` (and `-m.png`)

```text
Design a concept screenshot of the Suggest an edit page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Editorial Press" - a printed culture magazine or newspaper review section: calm, authoritative, text-forward.

Page purpose: Propose corrections to a title; drafts are kept in the browser. Route: /suggest/item/1042.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon (logged in as maya). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'Suggest an edit', quota line '7 of 10 suggestions left today', 'Draft restored' note, an error list, editable field blocks (Title, Synopsis in English, Release date, External IDs) each showing the current value and a proposed value with a lock icon on one locked field, a language select, Sources textarea, Note textarea, and 'Save draft' and 'Submit' buttons.

States to show: One field has a validation error.

Visual language: colour palette paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Typography: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius). All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Suggest an edit - Neon Noir Arcade

File: `docs/design-concepts/suggest-edit-noir.png` (and `-m.png`)

```text
Design a concept screenshot of the Suggest an edit page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Neon Noir Arcade" - a dark cinematic gaming dashboard: moody, high contrast, energetic but readable.

Page purpose: Propose corrections to a title; drafts are kept in the browser. Route: /suggest/item/1042.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon (logged in as maya). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'Suggest an edit', quota line '7 of 10 suggestions left today', 'Draft restored' note, an error list, editable field blocks (Title, Synopsis in English, Release date, External IDs) each showing the current value and a proposed value with a lock icon on one locked field, a language select, Sources textarea, Note textarea, and 'Save draft' and 'Submit' buttons.

States to show: One field has a validation error.

Visual language: colour palette void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Typography: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the dark theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Suggest an edit - Soft Paper Play

File: `docs/design-concepts/suggest-edit-paper.png` (and `-m.png`)

```text
Design a concept screenshot of the Suggest an edit page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Soft Paper Play" - a soft, minimal, playful product: approachable, light, a little whimsical, never childish.

Page purpose: Propose corrections to a title; drafts are kept in the browser. Route: /suggest/item/1042.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a bell icon and an account icon (logged in as maya). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'Suggest an edit', quota line '7 of 10 suggestions left today', 'Draft restored' note, an error list, editable field blocks (Title, Synopsis in English, Release date, External IDs) each showing the current value and a proposed value with a lock icon on one locked field, a language select, Sources textarea, Note textarea, and 'Save draft' and 'Submit' buttons.

States to show: One field has a validation error.

Visual language: colour palette cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Typography: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

### 28. API documentation (`docs`)

Route: `/docs/pagination`. Public documentation of the read-only API.

#### API documentation - Editorial Press

File: `docs/design-concepts/docs-editorial.png` (and `-m.png`)

```text
Design a concept screenshot of the API documentation page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Editorial Press" - a printed culture magazine or newspaper review section: calm, authoritative, text-forward.

Page purpose: Public documentation of the read-only API. Route: /docs/pagination.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. A left column with the page list (Overview, Authentication, Errors, Pagination, Rate limits, Coverage, Changelog, API reference) and an on-page table of contents; a reading column up to 70 characters wide with a heading, paragraphs, an inline code span, a code block with a cURL example (curl -H 'Authorization: Bearer hsk_...' https://api.hexascore.app/v1/items?limit=20), a response JSON snippet and a callout.

States to show: Current page highlighted in the sidebar.

Visual language: colour palette paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Typography: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius). All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### API documentation - Neon Noir Arcade

File: `docs/design-concepts/docs-noir.png` (and `-m.png`)

```text
Design a concept screenshot of the API documentation page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Neon Noir Arcade" - a dark cinematic gaming dashboard: moody, high contrast, energetic but readable.

Page purpose: Public documentation of the read-only API. Route: /docs/pagination.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. A left column with the page list (Overview, Authentication, Errors, Pagination, Rate limits, Coverage, Changelog, API reference) and an on-page table of contents; a reading column up to 70 characters wide with a heading, paragraphs, an inline code span, a code block with a cURL example (curl -H 'Authorization: Bearer hsk_...' https://api.hexascore.app/v1/items?limit=20), a response JSON snippet and a callout.

States to show: Current page highlighted in the sidebar.

Visual language: colour palette void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Typography: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the dark theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### API documentation - Soft Paper Play

File: `docs/design-concepts/docs-paper.png` (and `-m.png`)

```text
Design a concept screenshot of the API documentation page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Soft Paper Play" - a soft, minimal, playful product: approachable, light, a little whimsical, never childish.

Page purpose: Public documentation of the read-only API. Route: /docs/pagination.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. A left column with the page list (Overview, Authentication, Errors, Pagination, Rate limits, Coverage, Changelog, API reference) and an on-page table of contents; a reading column up to 70 characters wide with a heading, paragraphs, an inline code span, a code block with a cURL example (curl -H 'Authorization: Bearer hsk_...' https://api.hexascore.app/v1/items?limit=20), a response JSON snippet and a callout.

States to show: Current page highlighted in the sidebar.

Visual language: colour palette cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Typography: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

### 29. Terms and privacy (`legal`)

Route: `/terms`. Long-form legal text.

#### Terms and privacy - Editorial Press

File: `docs/design-concepts/legal-editorial.png` (and `-m.png`)

```text
Design a concept screenshot of the Terms and privacy page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Editorial Press" - a printed culture magazine or newspaper review section: calm, authoritative, text-forward.

Page purpose: Long-form legal text. Route: /terms.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'Terms of service', 'Last updated 2026-10-01', a sticky section list, numbered sections with headings and paragraphs, a contact link at the end.

States to show: Optimise for reading comfort.

Visual language: colour palette paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Typography: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius). All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Terms and privacy - Neon Noir Arcade

File: `docs/design-concepts/legal-noir.png` (and `-m.png`)

```text
Design a concept screenshot of the Terms and privacy page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Neon Noir Arcade" - a dark cinematic gaming dashboard: moody, high contrast, energetic but readable.

Page purpose: Long-form legal text. Route: /terms.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'Terms of service', 'Last updated 2026-10-01', a sticky section list, numbered sections with headings and paragraphs, a contact link at the end.

States to show: Optimise for reading comfort.

Visual language: colour palette void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Typography: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the dark theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Terms and privacy - Soft Paper Play

File: `docs/design-concepts/legal-paper.png` (and `-m.png`)

```text
Design a concept screenshot of the Terms and privacy page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Soft Paper Play" - a soft, minimal, playful product: approachable, light, a little whimsical, never childish.

Page purpose: Long-form legal text. Route: /terms.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'Terms of service', 'Last updated 2026-10-01', a sticky section list, numbered sections with headings and paragraphs, a contact link at the end.

States to show: Optimise for reading comfort.

Visual language: colour palette cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Typography: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

### 30. How the score works (`help-score`)

Route: `/help/score`. Explains the six-category scoring and the 5-rating minimum.

#### How the score works - Editorial Press

File: `docs/design-concepts/help-score-editorial.png` (and `-m.png`)

```text
Design a concept screenshot of the How the score works page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Editorial Press" - a printed culture magazine or newspaper review section: calm, authoritative, text-forward.

Page purpose: Explains the six-category scoring and the 5-rating minimum. Route: /help/score.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'How scores work', an intro, a block per type listing its six categories with a one-line description each, a mini diagram built from UI shapes showing six category bars averaging into one 8.4 badge, and sections 'Minimum ratings', 'Unusual activity', 'Confidence'.

States to show: The diagram has a legend with all four score band colours.

Visual language: colour palette paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Typography: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius). All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### How the score works - Neon Noir Arcade

File: `docs/design-concepts/help-score-noir.png` (and `-m.png`)

```text
Design a concept screenshot of the How the score works page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Neon Noir Arcade" - a dark cinematic gaming dashboard: moody, high contrast, energetic but readable.

Page purpose: Explains the six-category scoring and the 5-rating minimum. Route: /help/score.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'How scores work', an intro, a block per type listing its six categories with a one-line description each, a mini diagram built from UI shapes showing six category bars averaging into one 8.4 badge, and sections 'Minimum ratings', 'Unusual activity', 'Confidence'.

States to show: The diagram has a legend with all four score band colours.

Visual language: colour palette void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Typography: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the dark theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### How the score works - Soft Paper Play

File: `docs/design-concepts/help-score-paper.png` (and `-m.png`)

```text
Design a concept screenshot of the How the score works page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Soft Paper Play" - a soft, minimal, playful product: approachable, light, a little whimsical, never childish.

Page purpose: Explains the six-category scoring and the 5-rating minimum. Route: /help/score.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'How scores work', an intro, a block per type listing its six categories with a one-line description each, a mini diagram built from UI shapes showing six category bars averaging into one 8.4 badge, and sections 'Minimum ratings', 'Unusual activity', 'Confidence'.

States to show: The diagram has a legend with all four score band colours.

Visual language: colour palette cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Typography: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

### 31. Delete account help (`help-delete`)

Route: `/help/delete-account`. Public page explaining account deletion (also used by app stores).

#### Delete account help - Editorial Press

File: `docs/design-concepts/help-delete-editorial.png` (and `-m.png`)

```text
Design a concept screenshot of the Delete account help page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Editorial Press" - a printed culture magazine or newspaper review section: calm, authoritative, text-forward.

Page purpose: Public page explaining account deletion (also used by app stores). Route: /help/delete-account.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'Delete your account', short paragraphs for the 30-day grace period, what is erased, what is kept (anonymised ratings), numbered steps, and two buttons 'Go to account settings' and 'Log in'.

States to show: Use a three-step visual timeline.

Visual language: colour palette paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Typography: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius). All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Delete account help - Neon Noir Arcade

File: `docs/design-concepts/help-delete-noir.png` (and `-m.png`)

```text
Design a concept screenshot of the Delete account help page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Neon Noir Arcade" - a dark cinematic gaming dashboard: moody, high contrast, energetic but readable.

Page purpose: Public page explaining account deletion (also used by app stores). Route: /help/delete-account.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'Delete your account', short paragraphs for the 30-day grace period, what is erased, what is kept (anonymised ratings), numbered steps, and two buttons 'Go to account settings' and 'Log in'.

States to show: Use a three-step visual timeline.

Visual language: colour palette void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Typography: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the dark theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Delete account help - Soft Paper Play

File: `docs/design-concepts/help-delete-paper.png` (and `-m.png`)

```text
Design a concept screenshot of the Delete account help page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Soft Paper Play" - a soft, minimal, playful product: approachable, light, a little whimsical, never childish.

Page purpose: Public page explaining account deletion (also used by app stores). Route: /help/delete-account.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'Delete your account', short paragraphs for the 30-day grace period, what is erased, what is kept (anonymised ratings), numbered steps, and two buttons 'Go to account settings' and 'Log in'.

States to show: Use a three-step visual timeline.

Visual language: colour palette cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Typography: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

### 32. Contact (`contact`)

Route: `/contact`. Public contact form.

#### Contact - Editorial Press

File: `docs/design-concepts/contact-editorial.png` (and `-m.png`)

```text
Design a concept screenshot of the Contact page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Editorial Press" - a printed culture magazine or newspaper review section: calm, authoritative, text-forward.

Page purpose: Public contact form. Route: /contact.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'Contact us', intro, Name, Email, Topic select, Message textarea with a hint, a 'Send message' button and, in a second small frame, the success status 'Thanks, we received your message.'

States to show: Success frame shown.

Visual language: colour palette paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Typography: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius). All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Contact - Neon Noir Arcade

File: `docs/design-concepts/contact-noir.png` (and `-m.png`)

```text
Design a concept screenshot of the Contact page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Neon Noir Arcade" - a dark cinematic gaming dashboard: moody, high contrast, energetic but readable.

Page purpose: Public contact form. Route: /contact.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'Contact us', intro, Name, Email, Topic select, Message textarea with a hint, a 'Send message' button and, in a second small frame, the success status 'Thanks, we received your message.'

States to show: Success frame shown.

Visual language: colour palette void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Typography: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the dark theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Contact - Soft Paper Play

File: `docs/design-concepts/contact-paper.png` (and `-m.png`)

```text
Design a concept screenshot of the Contact page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Soft Paper Play" - a soft, minimal, playful product: approachable, light, a little whimsical, never childish.

Page purpose: Public contact form. Route: /contact.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Title 'Contact us', intro, Name, Email, Topic select, Message textarea with a hint, a 'Send message' button and, in a second small frame, the success status 'Thanks, we received your message.'

States to show: Success frame shown.

Visual language: colour palette cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Typography: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

### 33. Error pages (404 and 500) (`error`)

Route: `/this-page-does-not-exist`. Friendly 404 and server error pages.

#### Error pages (404 and 500) - Editorial Press

File: `docs/design-concepts/error-editorial.png` (and `-m.png`)

```text
Design a concept screenshot of the Error pages (404 and 500) page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Editorial Press" - a printed culture magazine or newspaper review section: calm, authoritative, text-forward.

Page purpose: Friendly 404 and server error pages. Route: /this-page-does-not-exist.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Two frames: a 404 with the code '404', title 'Page not found', body text, a 'Back home' button and the search box; a 500 with '500', 'Something went wrong on our side' and a retry link.

States to show: Both frames in one screenshot.

Visual language: colour palette paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Typography: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius). All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Error pages (404 and 500) - Neon Noir Arcade

File: `docs/design-concepts/error-noir.png` (and `-m.png`)

```text
Design a concept screenshot of the Error pages (404 and 500) page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Neon Noir Arcade" - a dark cinematic gaming dashboard: moody, high contrast, energetic but readable.

Page purpose: Friendly 404 and server error pages. Route: /this-page-does-not-exist.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Two frames: a 404 with the code '404', title 'Page not found', body text, a 'Back home' button and the search box; a 500 with '500', 'Something went wrong on our side' and a retry link.

States to show: Both frames in one screenshot.

Visual language: colour palette void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Typography: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the dark theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Error pages (404 and 500) - Soft Paper Play

File: `docs/design-concepts/error-paper.png` (and `-m.png`)

```text
Design a concept screenshot of the Error pages (404 and 500) page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Soft Paper Play" - a soft, minimal, playful product: approachable, light, a little whimsical, never childish.

Page purpose: Friendly 404 and server error pages. Route: /this-page-does-not-exist.

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. Two frames: a 404 with the code '404', title 'Page not found', body text, a 'Back home' button and the search box; a 500 with '500', 'Something went wrong on our side' and a retry link.

States to show: Both frames in one screenshot.

Visual language: colour palette cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Typography: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

### 34. Cookie banner, toasts and dialog (global overlays) (`overlays`)

Route: `any page (shown over the Home page)`. Global overlays: cookie consent bar, success toast, error toast, confirm dialog.

#### Cookie banner, toasts and dialog (global overlays) - Editorial Press

File: `docs/design-concepts/overlays-editorial.png` (and `-m.png`)

```text
Design a concept screenshot of the Cookie banner, toasts and dialog (global overlays) page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Editorial Press" - a printed culture magazine or newspaper review section: calm, authoritative, text-forward.

Page purpose: Global overlays: cookie consent bar, success toast, error toast, confirm dialog. Route: any page (shown over the Home page).

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. The Home page dimmed behind, with a bottom cookie bar (text, 'Essential only' and 'Allow embeds' buttons), a success toast 'Rating saved' bottom right, an error toast 'Too many requests, try again in a minute' above it, and a centred modal 'Report this list' (reason select, details textarea, Cancel and Submit).

States to show: All four overlays visible at once.

Visual language: colour palette paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Typography: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius). All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Cookie banner, toasts and dialog (global overlays) - Neon Noir Arcade

File: `docs/design-concepts/overlays-noir.png` (and `-m.png`)

```text
Design a concept screenshot of the Cookie banner, toasts and dialog (global overlays) page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Neon Noir Arcade" - a dark cinematic gaming dashboard: moody, high contrast, energetic but readable.

Page purpose: Global overlays: cookie consent bar, success toast, error toast, confirm dialog. Route: any page (shown over the Home page).

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. The Home page dimmed behind, with a bottom cookie bar (text, 'Essential only' and 'Allow embeds' buttons), a success toast 'Rating saved' bottom right, an error toast 'Too many requests, try again in a minute' above it, and a centred modal 'Report this list' (reason select, details textarea, Cancel and Submit).

States to show: All four overlays visible at once.

Visual language: colour palette void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Typography: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the dark theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Cookie banner, toasts and dialog (global overlays) - Soft Paper Play

File: `docs/design-concepts/overlays-paper.png` (and `-m.png`)

```text
Design a concept screenshot of the Cookie banner, toasts and dialog (global overlays) page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Soft Paper Play" - a soft, minimal, playful product: approachable, light, a little whimsical, never childish.

Page purpose: Global overlays: cookie consent bar, success toast, error toast, confirm dialog. Route: any page (shown over the Home page).

Layout and content blocks: Top bar: wordmark 'HexaScore', links Games, Movies, TV Series, Books, a search field, a light/dark toggle icon, then a 'Log in' link and a menu icon (logged out). Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. The Home page dimmed behind, with a bottom cookie bar (text, 'Essential only' and 'Allow embeds' buttons), a success toast 'Rating saved' bottom right, an error toast 'Too many requests, try again in a minute' above it, and a centred modal 'Report this list' (reason select, details textarea, Cancel and Submit).

States to show: All four overlays visible at once.

Visual language: colour palette cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Typography: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

### 35. Admin panel (`admin`)

Route: `/admin/reports/31`. Internal moderation and catalogue panel (server-rendered, English only).

#### Admin panel - Editorial Press

File: `docs/design-concepts/admin-editorial.png` (and `-m.png`)

```text
Design a concept screenshot of the Admin panel page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Editorial Press" - a printed culture magazine or newspaper review section: calm, authoritative, text-forward.

Page purpose: Internal moderation and catalogue panel (server-rendered, English only). Route: /admin/reports/31.

Layout and content blocks: Header: an admin top bar with the title 'HexaScore Admin' and a 'Back to site' link instead of the public bar. Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. No public header. Top bar, a left sidebar (Dashboard, Items, Users, Reports, Scoring, Feature flags, Ingestion, Audit log) and a main area showing a report record: title 'Report 31', a detail grid (Target list 12, Reporter jonas, Reason Spam, Created 2026-10-30), a 'Resolve' form (Outcome select: No action, Dismissed, Hide the list, Ban the user; Note textarea; Resolve button) and a 'Back to reports' link; plus an index table of 5 report rows in a second frame.

States to show: Dense, data-first layout, in the style's palette.

Visual language: colour palette paper #FAF7F0, surface #F1ECE0, ink #16130F, muted ink #5C554A, hairline rule #D9D2C3, signal red #C8321E (links, primary button), deep teal #1F4E5F (secondary); category accents game #5B3FC4, movie #C8321E, series #1F6F8B, book #2E7D4F. Typography: a high-contrast transitional serif for headlines and scores (Playfair Display or Didot mood, large and tight), a neutral grotesque for UI text (Inter or Helvetica Neue mood), tabular monospaced numerals for counts. Imagery and components: item covers sit large with a thin ink border and a small caption line beneath (title, year); hairline rules, drop caps, kicker labels in small caps, generous whitespace, a 12-column grid, subtle paper grain, square corners (2px radius). All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Admin panel - Neon Noir Arcade

File: `docs/design-concepts/admin-noir.png` (and `-m.png`)

```text
Design a concept screenshot of the Admin panel page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Neon Noir Arcade" - a dark cinematic gaming dashboard: moody, high contrast, energetic but readable.

Page purpose: Internal moderation and catalogue panel (server-rendered, English only). Route: /admin/reports/31.

Layout and content blocks: Header: an admin top bar with the title 'HexaScore Admin' and a 'Back to site' link instead of the public bar. Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. No public header. Top bar, a left sidebar (Dashboard, Items, Users, Reports, Scoring, Feature flags, Ingestion, Audit log) and a main area showing a report record: title 'Report 31', a detail grid (Target list 12, Reporter jonas, Reason Spam, Created 2026-10-30), a 'Resolve' form (Outcome select: No action, Dismissed, Hide the list, Ban the user; Note textarea; Resolve button) and a 'Back to reports' link; plus an index table of 5 report rows in a second frame.

States to show: Dense, data-first layout, in the style's palette.

Visual language: colour palette void #0B0D12, surface #151923, raised surface #1C2230, border #262C3B, text #EDEFF5, muted #9AA3B8; glowing category accents game #8B7CFF, movie #FF5C6C, series #22C7EE, book #5BD38A; primary button gradient #8B7CFF to #5B8CFF with a soft outer glow. Typography: a condensed geometric display sans for headlines and big scores (Rajdhani or Oxanium mood, uppercase section labels with wide tracking), Inter for UI text, monospaced digits for stats. Imagery and components: item covers are tall posters with a soft coloured glow matching the category accent, a slight vignette and 10px radius; cards are dark glass panels with 1px inner borders; a thin neon underline marks active navigation; subtle scanline or noise texture at 3 percent. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the dark theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```

#### Admin panel - Soft Paper Play

File: `docs/design-concepts/admin-paper.png` (and `-m.png`)

```text
Design a concept screenshot of the Admin panel page of HexaScore (hexascore.app), a website where people rate games, movies, TV series and books. Style: "Soft Paper Play" - a soft, minimal, playful product: approachable, light, a little whimsical, never childish.

Page purpose: Internal moderation and catalogue panel (server-rendered, English only). Route: /admin/reports/31.

Layout and content blocks: Header: an admin top bar with the title 'HexaScore Admin' and a 'Back to site' link instead of the public bar. Content width about 1152px centred on the 1440px canvas, 16 to 24px side padding. No public header. Top bar, a left sidebar (Dashboard, Items, Users, Reports, Scoring, Feature flags, Ingestion, Audit log) and a main area showing a report record: title 'Report 31', a detail grid (Target list 12, Reporter jonas, Reason Spam, Created 2026-10-30), a 'Resolve' form (Outcome select: No action, Dismissed, Hide the list, Ban the user; Note textarea; Resolve button) and a 'Back to reports' link; plus an index table of 5 report rows in a second frame.

States to show: Dense, data-first layout, in the style's palette.

Visual language: colour palette cream #FFF9F2, card white #FFFFFF, ink #2B2A3A, muted #6E6B80, peach #FFD9C2, mint #CDEFE0, lilac #DDD3FF, sky #CFE8FF, primary violet #7A5CFA (buttons, links), danger coral #E5484D. Typography: a friendly rounded sans for everything (Nunito or Quicksand mood), bold rounded numerals for scores, comfortable 17px body size and generous line height. Imagery and components: item covers are rounded 16px cards with a soft pastel drop shadow and a tinted backdrop in the category colour (lilac game, peach movie, sky series, mint book); pill-shaped chips and buttons; playful sticker-like score badges; lots of space, almost no borders. All interactive controls are at least 44px tall and text meets WCAG AA contrast. Score display: out of 10 with one decimal; score badge bands: below 5.0 red #F05653, 5.0 to 7.4 amber #E6AD00, 7.5 and above green #2F9F3D, no score grey #9B9FA5 (shown as 'Not enough ratings'). Rating categories: six rating categories per type: game = Gameplay, Fun, Sound, Technical, Visuals, Storytelling; movie = Screenplay, Acting, Direction, Cinematography, Soundtrack, Entertainment; series = Screenplay, Acting, Characters, Pacing, Production, Entertainment; book = Plot, Characters, Writing, Pacing, Originality, Impact. Ratings run 0.5 to 10 in half-star steps. Use only invented titles: games 'Hollow Meridian' (8.4) and 'Starfall Protocol' (4.6), movie 'The Salt Orchard' (7.9), series 'Paper Lanterns' (6.8), book 'A Quiet Engine' (8.1). A dark-mode toggle is shown in the header; render this frame in the light theme.

Render it as a realistic web UI screenshot (a flat, front-on UI mockup of a real website at a 1440x900 desktop browser viewport, no device frame, no illustration, no 3D scene), crisp legible text with plausible English copy, consistent spacing. Do not use any real brand logos, trademarks or real game, film or book cover art; cover art must be abstract, invented artwork. Output: one image, 1440x900 landscape (a taller full-page capture is fine).

Mobile variant (generate as a second image): the same page at 390x844 portrait. The header becomes brand and icons on the first row, a full-width search field on the second row and a horizontally scrolling row of the four category links in their accent colours on the third; content stacks in one column, the sticky side panel moves above the content, tables become stacked cards, and a bottom cookie banner may overlap the last 120px.
```
