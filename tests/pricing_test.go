package tests

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func TestPricingPage(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)

	cases := []struct{ path, currency string }{
		{"/pricing", "usd"},
		{"/pt-br/pricing", "brl"},
		{"/es/pricing", "eur"},
		{"/pricing?currency=eur", "eur"},
		{"/pricing?currency=XXX", "usd"},
		{"/pt-br/pricing?currency=XXX", "brl"},
	}

	for _, c := range cases {
		app.Client(t).Get(c.path).AssertOK().AssertInertia("Pricing/Index").
			PropEquals("currency", c.currency).
			HasProp("plans", 3).
			HasProp("coverage")
	}
}

// Owner decision 2026-11-02 (D-33): the page shows the limits the API enforces and the three prices per plan.
func TestPricingPlans(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)

	app.Client(t).Get("/pricing").AssertOK().AssertInertia("Pricing/Index").
		PropEquals("plans.0.id", "free").
		PropEquals("plans.1.id", "hobby").
		PropEquals("plans.1.quota", 10_000).
		PropEquals("plans.1.maxPageSize", 20).
		PropEquals("plans.1.rps", 1).
		PropEquals("plans.1.amounts.usd", 500).
		PropEquals("plans.1.amounts.brl", 2000).
		PropEquals("plans.1.amounts.eur", 500).
		PropEquals("plans.2.id", "pro").
		PropEquals("plans.2.quota", 50_000).
		PropEquals("plans.2.maxPageSize", 50).
		PropEquals("plans.2.rps", 3).
		PropEquals("plans.2.amounts.usd", 1000).
		PropEquals("plans.2.amounts.brl", 5000).
		PropEquals("plans.2.amounts.eur", 1000)
}

// P3.B.3.2, D-33: /pricing and /docs/coverage publish exactly resources/data/api_coverage.json (written from
// qa/sql/api_coverage.sql), in every locale; other docs pages do not carry it.
func TestCoverageRoutes(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("../resources/data/api_coverage.json")
	if err != nil {
		t.Fatal(err)
	}
	var want struct {
		MeasuredAt any `json:"measuredAt"`
		Rows       any `json:"rows"`
	}
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatal(err)
	}

	app := vukemtest.New(t, bootstrap.New)

	for _, prefix := range []string{"", "/pt-br", "/es"} {
		app.Client(t).Get(prefix+"/pricing").AssertOK().AssertInertia("Pricing/Index").
			PropEquals("coverage.docsPath", "/docs/coverage").
			PropEquals("coverage.measuredAt", want.MeasuredAt).
			PropEquals("coverage.rows", want.Rows)
		app.Client(t).Get(prefix+"/docs/coverage").AssertOK().AssertInertia("Docs/Page").
			PropEquals("slug", "coverage").
			PropEquals("coverage.measuredAt", want.MeasuredAt).
			PropEquals("coverage.rows", want.Rows)
		app.Client(t).Get(prefix+"/docs/changelog").AssertOK().AssertInertia("Docs/Page").
			PropEquals("slug", "changelog").
			MissingProp("coverage")
	}
}
