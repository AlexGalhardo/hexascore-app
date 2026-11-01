# IMPL-mobile-v1-ui

Branch `agent/senior/IMPL-mobile-v1-ui` (stacked on `agent/senior/IMPL-mobile-v1-tokens` @ 33ece5f). Worker: impl-mobile.

## What (all in `apps/mobile/src/ui/`)

- `Btn.tsx`: primary / secondary / ghost from round-2 item 12. `min-h-11` (44 px), 2 px ledge (`ledge-primary`, `ledge-secondary`) at rest, `active:` sinks 1 px onto a 1 px ledge; with Android "Remove animations" on (`AccessibilityInfo.isReduceMotionEnabled`, `motion.ts`) only the ledge changes. `accessibilityState.disabled`.
- `ScoreBadge.tsx`: band fill from `scoreBand`, number always printed (one decimal in the locale, "–" when null), ink `on-score` in both themes, TalkBack label passed by the caller (translated).
- `RatingInput.tsx`: round-2 item 6 on Android. Ten stars, half steps (`valueAt` maps the touch point to 0.5 steps), value printed beside ("8,5 / 10" in pt-BR), 44 px row. Accessibility: one `accessibilityRole="adjustable"` control with `accessibilityValue {min 0, max 10, now, text}` and increment/decrement actions (TalkBack swipe up/down = ±0.5, clamped 0.5..10); the star row is hidden from TalkBack. This is the native equivalent of the web radio group of 20.
- `Cover.tsx`: reserved 2:3 box (`aspect-[2/3]`) with `expo-image`, labelled with the title, surface fill when there is no image.
- `ThemeToggle.tsx`: moon in light, sun in dark (item 8), 44 px, label passed by the caller.
- `format.ts`: `formatScore` (Intl, one decimal, "–").

No strings are hard-coded except the glyphs; labels come from callers, which read the catalogues.

## Validation

- Tests first: `src/ui/ui.test.tsx`, `src/ui/more.test.tsx` (14 tests).
- `bun run --cwd apps/mobile lint`: clean. `typecheck`: clean. `test`: 18 suites green (125 + 4).
- No device run (owner-gated); NativeWind classes used (`active:`, `aspect-[2/3]`, `tabular-nums`, `translate-y-px`) are NativeWind v4 supported utilities, to confirm on device.

Changelog line: `Added: Android v1 primitives (button with depth, score badge, 2:3 cover, theme toggle) and an accessible ten-star half-step rating input.`

## Review (especialista, 2026-10-31)

verdict: CORRIGIDO_E_APROVADO
re-run: mobile `lint` ok, `typecheck` ok, `test` 18 suites / 129 passed
checked: round-2 items 6, 8, 12; a11y (one adjustable control with min/max/now/text and increment/decrement, 44 px targets, labels from callers, reduced motion), tokens only, no hard-coded strings except glyphs, scope, tests
fixes: 1. half-star overlay: the filled glyph box was as wide as the clipped half, so the half star rendered centred at 25 % of the cell instead of over its outline; the glyph box is now `w-[200%]` for halves (2010593)
notes: not checked on a device (owner-gated); NativeWind `active:`, `aspect-[2/3]`, `w-[200%]`, `translate-y-px` to confirm on device. The filled light star (`star` on `bg`) is below 3:1 as a graphic; the value is always printed beside, so colour is not the only signal.
gate on main after merge (c2f614f): mobile tsc ok, jest 129 passed
PRONTO_PARA_MERGE agent/especialista/IMPL-mobile-v1-ui-fix
