# REV-P3.A.5.3

verdict: APROVADO

final branch: agent/reviewer/P3.A.5.3-fix (291d41b)

checked: POST /api/v1/webhooks/{id}/test returns 200 {delivered,status}. Op.Status is set only on testWebhook; createWebhook stays 201. Free plan is 403 plan-required before lookup. Another user and a disabled endpoint are 404 and the URL is not called. Ping calls svc.deliver, so a failed ping does not bump failure_streak. OpenApp stays https-only with an empty Allow. openPing allows HTTP on 127.0.0.0/8 with Retries -1 only for KindTest. Fan-out still calls OpenApp. Prune deletes webhook_deliveries strictly older than 30 days; a row at exactly 30 days stays. Cron is 40 4 * * * UTC, kind webhooks.prune_deliveries. Counter api.webhook.deliveries uses attribute outcome. No delivery-log route, no new dependency, no vukem edit, no applied migration edit, no P3.A.6.

fix: attempt status and the outcome label select webhook_deliveries by event_id. The endpoint's newest row can be a concurrent fan-out and would report the wrong status.

acceptance on agent/reviewer/P3.A.5.3-fix: go test -tags=integration ./tests -count=1 -run "TestAPIWebhookTest|TestPruneDeliveries|TestMetrics_WebhookOutcome" with TEST_DATABASE_URL on 127.0.0.1:5433 -> ok (6.373s), exit 0. TestSecurity_A01_HexaRouteStackInventory fails only on pre-existing GET /api/items/{id}/ratings. The four webhook inventory lines are not that failure.
