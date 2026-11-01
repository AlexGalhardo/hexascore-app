//go:build integration

package seeders_test

import (
	"context"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/database/seeders"
)

func TestCommunityLoad(t *testing.T) {
	db := seededDB(t)
	seeder := &seeders.CommunityLoad{Count: 120}

	for range 2 {
		if err := seeder.Seed(context.Background(), db); err != nil {
			t.Fatal(err)
		}
	}

	for query, want := range map[string]int{
		`SELECT count(*) FROM suggestions WHERE status = 'submitted' AND meta @> '{"load":true}'`: 120,
		`SELECT count(*) FROM suggestions WHERE assigned_to IS NOT NULL`:                          0,
		`SELECT count(DISTINCT author_id) FROM suggestions`:                                       50,
		`SELECT count(*) FROM suggestions WHERE submitted_at IS NULL OR submitted_at > now()`:     0,
		`SELECT count(*) FROM suggestions s WHERE s.kind = 'create' AND s.target_id IS NOT NULL`:  0,
		`SELECT count(*) FROM suggestions WHERE submitted_at < now() - interval '30 hours'`:       0,
	} {
		if got := count(t, db, query); got != want {
			t.Errorf("%s = %d, want %d", query, got, want)
		}
	}
}

func TestCommunityLoadNotRegistered(t *testing.T) {
	for _, s := range seeders.All() {
		if _, ok := s.(*seeders.CommunityLoad); ok {
			t.Fatal("CommunityLoad is in seeders.All(); it is dev-only (bootstrap adds it outside production)")
		}
	}
}
