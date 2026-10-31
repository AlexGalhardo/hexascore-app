# FIX-app-client-regen regenerate packages/app-client/src/api.ts
status: concluida
branch: agent/senior/FIX-app-client-regen @ 7dfc80c
files: packages/app-client/src/api.ts
acceptance:
- docker run hexascore-app:sa codegen:client -prefix /app/v1 -o - (APP_ENV=development, APP_KEY, DATABASE_URL, as ci.yml) -> exit 0; diff: GetItemParams.type and ListItemsParams.type now "games" | "movies" | "series" | "books"
- bun run --filter @hexascore/app-client typecheck test -> see validation
validation: app-client typecheck and tests ok (output tail checked)
changelog: Fixed: app client regenerated for the item type enum.
roadmap: no box
risks / follow-ups: image built from the FIX-csp-inline-style branch (differs from main only in two TSX components, no API surface).

## Review (especialista, 2026-10-31)

verdict: APROVADO
re-run: `bun run --filter @hexascore/app-client typecheck test` on main 7a5bf01 → 2 pass; CI codegen diff from the image built on main (07c4129, Go and assets identical to 7a5bf01) in the Release gate `openapi` job
merged: 7a5bf01 on local `main`
PRONTO_PARA_MERGE agent/senior/FIX-app-client-regen
