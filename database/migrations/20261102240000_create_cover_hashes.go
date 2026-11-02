package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261102240000CreateCoverHashes records the content hash (sha256) and the 64-bit difference hash (dHash) of every
// cover object the covers job writes, with the source and licence it came from (RFC-001 section 5 E). The job reads
// it by object key to skip re-encoding the same picture when a second source offers it. One row per key, upserted
// in place, so the table grows with the catalogue and never needs a prune; the row goes away with nothing else (the
// objects are reaped by cover_gc, the hash is a few bytes). No user data.
func m20261102240000CreateCoverHashes() schema.Migration {
	return postgresMigration("20261102240000_create_cover_hashes",
		[]string{`CREATE TABLE cover_hashes (
  key text PRIMARY KEY,
  sha256 bytea NOT NULL,
  dhash bigint NOT NULL,
  width integer NOT NULL,
  height integer NOT NULL,
  bytes integer NOT NULL,
  source text NOT NULL,
  licence text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now()
)`},
		[]string{`DROP TABLE IF EXISTS cover_hashes`},
	)
}
