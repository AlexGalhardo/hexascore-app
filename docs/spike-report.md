# Phase 0 framework spike report

Findings of the Phase 0 spike (ROADMAP "Phase 0: framework spike"). Each microtask appends a section with the commands run, the decisive output and a verdict (works / workaround / blocked); P0.F.11.1 consolidates them into one table.

## P0.F.1.2 Better Auth hash format

Verdict: works (format pinned, fixtures produced, cross-verified outside Node).

Source read: `better-auth` 1.7.7 (the version `legacy/apps/web/package.json:32` pins) installed with `bun install --ignore-scripts` in a temporary folder outside `legacy/`. Better Auth delegates hashing to its dependency `@better-auth/utils` 0.4.2 (`better-auth/package.json:466`); `better-auth/dist/crypto/password.mjs:1-13` re-exports `hashPassword` / `verifyPassword` from `@better-auth/utils/password`, which resolves to `node:crypto` scrypt on Node and Bun and to `@noble/hashes` scrypt elsewhere. The legacy app does not override the hasher (`legacy/apps/web/lib/auth.ts` has no `password.hash` / `password.verify` option).

Parameters (`@better-auth/utils/dist/password.node.mjs:3-34`):

| Parameter | Value | Line |
| --- | --- | --- |
| Algorithm | scrypt (`node:crypto`) | 1, 11 |
| N | 16384 | 4 |
| r | 16 | 5 |
| p | 1 | 6 |
| dkLen | 64 bytes | 7 |
| maxmem | `128 * N * r * 2` (64 MiB) | 19 |
| Salt | 16 random bytes, hex-encoded to a 32-character string | 31 |
| Password input | `password.normalize("NFKC")` | 12 |
| Stored format | `<salt hex>:<derived key hex>` (32 + 1 + 128 characters) | 33 |

Pitfall for the Go verifier: the scrypt salt is the 32-character hex **text** (its ASCII bytes), not the 16 decoded bytes. `verifyPassword` splits on `:` and passes the hex string straight to scrypt (lines 35-41). Better Auth compares the hex strings with `===`; the Go verifier must use a constant-time compare instead.

Fixtures: `docs/spike/testdata/betterauth_hashes.json` (to be copied into the spike as `testdata/betterauth_hashes.json` by P0.F.1.3), generated from throwaway passwords with `hashPassword` and self-checked with `verifyPassword` (correct password verifies, password + `x` fails):

- `ascii`: 28-character ASCII password.
- `accented-decomposed-nfkc`: decomposed accents (`e` + U+0301, `c` + U+0327, `a` + U+0303) and the ligature U+FB01; its NFKC-composed equivalent (`nfkc_equivalent`) also verifies.
- `32-chars`: 32-character password with symbols.

Independent check with Python `hashlib.scrypt` (stdlib) over the same fixtures: salt as hex text matches 3/3, salt as decoded bytes matches 0/3, and skipping NFKC breaks the accented case only. This is the behaviour the compatibility hasher in P0.F.1.3 must reproduce.

2FA secrets (open issue 03-2, export not done here): the TOTP plugin decrypts `twoFactor.secret` with `symmetricDecrypt({ key: ctx.context.secretConfig, data })` (`better-auth/dist/plugins/two-factor/totp/index.mjs:122-125`, defined in `better-auth/dist/crypto/index.mjs:48`); backup codes use the same function when `storeBackupCodes` is `"encrypted"` (`two-factor/backup-codes/index.mjs:45`). The key is the Better Auth secret configuration, so the export script needs `BETTER_AUTH_SECRET` at export time.

Acceptance:

- `rg -n "scrypt|NFKC" <copy>/node_modules/better-auth/dist <copy>/node_modules/@better-auth` → `@better-auth/utils/dist/password.node.mjs:1` (`import { randomBytes, scrypt } from 'node:crypto'`), `:12` (`password.normalize("NFKC")`); constants quoted above with lines.
- `jq '.cases | length' docs/spike/testdata/betterauth_hashes.json` → `3`.
