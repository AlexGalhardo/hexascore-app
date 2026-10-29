# Changelog

The version is part of the path. v1 only changes in backwards-compatible ways.

## Versioning policy

Within `/v1` we only make additive changes: new endpoints, new optional parameters, new response fields and new error `detail` texts. Ignore unknown fields and do not rely on field order.

Breaking changes (removing or renaming a field, changing a type, making a parameter required) ship as a new path (`/v2`). Error `code` values, category keys and sort names are part of the contract.

## Deprecation and sunset

- A deprecated version keeps working for at least 6 months after its successor is released.
- Its responses carry `Deprecation` (RFC 9745) and `Sunset` (RFC 8594) headers with the date it stops working, plus a `Link` to this page.
- Account owners with recent traffic on the old version are emailed at deprecation and 30 days before sunset.

## Changes

### v1: first release

- Lists, details and genres for `games`, `movies`, `series` and `books`.
- `/search`, `/lookup` by external id and `/me/usage`.
