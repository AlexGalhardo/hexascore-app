package scoring

import (
	"math"
	"math/rand" //nolint:depguard // testing/quick requires *math/rand.Rand; fixtures, not secrets
	"reflect"
	"testing"
	"testing/quick"
	"time"
)

func flat(overall, weight float64) Rating {
	var c [Categories]float64
	for k := range c {
		c[k] = overall
	}
	return Rating{C: c, Weight: weight}
}

func prior(c float64) Params {
	p := Params{M: 10}
	for k := range p.Prior {
		p.Prior[k] = c
	}
	return p
}

// TestFormulaGolden is the SDD 02 §7.1 worked example.
func TestFormulaGolden(t *testing.T) {
	w5 := Weight(Factors{AccountAge: 3 * 24 * time.Hour, Shadow: 1}) // 0.25 · 0.7
	if math.Abs(w5-0.175) > 1e-12 {
		t.Fatalf("weight of a 3-day account without status = %v, want 0.175", w5)
	}
	ratings := []Rating{flat(90, 1), flat(85, 1), flat(80, 0.85), flat(95, 0.7), flat(100, w5), flat(75, 1)}
	r := Compute(ratings, prior(68))

	if math.Abs(r.B-73.4805) > 1e-4 {
		t.Fatalf("B = %.6f, want 73.4805", r.B)
	}
	if stars, visible := r.Stars(); stars != 7.3 || !visible {
		t.Fatalf("stars = %v visible=%v, want 7.3 true", stars, visible)
	}
	if r.N != 6 || math.Abs(r.W-4.725) > 1e-9 || r.Confidence != ConfidenceLow {
		t.Fatalf("N=%d W=%v confidence=%q", r.N, r.W, r.Confidence)
	}
	if math.Abs(r.SE-3.38) > 0.01 {
		t.Fatalf("SE = %.4f, want ≈ 3.38", r.SE)
	}
}

func TestWeightFactors(t *testing.T) {
	year := 400 * 24 * time.Hour
	cases := []struct {
		name string
		f    Factors
		want float64
	}{
		{
			"completed game with time-to-beat",
			Factors{AccountAge: year, Status: "completed", IsGame: true, HasTimeToBeat: true, Shadow: 1},
			1,
		},
		{
			"completed game without time-to-beat",
			Factors{AccountAge: year, Status: "completed", IsGame: true, Shadow: 1},
			0.9,
		},
		{
			"in progress, 10-day account",
			Factors{AccountAge: 10 * 24 * time.Hour, Status: "in_progress", Shadow: 1},
			0.5 * 0.85,
		},
		{"want, 60-day account", Factors{AccountAge: 60 * 24 * time.Hour, Status: "want", Shadow: 1}, 0.75 * 0.7},
		{"shadow downweight", Factors{AccountAge: year, Status: "completed", Shadow: 0.5}, 0.5},
		{"banned", Factors{AccountAge: year, Status: "completed", Shadow: 1, Banned: true}, 0},
		{"before release", Factors{AccountAge: year, Status: "completed", Shadow: 1, BeforeRelease: true}, 0},
	}
	for _, c := range cases {
		if got := Weight(c.f); math.Abs(got-c.want) > 1e-12 {
			t.Errorf("%s: %v, want %v", c.name, got, c.want)
		}
	}
}

// fixture is a random item: 1–40 ratings and a prior, generated for quick.
type fixture struct {
	Ratings []Rating
	P       Params
}

func step(r *rand.Rand) float64 { return float64(5 * (1 + r.Intn(20))) }

// Generate implements quick.Generator.
func (fixture) Generate(r *rand.Rand, _ int) reflect.Value {
	f := fixture{P: Params{M: float64(5 + r.Intn(46))}}
	for k := range f.P.Prior {
		f.P.Prior[k] = 5 + r.Float64()*95
	}
	for range 1 + r.Intn(40) {
		var c [Categories]float64
		for k := range c {
			c[k] = step(r)
		}
		f.Ratings = append(f.Ratings, Rating{C: c, Weight: 0.05 + r.Float64()*0.95})
	}
	return reflect.ValueOf(f)
}

func TestFormulaProperties(t *testing.T) {
	const eps = 1e-9
	cfg := &quick.Config{MaxCount: 10000, Rand: rand.New(rand.NewSource(1))}

	props := map[string]func(fixture) bool{
		"B in [5,100] and between prior and raw mean": func(f fixture) bool {
			r := Compute(f.Ratings, f.P)
			var sw, swo, c float64
			for _, x := range f.Ratings {
				sw += x.Weight
				swo += x.Weight * x.Overall()
			}
			for _, p := range f.P.Prior {
				c += p
			}
			c /= Categories
			lo, hi := math.Min(c, swo/sw), math.Max(c, swo/sw)
			return r.B >= 5-eps && r.B <= 100+eps && r.B >= lo-eps && r.B <= hi+eps
		},
		"raising a category never lowers B_k or B": func(f fixture) bool {
			before := Compute(f.Ratings, f.P)
			raised := append([]Rating(nil), f.Ratings...)
			k := len(raised[0].C) - 1
			if raised[0].C[k] >= 100 {
				return true
			}
			raised[0].C[k] += 5
			after := Compute(raised, f.P)
			return after.BK[k] >= before.BK[k]-eps && after.B >= before.B-eps
		},
		"a rating above B raises B, below lowers it": func(f fixture) bool {
			before := Compute(f.Ratings, f.P)
			hi, lo := flat(100, 1), flat(5, 1)
			up := Compute(append(append([]Rating(nil), f.Ratings...), hi), f.P)
			down := Compute(append(append([]Rating(nil), f.Ratings...), lo), f.P)
			return (hi.Overall() <= before.B || up.B > before.B) && (lo.Overall() >= before.B || down.B < before.B)
		},
		"raising a weight moves B toward that rating": func(f fixture) bool {
			before := Compute(f.Ratings, f.P)
			heavier := append([]Rating(nil), f.Ratings...)
			heavier[0].Weight += 1
			after := Compute(heavier, f.P)
			o := heavier[0].Overall()
			return math.Abs(after.B-o) <= math.Abs(before.B-o)+eps
		},
	}
	for name, prop := range props {
		if err := quick.Check(prop, cfg); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}
