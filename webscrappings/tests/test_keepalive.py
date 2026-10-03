def test_keep_alive_skips_ping_while_connection_is_busy() -> None:
    # No database: a fake connection records pings; the ping thread must not touch a connection held in `busy`.
    import threading

    from common.db import busy, keep_alive

    class FakeConn:
        closed = False
        pings = 0

        def execute(self, _sql: str) -> None:
            self.pings += 1

    fake = FakeConn()
    lock = busy(fake)  # type: ignore[arg-type]
    with lock:
        keep_alive(fake, interval=0.01)  # type: ignore[arg-type]
        threading.Event().wait(0.1)
        assert fake.pings == 0
    threading.Event().wait(0.1)
    fake.closed = True
    assert fake.pings > 0
