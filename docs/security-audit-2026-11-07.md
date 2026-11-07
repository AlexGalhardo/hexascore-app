# Security audit, 2026-11-07

Branch `feat/todo-2026-11`, commit `65f56203`. Method: the vendored `security-audit` skill (Cloudflare), adapted: one auditor, source-first, no delegated hunter waves. This is a **partial-coverage** run. No `findings.json` or `coverage.json` was produced because no record reached `confirmed`, and a ledger over a single-agent pass would overstate coverage.

## Summary

| Severity | Confirmed |
| --- | --- |
| Critical | 0 |
| High | 0 |
| Medium | 0 |
| Low | 0 |
| Informational | 0 |

No finding survived validation. Candidates that did not reach a demonstrable boundary violation are listed below.

## Reviewed

- Go SQL construction: every `fmt.Sprintf` query builds identifiers from fixed tables or validated specs; values are bound (`catalog/listing.go`, `redirects.go`, `jobs/prune_retention.go`).
- Raw HTML: the only `template.HTML` (`controllers/search_shell.go`) escapes every dynamic part.
- Outbound requests: webhooks (`app/webhooks`) and community cover fetch use `httpclient.Safe`, https only, public addresses only.
- Redirects: `locale_prefix.go` and `catalog_redirect.go` build same-site paths; `web.SafeRedirect` guards user-supplied targets.
- Render cache (`render_cache.go`): anonymous 200 only, no `Set-Cookie`, per-request CSP nonce swapped on replay, key covers locale/theme/consent/Accept-Language; canonical host is configured, not taken from `Host`.
- Authorization of suggestions (`policies/suggestion.go`, `community/draft.go`): author-only with 404-for-foreign; draft update locks on `author_id`.
- Proof-of-work challenge (`challenge/pow.go`): HMAC bound to route, expiry, single use via cache, fails closed.
- Secret comparisons use `subtle.ConstantTimeCompare` / `hmac.Equal`; no `math/rand` in security paths; no `InsecureSkipVerify`.
- Frontend (`resources/js`): no `dangerouslySetInnerHTML`, `innerHTML` or `eval`; user markdown links are restricted to absolute http(s) (`social/markdown.go`); `target="_blank"` sites carry `rel`.
- Workflows: every action pinned to a commit SHA, least-privilege `permissions`, no `pull_request_target`, no untrusted `${{ }}` in `run:`, agent review limited to same-repo PRs.
- Dockerfile: digest-pinned bases, non-root runtime user, `.dockerignore` excludes env files and `.git`.
- Mobile (`apps/mobile`): token in `expo-secure-store` with `AFTER_FIRST_UNLOCK_THIS_DEVICE_ONLY`; Android App Links are verified https hosts; no WebView; no cleartext config found.

## Dropped candidates (not validated)

1. `internal/ingest/covers/dhash.go` `Fingerprint` decodes a PNG/JPEG without the `MaxPixels` cap that `Cwebp.Encode` enforces afterwards. The only caller feeds bodies fetched from configured ingest sources, not user input, so no lower-trust principal reaches it. Hardening only: add a `DecodeConfig` pixel check before the decode.
2. Community cover upload (`community/uploads.go`) has a 300 px minimum but no maximum dimension; `app/seo/og_image.go` decodes stored covers with `image.Decode`. A user cover must first pass moderator approval and I could not confirm it reaches `og_image` unre-encoded, so the path is unproven. Worth a second look: add a maximum dimension in `measureCover`.
3. Mobile Google sign-in (`auth/googlePrompt.ts`) uses the implicit ID-token flow over the custom `hexascore://` scheme with a `Math.random` nonce. Custom schemes can be claimed by another Android app. ID-token verification (audience, nonce) lives in the vukem framework, outside this repo, so the impact cannot be shown from source. Needs validation by the owner against the framework's Google verifier.
4. `Linking.openURL(sourceUrl)` in `screens/Entity.tsx`: the URL comes from ingest-written Wikipedia sources, not from users; no lower-trust principal.

## Not covered

Billing/Stripe handlers, API-key and app-token issuance and expiry, 2FA and passkey flows, exports and signed backup links, admin React resources, SSR supervisor, `og_image` subprocess handling, OpenAPI surface, dependency audit, and the Go integration paths were only sampled by grep or not read. A deeper pass (multi-agent, `deep` profile) should start there.
