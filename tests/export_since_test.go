//go:build integration

package tests

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func TestExportSince_ReplayIntoLegacyShape(t *testing.T) {
	since := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	before := since.Add(-2 * time.Hour)
	after := since.Add(2 * time.Hour)

	var out bytes.Buffer

	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{
		Env:        map[string]string{"DATABASE_URL": pgtest.Database(t, "")},
		AppOptions: []vukem.Option{vukem.WithStdout(&out)},
	})

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	ctx := t.Context()
	sqlDB := db.SQL()
	changed := uuid.NewV7().String()
	untouched := uuid.NewV7().String()
	listed := uuid.NewV7().String()
	freshItem := uuid.NewV7().String()
	listID := uuid.NewV7().String()
	keyID := uuid.NewV7().String()

	if _, err := sqlDB.ExecContext(ctx, `
INSERT INTO items (id, type, slug, wikidata_qid) VALUES
  ($1, 'game', 'export-changed', 'Q1'),
  ($2, 'game', 'export-old', 'Q2'),
  ($3, 'game', 'export-listed', 'Q3'),
  ($4, 'game', 'export-fresh', 'Q4')`, changed, untouched, listed, freshItem); err != nil {
		t.Fatal(err)
	}

	var legacyUser, freshUser int64

	if err := sqlDB.QueryRowContext(ctx, `
INSERT INTO users (name, email, password, username, legacy_id, created_at, updated_at)
VALUES ('Legacy Ana', 'legacy-ana@example.test', 'argon2id-hash', 'lega', 'leg-a', $1, $1)
RETURNING id`, before).Scan(&legacyUser); err != nil {
		t.Fatal(err)
	}

	if err := sqlDB.QueryRowContext(ctx, `
INSERT INTO users (name, email, password, username, created_at, updated_at)
VALUES ('New Bea', 'new-bea@example.test', 'argon2id-hash', 'newb', $1, $1)
RETURNING id`, after).Scan(&freshUser); err != nil {
		t.Fatal(err)
	}

	insertRating := `INSERT INTO ratings (user_id, item_id, c1, c2, c3, c4, c5, c6, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	if _, err := sqlDB.ExecContext(
		ctx,
		insertRating,
		legacyUser,
		changed,
		80,
		85,
		90,
		95,
		100,
		75,
		before,
		after,
	); err != nil {
		t.Fatal(err)
	}

	if _, err := sqlDB.ExecContext(
		ctx,
		insertRating,
		legacyUser,
		untouched,
		20,
		20,
		20,
		20,
		20,
		20,
		before,
		before,
	); err != nil {
		t.Fatal(err)
	}

	if _, err := sqlDB.ExecContext(
		ctx,
		insertRating,
		freshUser,
		freshItem,
		70,
		70,
		70,
		70,
		70,
		70,
		after,
		after,
	); err != nil {
		t.Fatal(err)
	}

	if _, err := sqlDB.ExecContext(ctx, `
INSERT INTO item_status (user_id, item_id, status, updated_at) VALUES ($1, $2, 'completed', $3)`,
		legacyUser, changed, after); err != nil {
		t.Fatal(err)
	}

	if _, err := sqlDB.ExecContext(ctx, `
INSERT INTO time_to_beat (user_id, item_id, main_minutes, created_at, updated_at) VALUES ($1, $2, 90, $3, $3)`,
		legacyUser, changed, after); err != nil {
		t.Fatal(err)
	}

	if _, err := sqlDB.ExecContext(ctx, `
INSERT INTO lists (id, user_id, type, title, is_public, created_at, updated_at)
VALUES ($1, $2, 'game', 'After cutover', true, $3, $3)`, listID, legacyUser, after); err != nil {
		t.Fatal(err)
	}

	if _, err := sqlDB.ExecContext(ctx, `
