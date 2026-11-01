package eval_test

import (
	"bytes"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/hexascore-app/app/recommend/eval"
)

var day0 = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

func ratings(user string, n int) []eval.Rating {
	out := make([]eval.Rating, n)
	for i := range n {
		// Reverse order on purpose: the split must sort by time, not trust the input order.
		out[n-1-i] = eval.Rating{User: user, Item: user + string(rune('a'+i)), Overall: 80, At: day0.AddDate(0, 0, i)}
	}

	return out
}

func TestHoldOutSplit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		n           int
		train, test int
	}{
		{"10 ratings: 20 % is 2", 10, 8, 2},
		{"11 ratings: ceil(2.2) is 3", 11, 8, 3},
		{"4 ratings: min 2", 4, 2, 2},
		{"3 ratings: min 2 leaves 1 to train", 3, 1, 2},
		{"2 ratings: nothing to train on, user skipped", 2, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			train, test := eval.HoldOut(ratings("u", tt.n))
			if len(train) != tt.train || len(test) != tt.test {
				t.Fatalf("train/test = %d/%d, want %d/%d", len(train), len(test), tt.train, tt.test)
			}

			for _, tr := range train {
				for _, te := range test {
					if tr.At.After(te.At) {
						t.Fatalf("train rating %v is newer than test rating %v", tr.At, te.At)
					}
				}
			}
		})
	}
}

func TestPositives(t *testing.T) {
	t.Parallel()

	got := eval.Positives(
		[]eval.Rating{{Item: "a", Overall: 70}, {Item: "b", Overall: 69.9}, {Item: "c", Overall: 100}},
	)
	if len(got) != 2 || !got["a"] || !got["c"] {
		t.Fatalf("positives = %v, want a and c (overall >= 70)", got)
	}
}

func TestPrecisionAt10(t *testing.T) {
	t.Parallel()

	pos := map[string]bool{"a": true, "c": true, "z": true}
	recs := []string{"a", "x", "c", "y", "w", "v", "u", "t", "s", "r", "z"}

	// z is rank 11 and does not count.
	if got := eval.PrecisionAt(recs, pos, 10); got != 0.2 {
		t.Fatalf("P@10 = %v, want 0.2", got)
	}

	// A short list is still divided by k.
	if got := eval.PrecisionAt([]string{"a", "c"}, pos, 10); got != 0.2 {
		t.Fatalf("short list P@10 = %v, want 0.2", got)
	}
}

func TestNDCGAt10(t *testing.T) {
	t.Parallel()

	pos := map[string]bool{"a": true, "b": true, "c": true}
	got := eval.NDCGAt([]string{"a", "x", "b", "y"}, pos, 10)
	want := (1 + 1/math.Log2(4)) / (1 + 1/math.Log2(3) + 1/math.Log2(4))

	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("nDCG@10 = %v, want %v", got, want)
	}

	if got := eval.NDCGAt([]string{"a", "b", "c"}, pos, 10); math.Abs(got-1) > 1e-12 {
		t.Fatalf("ideal order nDCG = %v, want 1", got)
	}

	if got := eval.NDCGAt([]string{"a"}, nil, 10); got != 0 {
		t.Fatalf("no positives nDCG = %v, want 0", got)
	}
}

func TestBootstrapCI(t *testing.T) {
	t.Parallel()

	m, lo, hi := eval.BootstrapCI([]float64{0.3, 0.3, 0.3, 0.3}, 1000, 1)
	if math.Abs(m-0.3) > 1e-12 || math.Abs(lo-0.3) > 1e-12 || math.Abs(hi-0.3) > 1e-12 {
		t.Fatalf("constant sample: %v [%v, %v]", m, lo, hi)
	}

	vals := []float64{0, 1, 0, 1, 1, 0, 1, 1, 0, 1}
	m, lo, hi = eval.BootstrapCI(vals, 2000, 7)
	m2, lo2, hi2 := eval.BootstrapCI(vals, 2000, 7)

	if m != 0.6 || lo >= m || hi <= m || lo < 0 || hi > 1 {
		t.Fatalf("mean/CI = %v [%v, %v]", m, lo, hi)
	}

	if m != m2 || lo != lo2 || hi != hi2 {
		t.Fatal("same seed must give the same interval")
	}

	// A paired lift of +0.1 on every user excludes 0.
	if _, lo, _ := eval.BootstrapCI([]float64{0.1, 0.1, 0.1, 0.1, 0.1}, 500, 3); lo <= 0 {
		t.Fatalf("lo = %v, want > 0", lo)
	}
}

func TestEvaluateAndCSV(t *testing.T) {
	t.Parallel()

	cases := []eval.UserCase{
		{Recs: []string{"a", "b"}, Positives: map[string]bool{"a": true}},
		{Recs: []string{"b", "c"}, Positives: map[string]bool{"z": true}},
		{Recs: nil, Positives: map[string]bool{"a": true}},
	}

	rep := eval.Evaluate(cases, 10, 100, 1)
	if rep.Users != 3 || rep.Precision.Mean != (0.1+0+0)/3 {
		t.Fatalf("report = %+v", rep)
	}

	// 3 distinct recommended items out of a catalogue of 10.
	if got := eval.Coverage(cases, 10); got != 0.3 {
		t.Fatalf("coverage = %v, want 0.3", got)
	}

	var buf bytes.Buffer
	if err := eval.WriteCSV(
		&buf,
		[][]string{{"metric", "value"}, {"=cmd", "+cmd"}, {"@x", "-cmd"}, {"ok", "0.1"}, {"neg", "-2"}},
	); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	for _, bad := range []string{"\n=cmd", ",+cmd", "\n@x", ",-cmd"} {
		if strings.Contains(out, bad) {
			t.Fatalf("formula cell not guarded (%q) in:\n%s", bad, out)
		}
	}

	if !strings.Contains(out, "'=cmd") || !strings.Contains(out, "ok,0.1") || !strings.Contains(out, "neg,-2") {
		t.Fatalf("csv = %s", out)
	}
}
