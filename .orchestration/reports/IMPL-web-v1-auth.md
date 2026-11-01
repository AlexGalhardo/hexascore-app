# IMPL-web-v1-auth Login, signup and recovery in design v1

status: concluida
branch: agent/senior/IMPL-web-v1-auth @ 5246eee (stacked on agent/senior/IMPL-web-v1-home @ 2689965)
files: resources/js/components/auth/AuthFrame.tsx (new: AuthFrame, OrRule, GoogleButton, authLink), resources/js/components/{Field,PasswordField,SubmitButton}.tsx, resources/js/pages/Auth/{Login,Signup,Recover,Reset}.tsx, resources/js/pages/Auth/Signup.test.tsx (new), resources/lang/{en,pt-BR,es}.json (additive `auth.*` keys), docs/sdd/09-web.md (§7.4)
acceptance:
- `bunx vitest run resources/js/pages/Auth resources/js/components` → 45/45 (new: agreement links Terms/Privacy through the locale helper; Google button is a plain link to `/auth/google/redirect`)
- `bun run lint` → 185 files, no fixes; `bun run typecheck` → exit 0; `bun run test` → 34 files / 127 tests passed
- Browser (QA :3140): `.orchestration/reports/shots/impl-web/{desktop,mobile}-{light,dark}-_{login,signup,recover_password}.png` (full page). Compared with `web-{desktop,mobile}{,-dark}__{login,signup}__default.png`: eyebrow "Members", 72 px serif heading and lede on the left, bordered card on the right (mobile: no box, a top rule and the card title kept as `sr-only` h1, as in the mobile frame), Google button in Google's colours with the four-colour G and a 2 px ledge in its border colour, "or" rule, v1 fields (field border, field-bg, 48 px), "Show" text toggle in primary, remember me + forgot on one row, full-width primary submit, secondary passkey button, "New here? Create account".
- SDD 09 §7.4 now describes the rating widget as a radio group of 20 native radios (no slider, no PageUp/PageDown), matching `rating/StarRating.tsx`.
validation: lint ok, typecheck ok, vitest ok, browser ok (light/dark x 1440/390)
changelog: Changed: login, signup, password recovery and reset follow design v1, with a "Sign in with Google" / "Sign up with Google" button in Google's branding.
roadmap: no box
risks / follow-ups:
- Deviations from the signup frame kept on purpose, because the backend contract (`app/requests/auth.go`) requires them: the first field is `name` ("Name", the frame says "Nickname"; the nickname is chosen on Onboarding), "Confirm password" stays, and there is no terms checkbox: the existing unused `auth.signup.agree` line ("By creating an account you agree to the Terms and the Privacy policy", links localized) is shown above the submit instead. Dropping the confirmation field or adding an explicit terms checkbox is a backend change; not opened as a row, the owner can ask for it.
- Google colours are Tailwind arbitrary values (`#FFFFFF`/`#747775`/`#1F1F1F`, dark `#131314`/`#8E918F`/`#E3E3E3`), as in impl-mobile's `GoogleButton`: they are Google's, not design tokens (README round 2, item 15). Inter replaces Roboto (two-family budget), 44 px with the v1 ledge: the two deviations the README asks to confirm.
- The Google button points at `/auth/google/redirect` without the locale prefix (the route is registered once, like the challenge page link). When Google is not configured locally the redirect errors; production has it live.
- Field, PasswordField and SubmitButton are shared: Contact, API keys, Lists, Profile and the 2FA pages get the v1 field and primary button look too (no layout change there). PasswordField's toggle now reads "Show"/"Hide" visibly, accessible name stays "Show password"/"Hide password" from `auth.fields.show/hide` (was hard-coded English).
- The "Strength: … (advice, not a requirement)" line in PasswordField is still hard-coded English (pre-existing, not in the frame).
- No new binary assets (the G mark is inline SVG).

## Review (especialista, 2026-11-01)
verdict: CORRIGIDO_E_APROVADO
re-run: stack merged in agent/especialista/IMPL-web-batch2 (main merged in); `bun run lint` 185 files ok; `bun run typecheck` ok; `bun run test` green except 4 load timeouts in the full run (SearchPalette, SuggestCreate, Lists/Edit), all 15 green when re-run alone; mobile `src/i18n` Jest 7 ok
checked: sdd 09 §7.4, authz (no server change; Google link is the existing /auth/{provider}/redirect), secrets (none), i18n (additive keys in en/pt-BR/es; parity test green), scope (shared Field/PasswordField/SubmitButton restyle noted in the auth report), tests
fixes: 1. mobile bundled catalogues recopied after the new auth.* keys (0c4927e)
gate on main after merge: see FIX-app-item-cover-urls review; lighthouse budgets already red on main (FIX-lighthouse-budgets-main)
PRONTO_PARA_MERGE agent/especialista/IMPL-web-batch2 (merged f99bafd)
