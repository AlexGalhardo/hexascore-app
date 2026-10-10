"""Cover WebP + Vercel Blob path. fixtures/commons_cover.jpg is a downscaled copy of
"Bryan Cranston by Gage Skidmore 2.jpg" (Gage Skidmore, CC BY-SA 3.0, Wikimedia Commons)."""

import io
import json
from typing import Any

import httpx
import pytest
from PIL import Image
from pydantic import ValidationError

import covers
from tests.conftest import FIXTURES, mock_client

SOURCE = (FIXTURES / "commons_cover.jpg").read_bytes()
SOURCE_URL = "https://upload.wikimedia.org/cover.jpg"
BLOB_HOST = "https://store.public.blob.vercel-storage.com/"


def png(mode: str, size: tuple[int, int], **save: Any) -> bytes:
    out = io.BytesIO()
    Image.new(mode, size).save(out, "PNG", **save)
    return out.getvalue()


def blob_handler(calls: list[httpx.Request], source: bytes = SOURCE) -> Any:
    def handler(request: httpx.Request) -> httpx.Response:
        calls.append(request)
        if request.method == "GET":
            return httpx.Response(200, content=source, headers={"content-type": "image/jpeg"})
        return httpx.Response(200, json={"url": BLOB_HOST + request.url.params["pathname"]})

    return handler


def test_to_webp_resizes_and_strips_metadata() -> None:
    webp = covers.to_webp(SOURCE)
    with Image.open(io.BytesIO(webp)) as image:
        assert image.format == "WEBP"
        assert image.size == (519, 600)  # 720x833 source, aspect kept
        assert "exif" not in image.info and "icc_profile" not in image.info
    assert len(webp) < len(SOURCE) / 2


def test_to_webp_keeps_alpha_and_never_upscales() -> None:
    with Image.open(io.BytesIO(covers.to_webp(png("RGBA", (460, 215))))) as image:
        assert (image.mode, image.size) == ("RGBA", (460, 215))
    with Image.open(io.BytesIO(covers.to_webp(png("L", (40, 60))))) as image:
        assert (image.mode, image.size) == ("RGB", (40, 60))


def test_to_webp_applies_exif_orientation() -> None:
    exif = Image.Exif()
    exif[0x0112] = 6  # rotate 90° clockwise on display
    out = io.BytesIO()
    Image.new("RGB", (300, 200)).save(out, "JPEG", exif=exif)
    with Image.open(io.BytesIO(covers.to_webp(out.getvalue()))) as image:
        assert image.size == (200, 300)


def test_to_webp_thumb_fits_list_size() -> None:
    thumb = covers.to_webp(SOURCE, covers.THUMB_SIZE, covers.THUMB_QUALITY)
    with Image.open(io.BytesIO(thumb)) as image:
        assert image.size == (96, 111)  # 720x833 fits 96x144 by width, aspect kept
        assert "exif" not in image.info and "icc_profile" not in image.info
    assert len(thumb) < 8 * 1024
    with Image.open(io.BytesIO(covers.to_webp(png("RGB", (40, 60)), covers.THUMB_SIZE))) as image:
        assert image.size == (40, 60)


FULL = BLOB_HOST + covers.blob_pathname("item-1", SOURCE)
THUMB = BLOB_HOST + covers.blob_pathname("item-1", SOURCE, "-t")


def test_store_cover_uploads_content_addressed_full_and_thumb() -> None:
    calls: list[httpx.Request] = []
    urls = covers.store_cover(mock_client(blob_handler(calls)), "tok", "item-1", SOURCE_URL, None, None)

    pathname = covers.blob_pathname("item-1", SOURCE)
    assert pathname.startswith("covers/item-1-") and pathname.endswith(".webp")
    assert pathname.removesuffix(".webp") + "-t.webp" == THUMB.removeprefix(BLOB_HOST)
    assert urls == (FULL, THUMB)
    assert [r.method for r in calls] == ["GET", "PUT", "PUT"]
    for put in calls[1:]:
        assert put.headers["authorization"] == "Bearer tok"
        assert put.headers["x-content-type"] == "image/webp"
        assert put.headers["x-cache-control-max-age"] == "31536000"
        assert put.content[:4] == b"RIFF" and put.content[8:12] == b"WEBP"
    assert len(calls[2].content) < len(calls[1].content) / 3


def test_store_cover_skips_upload_when_same_source_is_stored() -> None:
    calls: list[httpx.Request] = []
    assert covers.store_cover(mock_client(blob_handler(calls)), "tok", "item-1", SOURCE_URL, FULL, THUMB) == (
        FULL,
        THUMB,
    )
    assert [r.method for r in calls] == ["GET"]


def test_store_cover_adds_only_missing_thumb() -> None:
    calls: list[httpx.Request] = []
    urls = covers.store_cover(mock_client(blob_handler(calls)), "tok", "item-1", SOURCE_URL, FULL, None)
    assert urls == (FULL, THUMB)
    assert [(r.method, r.url.params.get("pathname")) for r in calls] == [
        ("GET", None),
        ("PUT", covers.blob_pathname("item-1", SOURCE, "-t")),
    ]


def test_store_cover_reuploads_when_source_changed() -> None:
    calls: list[httpx.Request] = []
    stale = BLOB_HOST + "covers/item-1-0000000000000000.webp"
    stale_thumb = BLOB_HOST + "covers/item-1-0000000000000000-t.webp"
    urls = covers.store_cover(mock_client(blob_handler(calls)), "tok", "item-1", SOURCE_URL, stale, stale_thumb)
    assert urls == (FULL, THUMB) and [r.method for r in calls] == ["GET", "PUT", "PUT"]


