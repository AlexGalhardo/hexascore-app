# Monthly cost

A17 projected spend and the US$30 hard limit (D-27). The Railway usage API is not verified. `projected-spend.yml` logs a reminder on UTC day 20. Set the repository variable `RAILWAY_USAGE_USD` to the month-to-date figure when you want the formula (`month-to-date × days in cycle / elapsed days`). It alerts above US$27 from day 20 on. Telegram is sent only when `TELEGRAM_BOT_TOKEN` and `TELEGRAM_CHAT_ID` exist (P2.C.1.2).

On day 20, open Railway usage, compare the projection with US$27, and raise the US$30 hard limit before it stops the service if the month will pass US$27.
