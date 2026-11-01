# Runbooks

Operator procedures from SDD 12 §7.8 and §7.9. Alerts are `alert=<code>` lines in Railway log search (D-35: Railway native tools only, no Grafana). Rehearsals of R1 and R7 on staging are owner-gated and not recorded here yet.

| Alert | Meaning | Runbook |
| --- | --- | --- |
| A1, A2 | app down, readiness fails | [rollback](rollback.md), [ssr-crash-loop](ssr-crash-loop.md) |
| A3, A4, A16 | 5xx rate, latency, logs gone | [ssr-crash-loop](ssr-crash-loop.md), [rollback](rollback.md) |
| A5 | SSR fallback on a public route | [ssr-crash-loop](ssr-crash-loop.md) |
| A6 | job kind failed twice in a row | [ingestion-stuck](ingestion-stuck.md) for ingestion kinds, otherwise the job's own logs |
| A7 | ingestion failed or stale | [ingestion-stuck](ingestion-stuck.md) |
| A8 | backup missing | [restore](restore.md) |
| A9 | moderation backlog | [suggestion-backlog](suggestion-backlog.md) |
| A10 | quota flush failing | [quota-drift](quota-drift.md) |
| A11 | disk over 80 % | [disk-full](disk-full.md) |
| A12 | connection pool near its limit | [pg-connections](pg-connections.md) |
| A13, A14 | score drift, trending stale | job logs; same pattern as [ingestion-stuck](ingestion-stuck.md) |
| A15, A17 | cost | [cost](cost.md) |
| leaked credential | any | [leaked-secret](leaked-secret.md), then [incident-response](incident-response.md) |

| R | Runbook |
| --- | --- |
| R1 | [ssr-crash-loop](ssr-crash-loop.md) |
| R2 | [pg-connections](pg-connections.md) |
| R3 | [ingestion-stuck](ingestion-stuck.md) |
| R4 | [quota-drift](quota-drift.md) |
| R5 | [suggestion-backlog](suggestion-backlog.md) |
| R6 | [disk-full](disk-full.md) |
| R7 | [rollback](rollback.md) |
| R8 | [restore](restore.md) |
| R9 | [leaked-secret](leaked-secret.md) |

Search the logs with the code, for example `alert=A10`. Alert lines carry counts and names, never ids, emails or raw paths.

Release procedures: [sdk-release](sdk-release.md) (SDKs), [mobile-release](mobile-release.md) (Android app with EAS), [release-range](release-range.md) (how `scripts/release.sh` picks the changelog range).
