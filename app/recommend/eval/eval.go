// Package eval is the offline evaluation protocol of the recommenders (SDD 08 section 7.6): a temporal hold-out,
// precision@k and nDCG@k with a bootstrap confidence interval, coverage, and the monthly collaborative-filtering
// trigger check.
package eval

import (
	"cmp"
	"context"
	"fmt"
	"io"
	"math"
	"math/rand/v2" //nolint:depguard // seeded, reproducible resampling; nothing security-sensitive
	"slices"
	"strconv"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/export"
)

const (
	// HoldOutShare is the share of each user's latest ratings that is held out.
	HoldOutShare = 0.2
	// MinHeldOut is the smallest number of held-out ratings per user.
	MinHeldOut = 2
	// PositiveThreshold is the overall rating from which a held-out item counts as liked.
	PositiveThreshold = 70.0
	// K is the list length the metrics look at.
	K = 10
	// CFMinUsers is the number of users with at least CFMinRatings ratings that triggers the CF comparison.
	CFMinUsers = 5000
	// CFMinRatings is the ratings a user needs to count for the trigger.
	CFMinRatings = 10
)

// Rating is one rating of a user for the split.
type Rating struct {
	User    string
	Item    string
	Overall float64
	At      time.Time
}

// HoldOut splits one user's ratings by time: the latest max(MinHeldOut, ceil(20 %)) go to test, the rest to train.
// A user left with nothing to train on (fewer than MinHeldOut+1 ratings) yields no split.
func HoldOut(rs []Rating) (train, test []Rating) {
	s := slices.Clone(rs)
	slices.SortFunc(s, func(a, b Rating) int { return cmp.Or(a.At.Compare(b.At), cmp.Compare(a.Item, b.Item)) })

	k := max(MinHeldOut, int(math.Ceil(HoldOutShare*float64(len(s))-1e-9)))
	if len(s)-k < 1 {
		return nil, nil
	}

	return s[:len(s)-k], s[len(s)-k:]
}

// Positives are the held-out items rated at least PositiveThreshold.
func Positives(test []Rating) map[string]bool {
	out := map[string]bool{}

	for _, r := range test {
		if r.Overall >= PositiveThreshold {
			out[r.Item] = true
		}
	}

	return out
}

// PrecisionAt is the share of the first k recommendations that are positives; a short list is still divided by k.
func PrecisionAt(recs []string, positives map[string]bool, k int) float64 {
	var hits int

	for _, id := range recs[:min(k, len(recs))] {
		if positives[id] {
			hits++
		}
	}

	return float64(hits) / float64(k)
}

// NDCGAt is the binary-relevance nDCG of the first k recommendations; 0 when the user has no positives.
func NDCGAt(recs []string, positives map[string]bool, k int) float64 {
	var dcg, ideal float64

	for i, id := range recs[:min(k, len(recs))] {
		if positives[id] {
			dcg += 1 / math.Log2(float64(i)+2)
		}
	}

	for i := range min(k, len(positives)) {
		ideal += 1 / math.Log2(float64(i)+2)
	}

	if ideal == 0 {
		return 0
	}

	return dcg / ideal
}

// BootstrapCI returns the mean of values and its 95 % percentile bootstrap interval from iters resamples. The seed
// makes the interval reproducible. For a lift, pass the paired per-user differences.
func BootstrapCI(values []float64, iters int, seed uint64) (mean, lo, hi float64) {
	if len(values) == 0 {
		return 0, 0, 0
	}

	mean = avg(values)
	//nolint:gosec // seeded on purpose so a report is reproducible
	rng := rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
	means := make([]float64, iters)
	sample := make([]float64, len(values))

	for i := range means {
		for j := range sample {
			sample[j] = values[rng.IntN(len(values))]
		}

		means[i] = avg(sample)
	}

	slices.Sort(means)

	return mean, means[int(0.025*float64(iters))], means[min(iters-1, int(0.975*float64(iters)))]
}

func avg(v []float64) float64 {
	var s float64
	for _, x := range v {
		s += x
	}

	return s / float64(len(v))
}

// UserCase is one evaluated user: the list a recommender produced from the training ratings, and the held-out positives.
type UserCase struct {
	Recs      []string
	Positives map[string]bool
}

// Metric is a mean with its 95 % interval.
type Metric struct{ Mean, Lo, Hi float64 }

