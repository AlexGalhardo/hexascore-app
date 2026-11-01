package seeders_test

import (
	"context"
	"errors"
	"testing"
	"testing/fstest"

	"github.com/AlexGalhardo/vukem"

	"github.com/AlexGalhardo/hexascore-app/database/seeders"
)

// The load seeder inserts thousands of fake suggestions: production must refuse it before touching the database.
func TestCommunityLoad_ProductionRefuses(t *testing.T) {
	t.Parallel()

	env := map[string]string{
		"APP_ENV": "production",
		"APP_KEY": "base64:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
	}

	app, err := vukem.New(vukem.WithEnv(fstest.MapFS{}, func(k string) (string, bool) {
		v, ok := env[k]
		return v, ok
	}))
	if err != nil {
		t.Fatal(err)
	}

	err = (&seeders.CommunityLoad{Count: 10}).SeedApp(context.Background(), app)
	if !errors.Is(err, seeders.ErrCommunityLoadProduction) {
		t.Fatalf("SeedApp in production = %v, want ErrCommunityLoadProduction", err)
	}
}
