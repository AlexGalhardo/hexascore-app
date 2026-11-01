# R6: Disk full

Alert A11 (Postgres volume over 80 %, Railway monitor). Never resize the volume without the owner's yes (`AGENTS.md`).

1. Find the large relations:

```sql
SELECT relname, pg_size_pretty(pg_total_relation_size(oid)) FROM pg_class
WHERE relkind = 'r' ORDER BY pg_total_relation_size(oid) DESC LIMIT 15;
```

1. Prune the high-churn event tables that have a retention rule (`search_events`, `rating_events`) with the documented retention job; do not delete other tables by hand.
2. Rebuild a bloated search index: `/app/app search:reindex` (`railway ssh`) rewrites every SDD 07 §5 index with `REINDEX INDEX CONCURRENTLY` and rebuilds the suggest rows.
3. Still over 80 %: take a budget decision ([cost](cost.md)); the volume resize is the owner's.
