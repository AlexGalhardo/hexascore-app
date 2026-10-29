package backup

import (
	"bytes"
	"slices"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/crypt"
	"github.com/AlexGalhardo/vukem/storage"
)

func TestCountDiff(t *testing.T) {
	t.Parallel()

	want := map[string]int64{"users": 2, "items": 5}

	if d := CountDiff(want, map[string]int64{"users": 2, "items": 5}); d != nil {
		t.Fatalf("equal counts differ: %v", d)
	}

	got := CountDiff(want, map[string]int64{"users": 3, "extra": 1})
	exp := []string{"items: want 5, got 0", "users: want 2, got 3", "extra: not in the manifest, got 1"}

	if !slices.Equal(got, exp) {
		t.Fatalf("diff = %q, want %q", got, exp)
	}
}

func TestManifestRoundTripAndKey(t *testing.T) {
	t.Parallel()

	enc, err := crypt.New(bytes.Repeat([]byte{3}, 32))
	if err != nil {
		t.Fatal(err)
	}

	day1 := time.Date(2026, 11, 1, 9, 0, 0, 0, time.UTC)
	key := ObjectKey("daily", "test", day1)
	if mk := ManifestKey(
		key,
	); mk != "backups/manifests/postgres/daily/2026/11/01/hexascore-test-20261101T0900Z.dump.enc.json" {
		t.Fatalf("manifest key = %s", mk)
	}

	r := &Runner{Disk: storage.NewMemory(""), Enc: enc}
	if err := r.putManifest(t.Context(), Manifest{Key: key, Tables: map[string]int64{"users": 7}}); err != nil {
		t.Fatal(err)
	}

	m, err := r.ReadManifest(t.Context(), key)
	if err != nil || m.Tables["users"] != 7 {
		t.Fatalf("manifest = %+v, %v", m, err)
	}

	// The manifest is bound to its dump: another dump's key does not open it.
	if _, err := r.ReadManifest(t.Context(), ObjectKey("daily", "other", day1)); err == nil {
		t.Fatal("a missing manifest was read")
	}
}

// Child processes get the PG* variables and PATH only: pg_restore locates
// itself through PATH, and the app environment holds secrets.
func TestPgEnvIsMinimal(t *testing.T) {
	t.Setenv("PATH", "/usr/bin")
	t.Setenv("APP_KEY", "secret")

	env, err := pgEnv("postgres://u:p@db:5433/hx?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		"PGHOST=db", "PGPORT=5433", "PGUSER=u", "PGPASSWORD=p", "PGDATABASE=hx", "PATH=/usr/bin", "PGSSLMODE=disable",
	}
	if !slices.Equal(env, want) {
		t.Fatalf("env = %q, want %q", env, want)
	}
}
