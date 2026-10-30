//go:build integration

package changes_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/api/changes"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func TestRecord(t *testing.T) {
	db := changesDB(t)
	ctx := t.Context()
	id := insertItem(t, db, "game", "portal", "Q1")

	err := db.Tx(ctx, func(tx *database.Tx) error {
		return changes.Record(ctx, tx, changes.Change{
			ItemID: id, ItemType: "game", Action: "updated", Fields: []string{"title", "cover_key"},
		})
	})
	if err != nil {
		t.Fatal(err)
	}

	var fields string
	if err := db.QueryRowContext(ctx, `SELECT array_to_string(fields, ',') FROM api_changes WHERE item_id = $1`, id).
		Scan(&fields); err != nil {
		t.Fatal(err)
	}
	if fields != "title" {
		t.Fatalf("fields = %q, want title", fields)
	}
}

func TestHiddenOnlyChangeWritesNothing(t *testing.T) {
	db := changesDB(t)
	ctx := t.Context()
	id := insertItem(t, db, "movie", "alien", "Q2")

	err := db.Tx(ctx, func(tx *database.Tx) error {
		return changes.Record(ctx, tx, changes.Change{
			ItemID: id, ItemType: "movie", Action: "updated", Fields: []string{"cover_key", "tmdb"},
		})
	})
	if err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM api_changes WHERE item_id = $1`, id).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("rows = %d, want 0", n)
	}
}

func TestScoreChangeOnlyOnDecimal(t *testing.T) {
	same := changes.ScoreFieldNames(
		changes.ScoreSnap{Score: sql.NullFloat64{Float64: 87.4, Valid: true}, RatingCount: 12},
		changes.ScoreSnap{Score: sql.NullFloat64{Float64: 87.49, Valid: true}, RatingCount: 12},
	)
	if len(same) != 0 {
		t.Fatalf("same tenth wrote %v", same)
	}

	moved := changes.ScoreFieldNames(
		changes.ScoreSnap{Score: sql.NullFloat64{Float64: 87.4, Valid: true}, RatingCount: 12,
			Confidence: sql.NullString{String: "low", Valid: true}},
		changes.ScoreSnap{Score: sql.NullFloat64{Float64: 87.6, Valid: true}, RatingCount: 13,
			Confidence: sql.NullString{String: "low", Valid: true}},
	)
	if got := join(moved); got != "score,ratingCount" {
		t.Fatalf("fields = %q, want score,ratingCount", got)
	}
}

func changesDB(t *testing.T) *database.DB {
	t.Helper()
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{
		"DATABASE_URL": pgtest.Database(t, ""),
	}})
	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func insertItem(t *testing.T, db *database.DB, typ, slug, qid string) string {
	t.Helper()
	id := uuid.NewV7().String()
	if _, err := db.ExecContext(context.Background(),
		`INSERT INTO items (id, type, slug, wikidata_qid) VALUES ($1, $2, $3, $4)`, id, typ, slug, qid); err != nil {
		t.Fatal(err)
	}
	return id
}

func join(fields []string) string {
	out := ""
	for i, f := range fields {
		if i > 0 {
			out += ","
		}
		out += f
	}
	return out
}
