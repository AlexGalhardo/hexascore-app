"""Asks the web app to invalidate cached catalog pages. Best effort: failures are logged, never raised."""

import logging
import os

import httpx

from common.http import HttpClient

log = logging.getLogger("revalidate")


def revalidate(tags: list[str], client: HttpClient | None = None) -> bool:
    app_url, secret = os.environ.get("APP_URL"), os.environ.get("CRON_SECRET")
    if not app_url or not secret:
        log.info("APP_URL/CRON_SECRET not set, skipping revalidation")
        return False
    client = client or HttpClient(max_attempts=3)
    try:
        response = client.request(
            "POST",
            f"{app_url.rstrip('/')}/api/cron/revalidate",
            headers={"Authorization": f"Bearer {secret}"},
            json={"tags": tags},
        )
    except httpx.HTTPError as exc:
        log.warning("revalidate failed: %s", exc)
        return False
    if response.is_success:
        log.info("revalidated tags=%s", tags)
        return True
    log.warning("revalidate failed: HTTP %d", response.status_code)
    return False


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    revalidate(["catalog"])
