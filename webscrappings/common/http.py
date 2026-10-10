import logging
import os
import threading
import time
from collections.abc import Callable
from email.utils import parsedate_to_datetime
from typing import Any
from urllib.parse import urlsplit

import httpx

log = logging.getLogger(__name__)

DEFAULT_CONTACT_URL = "https://github.com/AlexGalhardo/claude-code-internet-api"
RETRY_STATUS = {429, 500, 502, 503, 504}

# Minimum seconds between requests per host, kept under each provider's published limits.
HOST_INTERVALS: dict[str, float] = {
    "query.wikidata.org": 1.0,
    "www.wikidata.org": 0.2,
    "commons.wikimedia.org": 0.2,
    # Anonymous Wikimedia REST answers 429 well below its documented ceiling: 5 req/s per wiki stays clear of it.
    "en.wikipedia.org": 0.2,
    "pt.wikipedia.org": 0.2,
    "es.wikipedia.org": 0.2,
    "openlibrary.org": 0.4,
    "store.steampowered.com": 1.6,
    "catalog.gog.com": 0.5,
    # TMDB allows ~40 req/s per IP; IGDB allows 4 req/s per client.
    "api.themoviedb.org": 0.04,
    "api.igdb.com": 0.3,
    "image.tmdb.org": 0.02,
    "images.igdb.com": 0.05,
    # Cover downloads and Blob uploads: CDNs and the Blob API take far more than the 2 req/s default, which would cap
    # covers at ~1 item/s (two uploads each). Open Library cover-by-ID requests are not rate-limited.
    # upload.wikimedia.org answers 429 ("bot is making too many requests") at 10 req/s; keep it at 2 req/s.
    "covers.openlibrary.org": 0.1,
    "upload.wikimedia.org": 0.5,
    "shared.akamai.steamstatic.com": 0.02,
    "images.gog-statics.com": 0.05,
    "vercel.com": 0.01,
}
DEFAULT_INTERVAL = 0.5


def user_agent() -> str:
    return f"HexascoreBot/0.1 (+{os.environ.get('INGEST_CONTACT_URL') or DEFAULT_CONTACT_URL})"


def retry_after_seconds(value: str | None, now: float) -> float | None:
    if not value:
        return None
    if value.strip().isdigit():
        return float(value)
    try:
        return max(0.0, parsedate_to_datetime(value).timestamp() - now)
    except TypeError, ValueError:
        return None


class HttpClient:
    def __init__(
        self,
        *,
        transport: httpx.BaseTransport | None = None,
        max_attempts: int = 5,
        timeout: float = 60.0,
        sleep: Callable[[float], None] = time.sleep,
        clock: Callable[[], float] = time.monotonic,
    ) -> None:
        self._client = httpx.Client(
            transport=transport,
            timeout=timeout,
            follow_redirects=True,
            headers={"User-Agent": user_agent()},
        )
        self._max_attempts = max_attempts
        self._sleep = sleep
        self._clock = clock
        self._last: dict[str, float] = {}
        self._lock = threading.Lock()

    def _throttle(self, host: str) -> None:
        # Thread-safe: each caller reserves the next free slot for the host under the lock, then sleeps
        # outside it, so concurrent workers share one per-host rate instead of each getting its own.
        interval = HOST_INTERVALS.get(host, DEFAULT_INTERVAL)
        with self._lock:
            now = self._clock()
            slot = max(now, self._last.get(host, float("-inf")) + interval)
            self._last[host] = slot
        if slot > now:
            self._sleep(slot - now)

    def request(self, method: str, url: str, **kwargs: Any) -> httpx.Response:
        host = urlsplit(url).hostname or ""
        for attempt in range(1, self._max_attempts + 1):
            self._throttle(host)
            try:
                response = self._client.request(method, url, **kwargs)
            except httpx.TransportError as exc:
                if attempt == self._max_attempts:
                    raise
                log.warning("%s %s failed (%s), retry %d", method, url, exc, attempt)
                self._sleep(min(2**attempt, 60))
                continue
            if response.status_code not in RETRY_STATUS or attempt == self._max_attempts:
                return response
            delay = retry_after_seconds(response.headers.get("Retry-After"), time.time())
            delay = min(delay if delay is not None else 2**attempt, 300)
            # Back off the whole host, not just this thread: concurrent workers would otherwise keep hitting it.
            with self._lock:
                self._last[host] = max(self._last.get(host, float("-inf")), self._clock() + delay)
            log.warning("%s %s -> %d, retry %d in %.0fs", method, url, response.status_code, attempt, delay)
            self._sleep(delay)
        raise AssertionError("unreachable")

    def get_json(self, url: str, **kwargs: Any) -> Any:
        response = self.request("GET", url, **kwargs)
        response.raise_for_status()
        return response.json()

    def close(self) -> None:
        self._client.close()
