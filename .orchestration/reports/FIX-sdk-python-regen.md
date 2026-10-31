# FIX-sdk-python-regen Regenerate the Python SDK
status: concluida
branch: agent/senior/FIX-sdk-python-regen @ 5b2ff0d
files: sdks/python/hexascore/ (generated: webhooks, changes, type enums; 4 modified, 16 new)
acceptance:
- generate (openapi-python-client==0.29.1 as in the `sdk:generate:python` script) inside `ghcr.io/astral-sh/uv:0.9.30-python3.13-bookworm-slim` → exit 0, only `sdks/python/hexascore` changed
- `uv run pytest -q` in `sdks/python` (same container) → 2 passed
validation: lint n/a (generated), tests python: ok
changelog: Fixed: regenerate the Python SDK so it covers webhooks and the changes feed.
roadmap: no box
risks / follow-ups: setups/setup-windows.ps1 pins no uv version; the container tag is the pin used here. The host has uv on PATH only through a stale winget entry (not installed). FIX-sdk-ts-generator is blocked: see Q-A8 in QUESTIONS.md.

## Review (especialista, 2026-10-31)
verdict: APROVADO
re-run: not regenerated (generated code; no host uv); checked instead that the 22 operations of `openapi/api-v1.json` match the generated modules (new: listChanges, list/create/delete/testWebhook; `type` enums GetItemType/ListItemsType from P2.H.3.1) and that `openapi/` has no drift between the branch base and main
checked: only `sdks/python/hexascore/` changed, generator pin 0.29.1 = `sdk:generate:python`, no secrets
gate: covered by the main gate (no Go/TS change)
PRONTO_PARA_MERGE agent/senior/FIX-sdk-python-regen (merged 0998e70); P2.H.3.1 box ticked (TS SDK regen tracked by FIX-sdk-ts-generator, Q-A8)
