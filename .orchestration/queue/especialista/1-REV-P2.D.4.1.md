# Review P2.D.4.1

Re-run:

`TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable go test -tags=integration ./tests -count=1 -run 'TestLGPD_ExportDeleteOptOut|TestRetentionJobs'`

Check `app/privacy/retention.go` only strips `_client.ip` and `_client.ua`, skips a missing `client_reports` table, and calls `Exports.Expire`. Policy objects in `resources/lang/{en,pt-BR,es}.json` must not contain Vercel, Neon, Upstash, passkey or AbacatePay. Do not tick a staging curl. Do not push. Report: `.orchestration/reports/P2.D.4.1.md`.
