package run

import (
	"context"
	"strings"

	"github.com/AlexGalhardo/vukem/flags"
)

// Kill switches, code default false (fail closed); production turns them on with FLAG_INGEST_<SOURCE>=true and
// a stored false at /admin/_flags wins within the cache TTL (SDD 06 §6).
var (
	enabled = flags.Bool("ingest_enabled", false).Describe("Master switch of every ingestion job.")

	bySource = func() map[string]flags.Flag[bool] {
		m := map[string]flags.Flag[bool]{}
		for _, s := range []string{"wikidata", "wikipedia", "openlibrary", "tmdb", "igdb", "steam", "gog", "synopsis", "covers", "youtube"} {
			m[s] = flags.Bool("ingest_"+s, false).Describe("Ingestion kill switch for " + s + ".")
		}

		return m
	}()
)

// All lists the flags to register with vukem.WithFlags.
func All() []flags.AnyFlag {
	out := []flags.AnyFlag{enabled}
	for _, f := range bySource {
		out = append(out, f)
	}

	return out
}

// Allowed reports whether a run of source may touch the network: the master switch and the source switch must
// both be on. A source-less job (normalise) needs only the master switch; an unknown source is off.
func Allowed(ctx context.Context, ff *flags.Flags, source string) bool {
	if ff == nil || !ff.Enabled(ctx, enabled) {
		return false
	}

	if source == "" {
		return true
	}

	f, ok := bySource[strings.SplitN(source, ":", 2)[0]]

	return ok && ff.Enabled(ctx, f)
}
