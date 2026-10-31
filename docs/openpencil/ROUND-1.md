# Round 1: three visual directions

Status 2026-10-31: v1 Editorial Press and v2 Neon Noir are delivered (19 frames each). v3 Soft Paper is pending (designer handoff, `.orchestration/handoff/designer.md`); its README already fixes palette and type.

Each version is one OpenPencil document with the pages `00 Cover`, `01 Foundations`, `02 Components`, `10 Web desktop 1440`, `11 Web mobile 390`, `20 Android 390x844` and `99 Assets` (real covers and shell components). PNGs are in each version's `frames/` folder, named `<platform>__<page-id>__<state>.png`.

## Side by side

| Frame | v1 Editorial Press | v2 Neon Noir | v3 Soft Paper |
| --- | --- | --- | --- |
| Cover | [cover](v1-editorial/frames/web__cover__default.png) | [cover](v2-noir/frames/web__cover__default.png) | pending |
| Foundations | [foundations](v1-editorial/frames/web__foundations__default.png) | [foundations](v2-noir/frames/web__foundations__default.png) | pending |
| Components | [components](v1-editorial/frames/web__components__default.png) | [components](v2-noir/frames/web__components__default.png) | pending |
| Home 1440 / 390 | [desktop](v1-editorial/frames/web-desktop__home__default.png) · [mobile](v1-editorial/frames/web-mobile__home__default.png) | [desktop](v2-noir/frames/web-desktop__home__default.png) · [mobile](v2-noir/frames/web-mobile__home__default.png) | pending |
| `/games` 1440 / 390 | [desktop](v1-editorial/frames/web-desktop__catalog-index__default.png) · [mobile](v1-editorial/frames/web-mobile__catalog-index__default.png) | [desktop](v2-noir/frames/web-desktop__catalog-index__default.png) · [mobile](v2-noir/frames/web-mobile__catalog-index__default.png) | pending |
| `/games/diablo-v` 1440 / 390 | [desktop](v1-editorial/frames/web-desktop__item__default.png) · [mobile](v1-editorial/frames/web-mobile__item__default.png) | [desktop](v2-noir/frames/web-desktop__item__default.png) · [mobile](v2-noir/frames/web-mobile__item__default.png) | pending |
| Search 1440 / 390 | [desktop](v1-editorial/frames/web-desktop__search__default.png) · [mobile](v1-editorial/frames/web-mobile__search__default.png) | [desktop](v2-noir/frames/web-desktop__search__default.png) · [mobile](v2-noir/frames/web-mobile__search__default.png) | pending |
| Library 1440 / 390 | [desktop](v1-editorial/frames/web-desktop__library__default.png) · [mobile](v1-editorial/frames/web-mobile__library__default.png) | [desktop](v2-noir/frames/web-desktop__library__default.png) · [mobile](v2-noir/frames/web-mobile__library__default.png) | pending |
| Login 1440 / 390 | [desktop](v1-editorial/frames/web-desktop__login__default.png) · [mobile](v1-editorial/frames/web-mobile__login__default.png) | [desktop](v2-noir/frames/web-desktop__login__default.png) · [mobile](v2-noir/frames/web-mobile__login__default.png) | pending |
| Pricing 1440 / 390 | [desktop](v1-editorial/frames/web-desktop__pricing__default.png) · [mobile](v1-editorial/frames/web-mobile__pricing__default.png) | [desktop](v2-noir/frames/web-desktop__pricing__default.png) · [mobile](v2-noir/frames/web-mobile__pricing__default.png) | pending |
| Android home / item | [home](v1-editorial/frames/android__home__default.png) · [item](v1-editorial/frames/android__item__default.png) | [home](v2-noir/frames/android__home__default.png) · [item](v2-noir/frames/android__item__default.png) | pending |
| Type | Playfair Display + Inter | Oxanium + Inter | Nunito |
| Default theme | light (paper) | dark (void) | light (cream) |
| Score mark | square band fill, ink number | ring in band colour with glow, number inside | pastel sticker chip, dark number |
| Density | highest (ranked tables, category columns) | medium (cover grids, panels) | lowest (planned) |

## v1 Editorial Press

Intent: a printed culture section, closest to the legacy "editorial data / almanac" principles. Strengths: the densest and most legible version; numbers lead (ranked lists, a `/games` table with all six category averages per row), hairline structure means almost no decoration to implement, and it maps cleanly to the current token set (light-first, 2 px radius, two font families). Risks: red is both primary and the movie accent (the primary button would need to switch to ink on movie pages); Playfair's old-style figures make the scores look uneven next to Inter (switch to lining figures or a different display face); it is the least "app-like" for the Android audience; covers are small, so weak cover art matters less but the page feels text-heavy.

## v2 Neon Noir

Intent: a dark, cinematic catalogue where covers and glowing score rings lead. Strengths: strongest visual identity and the most "product" feel on Android; cover-first grids make browsing fast; the score ring reads at any size and the band colour carries well on dark. Risks: dark-first is unusual for an SEO-heavy catalogue and needs a fully specified light twin (only tokens exist so far); gradients and glows add CSS weight and need care under the 30 kB CSS budget; placeholder covers (titles without art) look emptier in a cover grid than in a table; the uppercase condensed display face struggles with long pt-BR and es titles.

## v3 Soft Paper (pending)

Intent: approachable and warm, rounded cards with pastel section tints and sticker-like score chips, one family (Nunito), 17 px body. To be designed in the next run; see [v3-paper/README.md](v3-paper/README.md) for the fixed palette and type.

## Notes for every version

- Real content from production (read-only, 2026-10-31): titles, years, counts, covers, plans and prices. Production has no ratings yet, so every score number is an illustrative sample on the real 0.5 to 10 scale; the "–" state is real. Library and the "rated by" names are illustrative.
- Covers: all 23 titles use their real catalogue cover (96x144 JPEG downscales of the production thumbnails), applied by `scripts/openpencil/set_image_fills.py`.
- Scores use the production scale (out of 10, bands below 5.0, 5.0 to 7.4, 7.5 and above), not the 0 to 100 scale in the legacy notes and the designer brief.
- Production `/search` returned HTTP 500 for `q=diablo` and `q=witcher` while collecting content; the search frames use plausible counts.

## Decisões para o owner (round 1)

- [ ] Versão: v1 Editorial, v2 Noir, aguardar v3 Paper, ou mistura (diga quais partes de cada uma).
- [ ] Paleta: manter a da versão escolhida? v1: trocar o vermelho primário (conflita com o acento de filmes)? v2: tema escuro como padrão ou claro como padrão?
- [ ] Tipografia: v1 com Playfair (algarismos old-style) ou outra serifada com algarismos alinhados; v2 Oxanium em caixa alta ou só nos rótulos.
- [ ] Densidade: tabela densa com as seis categorias (v1) ou grade de capas (v2) no `/games` e na home.
- [ ] Nota: anel (v2), bloco quadrado (v1) ou selo pastel (v3).
