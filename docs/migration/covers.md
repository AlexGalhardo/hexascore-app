# Legacy cover copy

`scripts/legacy/blob-export.ts` (Bun, no dependency: `fetch`, `Bun.SQL`, `Bun.S3Client`) copies the covers the legacy database references from Vercel Blob to the S3 bucket, under the legacy key `covers/<id>-<hash16>.webp` (and `-t.webp` thumbnails) that the `catalog` import stage writes into `items.cover_key` (SDD 01 §5.4, SDD 11 §7.6).

## Run

The bucket credentials come from `S3_ENDPOINT`, `S3_REGION`, `S3_ACCESS_KEY_ID`, `S3_SECRET_ACCESS_KEY` and `S3_USE_PATH_STYLE`, as for the app. `NEON_RO_URL` is the owner's read-only role (P2.F.1.1), passed through the environment only.

```sh
# T-7 d: full copy
bun scripts/legacy/blob-export.ts --db "$NEON_RO_URL" --bucket <bucket> --manifest manifest.ndjson
# window step 5: covers of items updated since the full copy started
bun scripts/legacy/blob-export.ts --delta --since <ISO time> --db "$NEON_RO_URL" --bucket <bucket> --manifest manifest.ndjson
```

- Only `items.meta.coverUrl` and `coverThumbUrl` values that are Blob URLs are copied. Source URLs stored without the Blob (`covers.py` without a token) are counted as `notBlob` and get no `cover_key`; the cover job re-fetches them (SDD 06).
- Orphans (Blob objects no item references) are never copied. Their count is the Blob listing minus `referenced`, taken by the owner with the Blob token for the decommission (P2.G.4.3).
- The manifest gets one `{legacyUrl, key, sha256}` line per copied object. A key already in the manifest and present in the bucket is skipped without a download, so a re-run copies 0; a key in the manifest but missing from the bucket is copied again.
- The script prints `{referenced, notBlob, copied, skipped, failed}` and exits 1 when any copy failed (re-run it; copied keys are skipped).

## Counts

| Run | Referenced | notBlob | Copied | Skipped | Failed | Orphans skipped |
| --- | --- | --- | --- | --- | --- | --- |
| Unit test (`scripts/legacy/blob-export.test.ts`, first run then re-run) | 2 | 1 | 2, then 0 | 0, then 2 | 0 | n/a |
| Local smoke (scratch DB, rustfs path-style) | 0 | 2 | 0 | 0 | 0 | n/a |
| Staging rehearsal | owner-gated | | | | | |
| Production T-7 d and delta | owner-gated | | | | | |

`Bun.S3Client` with `virtualHostedStyle: false` (from `S3_USE_PATH_STYLE=true`) was verified against the local RustFS (write, exists, delete); the Railway bucket's addressing style is still to be checked on staging.
