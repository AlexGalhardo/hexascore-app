---
paths:
  - "apps/mobile/**"
  - ".github/workflows/mobile-release.yml"
---

# Mobile release rules

Procedure: [docs/runbooks/mobile-release.md](../../docs/runbooks/mobile-release.md). Skill: `hexascore-mobile-release`.

- EAS only through `bunx eas-cli@27.2.0` (pinned); profiles live in `apps/mobile/eas.json`. Android only until iOS enters (SDD 10).
- Never ask for, print or commit passwords, `EXPO_TOKEN`, keystores or the Play service account key. Login is the owner's (`eas login`); CI uses the `EXPO_TOKEN` secret.
- Stop and ask the owner on: uncommitted git changes before a release, any `production` version change, a failed build (diagnose first), every `eas submit` (per build id), any `eas update` to a production channel.
- `version` is edited in `app.config.ts` only after the owner agrees; `versionCode` is EAS remote with `autoIncrement`, never by hand.
- Lint, typecheck and Jest pass before any build.
- OTA (`eas update`) only for JS/asset changes with an unchanged fingerprint; anything native, a new permission or a new feature needs a store build. OTA stays off until `expo-updates` is approved.
- Submissions go to the Play internal track as drafts; promotion and staged rollout are owner decisions.
- No destructive EAS command (delete build, channel, branch, credentials) without an explicit owner yes.
