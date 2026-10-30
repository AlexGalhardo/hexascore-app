# Contract and baseline scans

Schemathesis is 4.29.0, the legacy pin (`schemathesis/schemathesis:4.29.0`). The host run used that version (`schemathesis --version`). ZAP is 2.17.0. `qa/zap/baseline.sh` pins the image index inspected on 2026-10-30:

`ghcr.io/zaproxy/zaproxy@sha256:030d3df57735dacbb83464854758e674dd9ad01d79d12a129a8e58b1fd384238`

The linux/amd64 manifest under that index is `sha256:0075662e5c287cf0a68caebabbfea872e5dc461912ad6d7345fa02a290d0ccb6`. Source: `docker buildx imagetools inspect ghcr.io/zaproxy/zaproxy:2.17.0`, and the 2.17.0 release at <https://github.com/zaproxy/zaproxy/releases/tag/v2.17.0>.

## Schemathesis

Command, both documents, `http://127.0.0.1:3000`, `--checks all`, config `qa/schemathesis/schemathesis.toml`:

```text
schemathesis --config-file qa/schemathesis/schemathesis.toml run openapi/api-v1.json --url http://127.0.0.1:3000 -H "Authorization: Bearer $HX_TEST_KEY" --checks all
```

`openapi/api-v1.json` (2026-10-30): 17/17 operations, 1187 passed, 10 skipped, exit 0. `listItems`, `getItem`, `getEntity`, `listEntityItems` and `listEditions` are in that document and were part of the run. Warnings only: three operations mostly 404 (no fixture id), and generated strings are often 400 because the document types them as plain strings.

`openapi/app-v1.json`: 27/27 operations, 1743 passed, 1 skipped, exit 0, with `--exclude-checks allow_header_conformance,unsupported_method`. `--checks all` turns those toml flags back on. They fail because `GET /{path...}` is the HTML 404: an undeclared GET under `/app/v1` answers 404, and a 405 `Allow` list includes that GET. The credential is one `hx_app_` token; `DELETE /app/v1/tokens/current` revokes it, so later cases are 401. That is a warning, not a failure.

400 and 428 are accepted on positive cases. A bad cursor, an unknown type and the signup challenge (`428`) are the documented answers, not server errors.

## ZAP baseline

ZAP 2.17.0, 2026-10-30, target `http://127.0.0.1:3000`. `docker pull` of `ghcr.io/zaproxy/zaproxy@sha256:030d3df57735dacbb83464854758e674dd9ad01d79d12a129a8e58b1fd384238` still dies with `short read` on the 625874126-byte layer. The run below is the official `zap-baseline.py` from tag `v2.17.0` (the script inside that image) with the official `ZAP_2.17.0_Crossplatform.zip` from the same release. Flags match `qa/zap/baseline.sh`: `-I -c qa/zap/rules.tsv -J zap-report.json -r zap-report.html -m 5`. `rules.tsv` is still only the comment header.

Result: 1836 URLs, 55 pass, 6 warn, 0 fail, exit 0. The JSON report has no `riskcode` 2 (Medium) or 3 (High).

| Rule | Risk in the report | What it is |
| --- | --- | --- |
| 10096 Timestamp Disclosure | Low | Evidence `1793346180` is 2026-10-30 04:43:00, the scan clock, on `/contact`. |
| 10024 Sensitive information in URL | Informational | The spider requested `GET /login` and `GET /signup` with `?password=ZAP`. The Inertia submit calls `preventDefault` and `form.post`. |
| 10027 Suspicious comments | Informational | Minified `/build` bundles matched `\bADMIN\b` and `\bQUERY\b` inside path and query helpers. |
| 10031 HTML attribute (potential XSS) | Informational | Allowlisted `sort` values (`-score`, `-release_date`) are the `<option value>`. An unknown sort returns 400 `invalid-query` with the value HTML-escaped. |
| 10109 Modern web application | Informational | 404 pages ship the module script and no crawlable links. ZAP says no change is required. |
| 10112 Session management response | Informational | `/library` and `/profile` redirect with cookie `__Host-vukem_session`. ZAP says this is not a vulnerability. |
