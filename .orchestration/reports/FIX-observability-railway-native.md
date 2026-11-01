# FIX-observability-railway-native
status: concluida
branch: agent/senior/FIX-observability-railway-native @ 21c1a30 (from main 1198add)
files: removed ops/grafana/, app/metrics/, tests/metrics_test.go, app/scheduler/instrument.go; edited bootstrap/{app,recommend_provider}.go, app/webhooks/sender.go, app/recommend/{primitives/flush,trending/snapshot}.go, tests/{webhook_ops,lgpd}_test.go, go.mod/go.sum (go mod tidy drops otel, otel/metric, otel/sdk/metric as direct deps), resources/lang/{en,pt-BR,es}.json (policy processors line), docs/sdd/{12-observability,13-security,open-issues}.md, docs/legal/privacy-review.md
acceptance:
- `go build ./...`, `go vet -tags integration ./...` clean
- `go test -tags integration ./tests -run 'TestAPIWebhookTest|TestPruneDeliveries|TestReady|TestLGPD|Policy'` ok; `go test ./app/recommend/... ./app/webhooks/... ./app/scheduler/...` ok
- golangci-lint on bootstrap, app/recommend, app/webhooks, app/scheduler, tests: 1 issue, pre-existing and not touched (tests/foryou_test.go:209 perfsprint, from P6.D.2.2)
- no remaining Grafana/promtool reference outside CHANGELOG.md, ROADMAP.md (orchestrator) and the grafana/k6 load-test image in qa/ (the k6 runner, not observability)
decisions:
- Railway OTLP takes traces only, so every metric instrument was dropped (no export target): http/db/queue/ssr/schedule/webhook/trending counters and gauges. Replaced by structured logs: SSR fallback already logs; trending snapshot logs eligible/prior/churn per type; dropped trending counter keys log alert=trending_counter_dropped. webhook deliveryOutcome helper deleted (metric-only).
- SDD 12 §6.2 is now "Signals" (Railway service metrics, log search, traces, alert= log lines); §7.3-7.6 rewritten Railway-native; alerts A1-A17 re-sourced (Railway healthcheck/monitor/usage, post-deploy smoke test, in-app checks logging alert=<code>). Policy page processors no longer list Grafana Cloud.
risks: the in-app checks of §7.5 (A3, A4, A6-A9, A12-A14, A16) are specified but not implemented here; they are new work for the ROADMAP lines below. heartbeat age is still used by /up readiness only.
roadmap lines for the orchestrator to change: P2.C.2.1 (1541-1549, Grafana stack -> Railway OTLP traces, OTEL_METRICS_EXPORTER=none), P2.C.2.2 (1551-1559, [x] metrics catalogue now removed; note the reversal), P2.C.3.1 (1563-1571, promtool/Grafana rules -> in-app alert= checks with fire/clear fixtures), P2.C.3.2 (1573-1581, Grafana synthetics -> Railway healthcheck), P2.C.3.4-ish dashboards (1604-1611, ops/grafana dashboards and 6,000 series), 1631 (runbook links to ops/alerts/rules.yaml), 1643 (TestMetrics_ in acceptance), 1716 (policy processors list), 1807-1809 (backup gauge/A8), 2088 (ops/grafana/synthetics.md), 2399-2404 (TestMetrics_WebhookOutcome), 2891-2896 (community metrics, grafana moderation, A9, TestMetrics_CommunityGauges), 3591 (ingestion metrics.go, ops/grafana/ingestion.json, a7.yaml), 3657 (a14.yaml); also SDD 06 §10, 08 §6, 09 and SDDs 02/04/05/07 metric rows still name hexa.* instruments (not edited here, outside the row's file list).
changelog: Removed: app metrics instruments and Grafana alert artifacts; Changed: observability is Railway-native only (D-35 amended); policy page processors list.
roadmap: D-35

## Review (especialista, 2026-10-31)
verdict: CORRIGIDO_E_APROVADO
re-run: `go build ./...`, `go vet -tags integration ./...` → ok (main 679005b merged in, clean); `go test ./app/recommend/... ./app/webhooks/... ./app/scheduler/...` → ok; golangci-lint on bootstrap, app/recommend, app/webhooks, app/scheduler → 0 issues; `--build-tags integration ./tests/...` → 0 issues after fix 1; markdownlint on the four docs → 0
checked: SDD 12 §6.2/§7.5 (D-35 amended), authz (none touched), secrets (none), sql (deleted query only), i18n (policy processors line edited in en/pt-BR/es, parity kept), scope (go.mod tidy drops otel; testify shows up as indirect from tidy), tests
fixes: 1. tests/foryou_test.go perfsprint (strconv.FormatInt) (cb251c1) 2. SDD 12 §6.2 names the non-A-code key `alert=trending_counter_dropped` emitted by primitives/flush.go (d2e386d)
pending (Docker down): `go test -tags integration ./tests -run 'TestAPIWebhookTest|TestPruneDeliveries|TestReady|TestLGPD|Policy|TestForYou|TestRecoEval'`
gate on main after merge (3cc01ff): go build ok, go vet -tags integration ok; full gate waits Docker
PRONTO_PARA_MERGE agent/especialista/FIX-observability-railway-native-fix
