//go:build integration

// Package scoring_test checks the SDD 02 scoring schema against PostgreSQL.
package scoring_test

import (
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func TestRatingConstraints(t *testing.T) {
	url := pgtest.Database(t, "")
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{"DATABASE_URL": url}})
	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}
	ctx := t.Context()
	sql := db.SQL()

	var userID int64
	if err := sql.QueryRowContext(ctx,
		`INSERT INTO users (name, email, password, created_at, updated_at)
VALUES ('Ana', 'ana@example.com', 'x', now(), now()) RETURNING id`).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	itemID := uuid.NewV7().String()
	if _, err := sql.ExecContext(
		ctx,
		`INSERT INTO items (id, type, slug, wikidata_qid) VALUES ($1, 'game', 'portal', 'Q209207')`,
		itemID,
	); err != nil {
		t.Fatal(err)
	}

	insert := `INSERT INTO ratings (user_id, item_id, c1, c2, c3, c4, c5, c6) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	if _, err := sql.ExecContext(ctx, insert, userID, itemID, 5, 10, 15, 20, 25, 100); err != nil {
		t.Fatalf("valid rating: %v", err)
	}
	var overall float64
	if err := sql.QueryRowContext(ctx, `SELECT overall FROM ratings WHERE user_id = $1 AND item_id = $2`,
		userID, itemID).Scan(&overall); err != nil {
		t.Fatal(err)
	}
	if want := 175.0 / 6; overall < want-1e-9 || overall > want+1e-9 {
		t.Fatalf("overall = %v, want %v (unrounded, D-13)", overall, want)
	}

	other := uuid.NewV7().String()
	if _, err := sql.ExecContext(ctx,
		`INSERT INTO items (id, type, slug, wikidata_qid) VALUES ($1, 'movie', 'heat', 'Q45672')`, other); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name   string
		item   string
		c1     int
		wantIn string
	}{
		{"duplicate user+item", itemID, 50, "duplicate key"},
		{"not a half-star step", other, 52, "ratings_half_star"},
		{"below 0.5 star", other, 0, "ratings_half_star"},
		{"above 5 stars", other, 105, "ratings_half_star"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := sql.ExecContext(ctx, insert, userID, c.item, c.c1, 50, 50, 50, 50, 50)
			if err == nil || !strings.Contains(err.Error(), c.wantIn) {
				t.Fatalf("want %q error, got %v", c.wantIn, err)
			}
		})
	}

	if _, err := sql.ExecContext(ctx, `UPDATE ratings SET overall = 1 WHERE user_id = $1`, userID); err == nil {
		t.Fatal("generated column overall must not be writable")
	}
}
