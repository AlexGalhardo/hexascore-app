package trending

import "github.com/AlexGalhardo/vukem/flags"

// V2Flag switches the landing page between the RFC-002 snapshot (on, default) and the earlier most-rated list (off,
// the rollback of RFC-002 section 8). Env FLAG_TRENDING_V2.
var V2Flag = flags.Bool("trending_v2", true).
	Describe("Landing-page trending from the 7-day snapshot (RFC-002); off restores the most-rated list")
