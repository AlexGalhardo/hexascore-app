# Recommendation evaluation protocol

Offline evaluation of the "For you" recommender and the gate for collaborative filtering (CF). Source of truth: [SDD 08](../sdd/08-recommendation.md) section 7.6. Code: `app/recommend/eval`.

## Protocol

1. Per user, sort ratings by time. The latest `max(2, ceil(20 %))` ratings are the test set, the rest train (`eval.HoldOut`). A user with fewer than 3 ratings is skipped.
2. Held-out items rated at least 70 overall are the positives (`eval.Positives`).
3. Build each recommender's list from the training ratings only, then score the top 10 with `eval.PrecisionAt` and `eval.NDCGAt` (binary relevance; a short list is still divided by 10 for precision).
4. `eval.Evaluate` returns the means with a 95 % percentile bootstrap interval (`eval.BootstrapCI`, 2,000 resamples, fixed seed so a run is reproducible). For a lift, bootstrap the paired per-user differences.
5. Coverage (`eval.Coverage`) is the share of the catalogue appearing in at least one list.
6. Write the report with `eval.WriteCSV` (`encoding/csv` with the formula guard: a cell starting with `=`, `+`, `-`, `@`, tab or CR gets a leading apostrophe).

## Monthly CF trigger

`jobs.RecoEvalMonthly` runs on day 1 at 05:30 UTC. It counts users with at least 10 ratings. Below 5,000 it records `triggered,false` and computes nothing else. At 5,000 or more it records `triggered,true`; the comparison is then run by hand. The CSV goes to `RECO_EVAL_DIR/reco-eval-YYYY-MM.csv` when that variable is set, else only the log has the numbers.

## Decision rule

CF ships only through a new ADR, and only with all of:

- precision@10 lift of at least 10 % over the content baseline;
- a 95 % bootstrap interval of the lift that excludes 0;
- coverage of at least 80 % of the baseline's.

The content baseline is recorded at launch + 90 days (an owner calendar item, SDD 08 section 11 issue 7).