INSERT INTO list_items (list_id, item_id, position, added_at) VALUES ($1, $2, 1, $3)`,
		listID, listed, after); err != nil {
		t.Fatal(err)
	}

	hash := strings.Repeat("ab", 32)

	if _, err := sqlDB.ExecContext(ctx, `
INSERT INTO api_keys (id, user_id, name, prefix, hash, created_at) VALUES ($1, $2, 'replay', 'hx_live_test', $3, $4)`,
		keyID, legacyUser, hash, after); err != nil {
		t.Fatal(err)
	}

	conn, err := sqlDB.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = conn.Close() })

	ddl := []string{
		`CREATE SCHEMA legacy_shape`,
		`CREATE TYPE legacy_shape.item_type AS ENUM ('game','movie','series','book')`,
		`CREATE TYPE legacy_shape.item_status_value AS ENUM ('want','in_progress','completed','dropped')`,
		`CREATE TABLE legacy_shape."user" (
  id text PRIMARY KEY,
  name text NOT NULL,
  email text NOT NULL UNIQUE,
  email_verified boolean NOT NULL DEFAULT false,
  username text,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
)`,
		`CREATE TABLE legacy_shape.verification (
  id text PRIMARY KEY,
  identifier text NOT NULL,
  value text NOT NULL,
  expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
)`,
		`CREATE TABLE legacy_shape.api_keys (
  id uuid PRIMARY KEY,
  user_id text NOT NULL,
  name text NOT NULL,
  prefix text NOT NULL,
  hash text NOT NULL UNIQUE,
  created_at timestamptz NOT NULL
)`,
		`CREATE TABLE legacy_shape.ratings (
  user_id text NOT NULL,
  item_id uuid NOT NULL,
  c1 smallint NOT NULL, c2 smallint NOT NULL, c3 smallint NOT NULL,
  c4 smallint NOT NULL, c5 smallint NOT NULL, c6 smallint NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  PRIMARY KEY (user_id, item_id)
)`,
		`CREATE TABLE legacy_shape.item_status (
  user_id text NOT NULL,
  item_id uuid NOT NULL,
  status legacy_shape.item_status_value NOT NULL,
  updated_at timestamptz NOT NULL,
  PRIMARY KEY (user_id, item_id)
)`,
		`CREATE TABLE legacy_shape.time_to_beat (
  user_id text NOT NULL,
  item_id uuid NOT NULL,
  main_minutes int, extras_minutes int, completionist_minutes int,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  PRIMARY KEY (user_id, item_id)
)`,
		`CREATE TABLE legacy_shape.lists (
  id uuid PRIMARY KEY,
  user_id text NOT NULL,
  type legacy_shape.item_type NOT NULL,
  title text NOT NULL,
  description text,
  is_public boolean NOT NULL,
  hidden boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
)`,
		`CREATE TABLE legacy_shape.list_items (
  list_id uuid NOT NULL,
  item_id uuid NOT NULL,
  position int NOT NULL,
  note text,
  PRIMARY KEY (list_id, item_id)
)`,
		`SET search_path TO legacy_shape`,
	}

	for _, stmt := range ddl {
		if _, err := conn.ExecContext(ctx, stmt); err != nil {
			t.Fatalf("ddl: %v", err)
		}
	}

	if _, err := conn.ExecContext(ctx, `
INSERT INTO "user" (id, name, email, email_verified, username, created_at, updated_at)
VALUES ('leg-a', 'Legacy Ana', 'legacy-ana@example.test', true, 'lega', $1, $1)`, before); err != nil {
		t.Fatal(err)
	}

	if _, err := conn.ExecContext(ctx, `