// Report is the outcome of one evaluation.
type Report struct {
	Users     int
	Precision Metric
	NDCG      Metric
}

// Evaluate computes precision@k and nDCG@k over the users, each with a bootstrap interval.
func Evaluate(cases []UserCase, k, iters int, seed uint64) Report {
	p := make([]float64, len(cases))
	n := make([]float64, len(cases))

	for i, c := range cases {
		p[i], n[i] = PrecisionAt(c.Recs, c.Positives, k), NDCGAt(c.Recs, c.Positives, k)
	}

	rep := Report{Users: len(cases)}
	rep.Precision.Mean, rep.Precision.Lo, rep.Precision.Hi = BootstrapCI(p, iters, seed)
	rep.NDCG.Mean, rep.NDCG.Lo, rep.NDCG.Hi = BootstrapCI(n, iters, seed+1)

	return rep
}

// Coverage is the share of the catalogue that appears in at least one user's list.
func Coverage(cases []UserCase, catalogue int) float64 {
	seen := map[string]bool{}

	for _, c := range cases {
		for _, id := range c.Recs {
			seen[id] = true
		}
	}

	if catalogue == 0 {
		return 0
	}

	return float64(len(seen)) / float64(catalogue)
}

// WriteCSV writes rows with export.EncodeCSV, whose SafeCell guard keeps a spreadsheet from reading a cell as a
// formula; plain numbers stay as they are.
func WriteCSV(w io.Writer, rows [][]string) error {
	all := func(yield func([]string, error) bool) {
		for _, r := range rows {
			if !yield(r, nil) {
				return
			}
		}
	}

	return export.EncodeCSV(w, nil, all, func(r []string) []string { return r })
}

// Rows renders a report as metric, value, ci_low, ci_high rows with a header.
func (r Report) Rows(coverage float64) [][]string {
	f := func(v float64) string { return fmt.Sprintf("%.4f", v) }

	return [][]string{
		{"metric", "value", "ci_low", "ci_high"},
		{"users", strconv.Itoa(r.Users), "", ""},
		{"precision@10", f(r.Precision.Mean), f(r.Precision.Lo), f(r.Precision.Hi)},
		{"ndcg@10", f(r.NDCG.Mean), f(r.NDCG.Lo), f(r.NDCG.Hi)},
		{"coverage", f(coverage), "", ""},
	}
}

// Trigger is the monthly CF gate.
type Trigger struct {
	Users     int
	Triggered bool
	At        time.Time
}

// Thresholds override the CF trigger for tests; the zero value is the SDD one.
type Thresholds struct{ MinUsers, MinRatings int }

// CheckTrigger counts the users with at least MinRatings ratings. Below MinUsers it computes nothing else.
func CheckTrigger(ctx context.Context, db *database.DB, now time.Time, th Thresholds) (Trigger, error) {
	if th.MinUsers == 0 {
		th.MinUsers = CFMinUsers
	}

	if th.MinRatings == 0 {
		th.MinRatings = CFMinRatings
	}

	t := Trigger{At: now.UTC()}

	err := db.QueryRowContext(ctx,
		`SELECT count(*) FROM (SELECT user_id FROM ratings GROUP BY user_id HAVING count(*) >= $1) x`,
		th.MinRatings).Scan(&t.Users)
	t.Triggered = t.Users >= th.MinUsers

	return t, err
}

// TriggerRows renders the trigger check as report rows.
func (t Trigger) TriggerRows() [][]string {
	return [][]string{
		{"metric", "value"},
		{"checked_at", t.At.Format(time.RFC3339)},
		{"users_with_10_ratings", strconv.Itoa(t.Users)},
		{"triggered", strconv.FormatBool(t.Triggered)},
	}
}

// RunMonthly is the monthly check: it records the trigger and nothing else. When the gate is met the report says so;
// the comparison itself (hold-out of both recommenders through Evaluate) is run by hand against the
// baseline recorded at launch + 90 days, see docs/reco-eval/README.md.
func RunMonthly(ctx context.Context, db *database.DB, now time.Time, th Thresholds, w io.Writer) (Trigger, error) {
	t, err := CheckTrigger(ctx, db, now, th)
	if err != nil {
		return t, err
	}

	rows := t.TriggerRows()
	if t.Triggered {
		rows = append(rows, []string{"next_step", "run the hold-out comparison (docs/reco-eval/README.md)"})
	}

	return t, WriteCSV(w, rows)
}
