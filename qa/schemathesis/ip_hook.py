"""One client address per Schemathesis call.

/api/v1 allows 120 requests a minute per IP (SDD 13). A scan sends more than
that from the Docker gateway, and the missing-Authorization check then sees
429 instead of 401. The compose app trusts 172.16.0.0/12, same as qa/k6.
"""

import schemathesis

_n = 0


def _fake_ip(n: int) -> str:
    return f"10.{(n >> 8) % 250}.{n % 250}.{1 + ((n * 7) % 250)}"


@schemathesis.hook
def before_call(context, case, kwargs) -> None:
    global _n
    _n += 1
    case.headers["X-Forwarded-For"] = _fake_ip(_n)
