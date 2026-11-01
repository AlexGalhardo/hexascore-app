# R9: Leaked secret

Any credential in a log, commit, screenshot or third-party breach. Severity is SEV1 for `APP_KEY` and Stripe keys ([incident-response](incident-response.md)). Never print variable values.

1. Revoke the secret at the provider (Stripe, Resend, Google OAuth, Telegram, TMDB, Railway token, GitHub token).
2. Set the new value in Railway variables and redeploy.
3. `APP_KEY`: follow <https://vukem.dev/docs/security/encryption> "Key rotation" and "After a leak". Never keep a leaked key in `APP_PREVIOUS_KEYS`.
4. Search Railway logs since the leak for use of the old credential. Export them first; retention is short.
5. Record the incident and run the notification steps in [incident-response](incident-response.md).
