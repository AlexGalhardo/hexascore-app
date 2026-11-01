# Security policy

## Supported versions

Only the latest release and the code running at <https://hexascore.app> receive security fixes. Older releases are not patched.

## Reporting a vulnerability

Please do not open a public issue, discussion or pull request for a security vulnerability.

Report it privately through one of these channels:

- GitHub: the [Report a vulnerability](https://github.com/AlexGalhardo/hexascore-app/security/advisories/new) form (private security advisory).
- Email: [support@hexascore.app](mailto:support@hexascore.app), with "Security" in the subject.

Include what you found, where (URL, endpoint or file), how to reproduce it, and the impact you expect. Proofs of concept are welcome; do not include other users' data.

## What to expect

- Acknowledgement within 3 business days.
- An assessment and, when confirmed, a fix plan within 14 days, depending on severity.
- Credit in the release notes once the fix ships, unless you prefer to stay anonymous.

## Scope

In scope: this repository, <https://hexascore.app>, `www.hexascore.app`, the public API at `api.hexascore.app` and the Android app.

Out of scope:

- Denial of service and volumetric or load testing.
- Automated scanning that generates heavy traffic; the site rate-limits and blocks automated clients.
- Social engineering, phishing and physical attacks.
- Findings in third-party services (Railway, Cloudflare, Resend, Stripe, Google) that are not caused by HexaScore's configuration.
- Reports without a demonstrated security impact (for example missing headers on pages without sensitive data).

## Safe harbour

Research done in good faith, within this scope, without accessing or modifying other users' data and without degrading the service, will not lead to legal action from the maintainer. Use your own test accounts.

## Data protection requests

Requests about personal data under the LGPD go to [privacy@hexascore.app](mailto:privacy@hexascore.app), not to the security channel.
