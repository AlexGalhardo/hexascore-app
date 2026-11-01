# Monthly cost

A17 projected spend and the US$30 hard limit (D-27). The Railway usage API is not verified. `projected-spend.yml` logs a reminder on UTC day 20. Set the repository variable `RAILWAY_USAGE_USD` to the month-to-date figure when you want the formula (`month-to-date × days in cycle / elapsed days`). It alerts above US$27 from day 20 on. Telegram is sent only when `TELEGRAM_BOT_TOKEN` and `TELEGRAM_CHAT_ID` exist (P2.C.1.2).

On day 20, open Railway usage, compare the projection with US$27, and raise the US$30 hard limit before it stops the service if the month will pass US$27.

## Monthly checklist

Run on the first day of the month (SDD 12 section 7.9).

1. `railway usage limit status`.
2. Invoice against the US$30 hard limit.
3. Log volume and trace span count against Railway limits.
4. Postgres size against the plan.
5. Bucket size (about 2.1 GB per 100k items, NFR-06-7).
6. Egress; cover egress over US$5 per month triggers the CDN decision (SDD 11 open issue 2).
7. Client intake volume (`client_reports` rows per day) and that the 7-day prune ran.

Budget-increase predictors (the first step is splitting out the worker, D-08): CPU over 70 % for 7 days, Postgres over 60 % of the plan, night job backlog over 1 hour.
