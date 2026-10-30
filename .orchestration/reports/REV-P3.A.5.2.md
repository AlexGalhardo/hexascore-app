# REV-P3.A.5.2

verdict: APROVADO

final branch: agent/implementer/P3.A.5.2 (8b89125)

checked: fan-out reads at most 10,000 api_changes after webhook_cursor, coalesces one event per item, and enqueues item.batch of at most 100 through webhook.Sender and SQLDeliveryLog. The cursor advances in the same transaction as DispatchKindTx. Payload is ids and field names only. 410 calls OnGone and does not mail. Three exhausted deliveries disable the endpoint and send one mail. OpenApp uses zero Options. The https CHECK is dropped only in the integration test database. Retries are the sender knob (8 from 30s) on the queue registration. No delivery-log route and no P3.A.5.3. Secret stays encrypted; fan-out does not log the secret or the URL.

acceptance on the implementer branch: go test ./app/webhooks -count=1 -run "TestCoalesce|TestChunks" ok (0.022s); go test -tags=integration ./tests -count=1 -run TestWebhookDelivery ok (2.847s).
