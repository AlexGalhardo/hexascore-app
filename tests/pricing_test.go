package tests

import (
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
			HasProp("plans").
			HasProp("coverage")
	}
}
