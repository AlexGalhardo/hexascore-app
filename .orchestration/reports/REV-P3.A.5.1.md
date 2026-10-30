# REV-P3.A.5.1

verdict: CORRIGIDO_E_APROVADO

final branch: agent/reviewer/P3.A.5.1-fix (4f6ddf0)

fix: createWebhook answered createdAt 0001-01-01 because Create never read the column default. HTTPS:// passed PublicURL (url.Parse lowercases the scheme) and then failed CHECK (url LIKE 'https://%') as an unhandled 500. endpoints.go now returns RETURNING created_at and stores a lowercase https:// scheme.

checked: secret from webhook.NewSecret, sealed with AD webhook_endpoints.secret_enc, absent from list SQL and WebhookJSON. Pro cap 2, business cap 10, any other plan 403 plan-required. Private, link-local, metadata, userinfo, and mixed DNS fail closed. Create does not fetch the URL. Quota skips /api/v1/webhooks. Delete is parameterized and scoped by user_id. New migration only. No fan-out, delivery log, dependency, or popularity column.

acceptance on the fix branch: go test ./app/webhooks -count=1 -run TestPublicURL ok; go test -tags=integration ./tests -count=1 -run 'TestAPIWebhooks|TestSecurity_A10_WebhookSSRF|TestSecurity_A02_WebhookSecretEncrypted' ok.