def test_upload_rejects_unexpected_blob_response() -> None:
    client = mock_client(lambda _: httpx.Response(200, json={"pathname": "x"}))
    with pytest.raises(ValidationError):
        covers.upload_to_blob(client, "tok", "covers/x.webp", b"webp")


class FakeConn:
    rowcount = 0

    def __init__(self) -> None:
        self.updates: dict[str, dict] = {}
        self.statements: list[str] = []

    def execute(self, sql: str, *_: object) -> FakeConn:
        self.statements.append(sql)
        return self

    def cursor(self) -> FakeConn:
        return self

    def __enter__(self) -> FakeConn:
        return self

    def __exit__(self, *_: object) -> None:
        pass

    def executemany(self, _sql: str, params_seq: list[tuple[Any, Any]]) -> None:
        for params in params_seq:
            self.updates[params[1]] = json.loads(json.dumps(params[0].obj))

    def commit(self) -> None:
        pass


def steam(header: str) -> dict[str, dict]:
    return {"steam": {"header_image": header}}


def test_commons_cover_only_from_film_poster() -> None:
    claim = [{"value": "Mario arcade.jpg"}]
    assert covers.commons_file({"wikidata": {"claims": {"P18": claim}}}) is None
    assert covers.commons_file({"wikidata": {"claims": {"P18": claim, "P3383": [{"value": "Poster.jpg"}]}}}) == (
        "Poster.jpg"
    )


def test_only_a_full_run_removes_loose_commons_covers(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.delenv("BLOB_READ_WRITE_TOKEN", raising=False)
    monkeypatch.setattr(covers, "targets", lambda *_: [])
    limited, sharded, full = FakeConn(), FakeConn(), FakeConn()
    covers.run(limited, limit=20, client=mock_client(blob_handler([])))
    covers.run(sharded, client=mock_client(blob_handler([])), shard=(1, 4))
    covers.run(full, client=mock_client(blob_handler([])))
    assert limited.statements == []
    assert sharded.statements == []
    assert len(full.statements) == 1
    assert "'Wikimedia Commons'" in full.statements[0]


def test_run_without_token_keeps_source_url(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.delenv("BLOB_READ_WRITE_TOKEN", raising=False)
    monkeypatch.setattr(covers, "targets", lambda *_: [("a", "game", steam(SOURCE_URL), None, None)])
    conn = FakeConn()
    calls: list[httpx.Request] = []
    assert covers.run(conn, client=mock_client(blob_handler(calls))) == 1
    assert conn.updates == {"a": {"coverUrl": SOURCE_URL}}
    assert calls == []


def test_run_isolates_per_item_failures(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("BLOB_READ_WRITE_TOKEN", "tok")
    rows = [
        ("broken", "game", steam("https://cdn.example/broken.jpg"), None, None),
        ("gone", "game", steam("https://cdn.example/404.jpg"), None, None),
        ("ok", "game", steam(SOURCE_URL), None, None),
    ]
    monkeypatch.setattr(covers, "targets", lambda *_: rows)
    ok = blob_handler([])

    def handler(request: httpx.Request) -> httpx.Response:
        if request.url.path == "/broken.jpg":
            return httpx.Response(200, content=b"not an image")
        if request.url.path == "/404.jpg":
            return httpx.Response(404)
        return ok(request)

    conn = FakeConn()
    assert covers.run(conn, client=mock_client(handler)) == 1
    assert list(conn.updates) == ["ok"]
    assert conn.updates["ok"]["coverUrl"] == BLOB_HOST + covers.blob_pathname("ok", SOURCE)
    assert conn.updates["ok"]["coverThumbUrl"] == BLOB_HOST + covers.blob_pathname("ok", SOURCE, "-t")


def test_run_backfills_thumb_for_legacy_cover(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("BLOB_READ_WRITE_TOKEN", "tok")
    monkeypatch.setattr(covers, "targets", lambda *_: [("item-1", "game", steam(SOURCE_URL), FULL, None)])
    conn = FakeConn()
    calls: list[httpx.Request] = []
    assert covers.run(conn, client=mock_client(blob_handler(calls))) == 1
    assert conn.updates == {"item-1": {"coverUrl": FULL, "coverThumbUrl": THUMB}}
    assert [r.method for r in calls] == ["GET", "PUT"]


def test_run_skips_item_with_both_sizes_current(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("BLOB_READ_WRITE_TOKEN", "tok")
    monkeypatch.setattr(covers, "targets", lambda *_: [("item-1", "game", steam(SOURCE_URL), FULL, THUMB)])
    conn = FakeConn()
    assert covers.run(conn, client=mock_client(blob_handler([]))) == 0
    assert conn.updates == {}


def test_run_processes_items_concurrently(monkeypatch: pytest.MonkeyPatch) -> None:
    import threading
    import time

    monkeypatch.delenv("BLOB_READ_WRITE_TOKEN", raising=False)
    rows = [(f"i{n}", "game", {}, None, None) for n in range(20)]
    monkeypatch.setattr(covers, "targets", lambda *_: rows)
    active = {"now": 0, "max": 0}
    lock = threading.Lock()

    def slow_cover(item_type: str, records: dict[str, dict]) -> covers.Cover:
        with lock:
            active["now"] += 1
            active["max"] = max(active["max"], active["now"])
        time.sleep(0.02)
        with lock:
            active["now"] -= 1
        return covers.Cover(SOURCE_URL)

    monkeypatch.setattr(covers, "pick_source_cover", slow_cover)
    conn = FakeConn()
    assert covers.run(conn, client=mock_client(blob_handler([])), workers=4) == 20
    assert len(conn.updates) == 20
    assert active["max"] > 1
