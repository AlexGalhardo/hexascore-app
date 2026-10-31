# FIX-passkeys-register-e2e passkeys.spec green in all projects
status: concluida
branch: agent/senior/FIX-passkeys-register-e2e @ cf15b8c
files: docker-compose.yml (APP_URL), qa/e2e/specs/passkeys.spec.ts
acceptance:
- passkeys.spec.ts on stack sa (:3100, E2E_BASE_URL=http://127.0.0.1:3100) -> 6 passed (6 projects)
root cause (two, no app or vukem bug): (1) the spec reused E2E_BASE_URL (127.0.0.1 by default and in CI); WebAuthn refuses an IP as RP ID, so create() fails in the browser right after /user/passkeys/options. (2) The app derives the RP origin from APP_URL (default http://localhost:3000, port included); any parallel stack on another port got POST /user/passkeys -> 422 "passkey ceremony rejected" (origin mismatch). CDP virtual authenticator and JSON decoding are fine.
fix: compose sets APP_URL=http://localhost:${QA_PORT:-3000} (default stack unchanged); spec rewrites //127.0.0.1 to //localhost.
validation: biome ok on the spec; mail helper reads link paths only, so the APP_URL port change is safe for other specs (full e2e not re-run)
changelog: Fixed: passkey e2e uses a localhost origin whose port matches the QA stack.
roadmap: no box
risks / follow-ups: CI e2e (qa-up.sh) uses port 3000, unchanged. Reviewer: confirm scripts/ci/qa-up.sh does not set its own APP_URL.

## Review (especialista, 2026-10-31)

verdict: APROVADO
checked: CI origin. `scripts/ci/qa-up.sh` sets no APP_URL and no QA_PORT, so the compose default gives `APP_URL=http://localhost:3000`; ci.yml e2e uses the default base `http://127.0.0.1:3000`, which the spec rewrites to `localhost:3000`: same origin. The smoke job (ci.yml l.221) already passes `APP_URL=http://localhost:3000`. Other specs only read mail link paths, so the APP_URL host change is safe; confirmed by the full e2e in the Release gate on stack esp (:3110, `APP_URL=http://localhost:3110`)
merged: 07c4129 on local `main`
PRONTO_PARA_MERGE agent/senior/FIX-passkeys-register-e2e
