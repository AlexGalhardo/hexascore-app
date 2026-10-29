package scoring

import "time"

// saoPaulo is the release-gate time zone (SDD 02 §7.4).
var saoPaulo = mustZone("America/Sao_Paulo")

func mustZone(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.FixedZone("BRT", -3*60*60) // no tzdata: Brazil has no DST since 2019
	}
	return loc
}

// EffectiveRelease is the date ratings open (SDD 02 §7.4): day → that day,
// month → last day of the month, year → 31 December; unknown with a date
// behaves as day.
func EffectiveRelease(date time.Time, precision string) time.Time {
	d := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	switch precision {
	case "month":
		return d.AddDate(0, 1, -d.Day())
	case "year":
		return time.Date(d.Year(), 12, 31, 0, 0, 0, 0, time.UTC)
	default:
		return d
	}
}

// RatedBeforeRelease reports whether a rating made at ratedAt falls before
// the item's effective release. Items without a date are gated only while
// announced or cancelled (open issue 02-1).
func RatedBeforeRelease(ratedAt time.Time, release *time.Time, precision, status string) bool {
	if release == nil {
		return status == "announced" || status == "cancelled"
	}
	local := ratedAt.In(saoPaulo)
	day := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC)
	return day.Before(EffectiveRelease(*release, precision))
}
