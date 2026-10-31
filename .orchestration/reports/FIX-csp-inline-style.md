# FIX-csp-inline-style drop inline style attributes
status: concluida
branch: agent/senior/FIX-csp-inline-style @ 1797815
files: resources/js/components/CoverThumb.tsx, resources/js/components/rating/StarRating.tsx, shell.test.tsx, StarRating.test.tsx
acceptance:
- csp.spec.ts on QA stack sa (:3100) -> 12 passed (6 projects x 2)
- lighthouse@12.6.1 best-practices (lighthouserc mobile settings) on the 5 URLs -> 1, 1, 1, 1, 1 (min 0.95)
validation: typecheck ok, biome ok, vitest components 26 passed
cause: React style prop on the CoverThumb placeholder (width/height) and the half-star clip (StarRating) is SSR-rendered as a style attribute, blocked by style-src-attr. Replaced by Tailwind classes (w-8 h-12 / w-11 h-16 / w-36 h-52; w-1/2 / w-full). CSP untouched.
changelog: Fixed: cover placeholder and half-star no longer emit inline style attributes (CSP violation, Lighthouse best-practices).
roadmap: no box
risks / follow-ups: mail template contact_message.html keeps an inline style (email, not under the CSP).

## Review (especialista, 2026-10-31)

verdict: APROVADO
re-run: `bun run test` → 21 files / 74 tests; `bun run lint` → 4 pre-existing warnings; no `style={` left in `resources/js` outside tests
checked: CSP header untouched (no `unsafe-inline`), classes are static literals so Tailwind emits them, new unit test asserts no `[style]`
merged: f3846ee on local `main`; csp.spec and Lighthouse in the Release gate
PRONTO_PARA_MERGE agent/senior/FIX-csp-inline-style
