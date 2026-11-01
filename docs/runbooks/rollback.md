# R7: Rollback

Alerts A1, A2, A3. Source: <https://docs.railway.com/guides/roll-back-bad-deploy>. A push to `main` deploys production, so a bad merge is rolled back here, not by a force-push.

1. Railway dashboard, service `app`, Deployments: roll back to the previous deployment with status `SUCCESS`. It restores the image and variables within image retention.
2. Migrations are expand/contract (ADR 0014), so the previous code stays valid on the new schema. Do not run a down migration in production.
3. Verify with `curl -fsS https://hexascore.app/up` and `curl -s "https://hexascore.app/up?checks=1"`; all four checks `ok`.
4. Update the status page (`status.hexascore.app`) when users saw errors.
5. Revert the bad commit through a PR so the next deploy does not reintroduce it.
