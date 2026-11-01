"""Contract tests for the hand-written packaging files (no network)."""

import json
import tomllib
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
REPO = ROOT.parents[1]


def test_version_matches_openapi_and_typescript_sdk():
    version = tomllib.loads((ROOT / "pyproject.toml").read_text(encoding="utf-8"))["project"]["version"]
    contract = json.loads((REPO / "openapi" / "api-v1.json").read_text(encoding="utf-8"))["info"]["version"]
    npm = json.loads((REPO / "sdks" / "ts" / "package.json").read_text(encoding="utf-8"))["version"]
    assert version == contract == npm


def test_readme_documents_quota_headers_not_legacy_ratelimit():
    readme = (ROOT / "README.md").read_text(encoding="utf-8").lower()
    assert "x-quota-remaining" in readme
    assert "ratelimit-" not in readme
