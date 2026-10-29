---
paths:
  - "**/*.go"
  - "ROADMAP.md"
---

# Security rules

- The OWASP categories in `docs/sdd/13-security.md` are a contract: each behaviour names the test that proves it (`TestSecurity_<Axx>_<Behaviour>`, vukem convention). Until SDD 13 exists, ROADMAP.md names the category per task.
- Secure by default; unsafe behaviour is an explicit opt-out with an "Insecure" name, never a missing opt-in.
- Fail closed: if a check cannot run (store down, policy missing, key missing), deny.
- Misconfiguration that weakens security is a boot error, not a warning.
- SQL: bound parameters only; identifiers validated against model metadata; never concatenate input into SQL.
- HTML via `html/template`; never wrap user input in `template.HTML`.
- Randomness from `crypto/rand`; compare secrets with `subtle.ConstantTimeCompare` / `hmac.Equal`.
- Never log secrets, tokens, passwords, `Authorization` or `Cookie` headers; use `config.Secret` (vukem name, not yet confirmed at <https://vukem.dev/docs>).
- Client-facing errors never include stack traces, SQL, paths or config (CWE-209).
- File access through `os.Root`; outbound requests to user-supplied URLs through `httpclient.Safe()`.
- Load the `golang-security` skill for any change in accounts, payments, uploads, webhooks or API keys.
