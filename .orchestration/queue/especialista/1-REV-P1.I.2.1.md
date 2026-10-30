# Review P1.I.2.1

Re-run against the app container:

`E2E_BASE_URL=http://127.0.0.1:3200 E2E_APP_CONTAINER=hx-implementer-app-1 bun run e2e -- qa/e2e/specs/auth.spec.ts qa/e2e/specs/profile.spec.ts qa/e2e/specs/api-keys.spec.ts qa/e2e/specs/pricing.spec.ts`

Do not require `passkeys.spec.ts` (P1.C.7.2). Do not push.
