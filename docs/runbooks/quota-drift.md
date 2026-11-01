# R4: Quota drift

Alert A10 (`alert=A10`, 5 consecutive failed quota flushes). The meter keeps deltas in memory and fails closed with 503 after 10 minutes unflushed (SDD 05 §7). A mismatch between the counter and the daily table is customer-facing: always correct in the customer's favour (NFR-05-2).

1. Find the cause in the `alert=A10` and `api usage flush failed` lines (database error text). A database outage or lock is the usual cause; fix that first ([pg-connections](pg-connections.md)).
2. Recount a customer's month from the daily table and compare with the counter:

```sql
SELECT sum(count - refunded) AS daily FROM api_usage_daily
WHERE user_id = <id> AND day >= date_trunc('month', now())::date;
SELECT used FROM api_quota_counters WHERE user_id = <id> AND period = to_char(now(), 'YYYYMM');
```

1. If the counter is higher, lower it to the recount (customer's favour). If lower, leave it.
2. A service restart flushes pending deltas on shutdown and reloads counters from the database.
