package normalize

import (
	"regexp"
	"strconv"
	"time"
)

// Release precisions stored in items.release_precision.
const (
	PrecisionDay     = "day"
	PrecisionMonth   = "month"
	PrecisionYear    = "year"
	PrecisionUnknown = "unknown"
)

// Date is a parsed Wikidata time with the precision it was given at.
type Date struct {
	Time      time.Time
	Precision string
}

var timeRe = regexp.MustCompile(`^\+(\d{1,4})-(\d\d)-(\d\d)T`)

// ParseTime reads a trimmed Wikidata time value ({"time","precision"}).
// Month and day are padded to 1 below their precision; precision under year
// (Wikidata 9) and BCE dates do not parse.
func ParseTime(value any) (Date, bool) {
	m, ok := value.(map[string]any)
	if !ok {
		return Date{}, false
	}

	raw, _ := m["time"].(string)
	prec, _ := m["precision"].(float64)
	match := timeRe.FindStringSubmatch(raw)

	if match == nil || prec < 9 {
		return Date{}, false
	}

	year, _ := strconv.Atoi(match[1])
	month, _ := strconv.Atoi(match[2])
	day, _ := strconv.Atoi(match[3])

	precision := PrecisionYear

	switch {
	case prec >= 11:
		precision = PrecisionDay
	case prec >= 10:
		precision = PrecisionMonth
	}

	if prec < 10 || month == 0 {
		month = 1
	}

	if prec < 11 || day == 0 {
		day = 1
	}

	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	// time.Date normalises overflow (Feb 30); the legacy date() raised instead.
	if t.Year() != year || int(t.Month()) != month || t.Day() != day {
		return Date{}, false
	}

	return Date{Time: t, Precision: precision}, true
}

// finer ranks day over month over year, to break a tie on the same date.
func finer(a, b Date) bool {
	rank := map[string]int{PrecisionDay: 3, PrecisionMonth: 2, PrecisionYear: 1}

	return rank[a.Precision] > rank[b.Precision]
}
