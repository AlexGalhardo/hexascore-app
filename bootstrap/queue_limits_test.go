//go:build integration

package bootstrap_test

import (
	"encoding/base64"
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

// The production entry point (no test options) keeps ingestion capped at 2 jobs when QUEUE_LIMITS is unset.
func TestNewDefaultsQueueLimits(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("QUEUE_LIMITS", "")
	t.Setenv("APP_KEY", "base64:"+base64.StdEncoding.EncodeToString(make([]byte, 32)))
	t.Setenv("DATABASE_URL", pgtest.Database(t, ""))

	app, err := bootstrap.New()
	if err != nil {
		t.Fatal(err)
	}

	if got := app.Source.Get("QUEUE_LIMITS"); got != "ingest=2" {
		t.Fatalf("QUEUE_LIMITS = %q, want ingest=2", got)
	}
}