INSERT INTO ratings (user_id, item_id, c1, c2, c3, c4, c5, c6, created_at, updated_at)
VALUES ('leg-a', $1, 10, 10, 10, 10, 10, 10, $2, $2),
       ('leg-a', $3, 20, 20, 20, 20, 20, 20, $2, $2)`, changed, before, untouched); err != nil {
		t.Fatal(err)
	}

	if err := app.Call(ctx, "hexascore:export-since", []string{"--since", since.Format(time.RFC3339)}); err != nil {
		t.Fatal(err)
	}

	replay(t, ctx, conn, out.String())

	var c1, c2, c3, c4, c5, c6 int

	if err := conn.QueryRowContext(ctx, `SELECT c1,c2,c3,c4,c5,c6 FROM ratings WHERE user_id = 'leg-a' AND item_id = $1`, changed).
		Scan(&c1, &c2, &c3, &c4, &c5, &c6); err != nil {
		t.Fatal(err)
	}

	if c1 != 80 || c2 != 85 || c3 != 90 || c4 != 95 || c5 != 100 || c6 != 75 {
		t.Fatalf("replayed rating = %d %d %d %d %d %d", c1, c2, c3, c4, c5, c6)
	}

	if err := conn.QueryRowContext(ctx, `SELECT c1 FROM ratings WHERE user_id = 'leg-a' AND item_id = $1`, untouched).
		Scan(&c1); err != nil {
		t.Fatal(err)
	}

	if c1 != 20 {
		t.Fatalf("rating from before the cutover = %d, want 20", c1)
	}

	var status string

	if err := conn.QueryRowContext(ctx, `SELECT status::text FROM item_status WHERE user_id = 'leg-a' AND item_id = $1`, changed).
		Scan(&status); err != nil {
		t.Fatal(err)
	}

	if status != "completed" {
		t.Fatalf("status = %s", status)
	}

	var minutes int

	if err := conn.QueryRowContext(ctx, `SELECT main_minutes FROM time_to_beat WHERE user_id = 'leg-a' AND item_id = $1`, changed).
		Scan(&minutes); err != nil {
		t.Fatal(err)
	}

	if minutes != 90 {
		t.Fatalf("main_minutes = %d", minutes)
	}

	var title, listUser string

	if err := conn.QueryRowContext(ctx, `
SELECT l.title, l.user_id FROM lists l JOIN list_items li ON li.list_id = l.id WHERE l.id = $1 AND li.item_id = $2`,
		listID, listed).Scan(&title, &listUser); err != nil {
		t.Fatal(err)
	}

	if title != "After cutover" || listUser != "leg-a" {
		t.Fatalf("list = %s user %s", title, listUser)
	}

	var keyUser string

	if err := conn.QueryRowContext(ctx, `SELECT user_id FROM api_keys WHERE hash = $1`, hash).
		Scan(&keyUser); err != nil {
		t.Fatal(err)
	}

	if keyUser != "leg-a" {
		t.Fatalf("api key user = %s", keyUser)
	}

	freshID := fmt.Sprintf("hs:%d", freshUser)
	var resetValue, freshName string

	if err := conn.QueryRowContext(ctx, `
SELECT u.name, v.value FROM "user" u JOIN verification v ON v.id = 'reset:' || $1
WHERE u.id = $1`, freshID).Scan(&freshName, &resetValue); err != nil {
		t.Fatalf("new account: %v\n%s", err, out.String())
	}

	if freshName != "New Bea" || resetValue != "required" {
		t.Fatalf("new account %s reset %s", freshName, resetValue)
	}

	if err := conn.QueryRowContext(ctx, `SELECT c1 FROM ratings WHERE user_id = $1 AND item_id = $2`, freshID, freshItem).
		Scan(&c1); err != nil {
		t.Fatal(err)
	}

	if c1 != 70 {
		t.Fatalf("new account rating = %d", c1)
	}
}

func replay(t *testing.T, ctx context.Context, conn *sql.Conn, script string) {
	t.Helper()

	for _, stmt := range strings.Split(script, "\n") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}

		if _, err := conn.ExecContext(ctx, stmt); err != nil {
			t.Fatalf("replay %s: %v", stmt, err)
		}
	}
}
