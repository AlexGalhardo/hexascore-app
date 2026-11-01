# HexaScore app icon candidates

Three hand-written SVG candidates (owner round 2026-11-01, item 4). Colours come from the design tokens: primary `#5b3fc4`, dark-mode primary `#a796ff`, dark ledge `#5e4db8`, ink `#13161b`. No fonts, no third-party marks.

| Candidate | 512 px | 32 px | 16 px | Idea |
| --- | --- | --- | --- | --- |
| 1 (wired now) | ![Candidate 1](hexascore-icon-1-512.png) | ![Candidate 1, 32 px](hexascore-icon-1-32.png) | ![Candidate 1, 16 px](hexascore-icon-1-16.png) | Purple hexagon with a white slab-serif "H": the editorial heading face reduced to a mark |
| 2 | ![Candidate 2](hexascore-icon-2-512.png) | ![Candidate 2, 32 px](hexascore-icon-2-32.png) | ![Candidate 2, 16 px](hexascore-icon-2-16.png) | Purple tile with a white hexagon half filled, a score gauge |
| 3 | ![Candidate 3](hexascore-icon-3-512.png) | ![Candidate 3, 32 px](hexascore-icon-3-32.png) | ![Candidate 3, 16 px](hexascore-icon-3-16.png) | Ink tile with a hexagon of six wedges, five lit: six rating categories |

Files per candidate `N`: `hexascore-icon-N.svg` plus PNG exports at 16, 32, 180, 192 and 512 px (`hexascore-icon-N-<size>.png`).

## Switching the icon

The navbar (`APP_ICON` in `resources/js/components/Nav.tsx`) and the HTML head (`resources/views/app.html`: SVG favicon and `apple-touch-icon`) read only `app-icon.svg` and `app-icon-180.png`, which are copies of the chosen candidate. To pick candidate 2:

```sh
cd resources/public/icons
cp hexascore-icon-2.svg app-icon.svg
cp hexascore-icon-2-180.png app-icon-180.png
```

`/favicon.ico` is unchanged; browsers that support SVG favicons use `app-icon.svg` instead.

## Regenerating the PNGs

The exports were made with the `resvg` CLI 0.48.1 (not a project dependency):

```sh
for n in 1 2 3; do for s in 16 32 180 192 512; do
  resvg -w "$s" -h "$s" "hexascore-icon-$n.svg" "hexascore-icon-$n-$s.png"
done; done
```
