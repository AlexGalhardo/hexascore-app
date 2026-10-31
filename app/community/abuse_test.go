package community

import (
	"context"
	"errors"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

func TestHoneypotBlocksOnThirdHit(t *testing.T) {
	db := openTierDB(t)
	ctx := context.Background()
	userID := insertTierAuthor(t, db, "honeypot-block@example.test")
	now := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)

	for range 2 {
		err := RejectHoneypot(ctx, db, nil, userID, "https://spam.example", now)
		var problem *Problem
		if !errors.As(err, &problem) || problem.Status != 422 || problem.Code != "validation" {
			t.Fatalf("hit err %#v", err)
		}
	}

	var stats int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM contributor_stats WHERE user_id = $1`, userID).
		Scan(&stats); err != nil {
		t.Fatal(err)
	}
	if stats != 0 {
		t.Fatalf("stats rows %d before the third hit", stats)
	}

	if err := RejectHoneypot(ctx, db, nil, userID, "https://spam.example", now); err == nil {
		t.Fatal("third hit accepted")
	}
	got := readTier(t, db, userID)
	if got.tier != "blocked" || !got.blockedUntil.Equal(now.Add(blockFor)) {
		t.Fatalf("tier %+v", got)
	}
}

func TestHoneypotIgnoresHitsOutsideSevenDays(t *testing.T) {
	db := openTierDB(t)
	ctx := context.Background()
	userID := insertTierAuthor(t, db, "honeypot-window@example.test")
	now := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)

	if err := RejectHoneypot(ctx, db, nil, userID, "https://spam.example", now.Add(-8*24*time.Hour)); err == nil {
		t.Fatal("old hit accepted")
	}
	for range 2 {
		if err := RejectHoneypot(ctx, db, nil, userID, "https://spam.example", now); err == nil {
			t.Fatal("hit accepted")
		}
	}

	var stats int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM contributor_stats WHERE user_id = $1`, userID).
		Scan(&stats); err != nil {
		t.Fatal(err)
	}
	if stats != 0 {
		t.Fatalf("old hit counted toward the block, stats %d", stats)
	}
}

func TestCheckSourcesStoresStatus(t *testing.T) {
	db := openTierDB(t)
	ctx := context.Background()
	userID := insertTierAuthor(t, db, "sources@example.test")
	id := insertTierDraft(t, db, userID, 1, "draft", `{}`, time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC))

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	var hits atomic.Int32
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			hits.Add(1)
			_ = conn.Close()
		}
	}()

	raw := "https://" + ln.Addr().String() + "/secret"
	if _, err := db.ExecContext(
		ctx,
		`UPDATE suggestions SET source_urls = $2::text[] WHERE id = $1`,
		id,
		models.TextArray{raw, "https://bit.ly/x"},
	); err != nil {
		t.Fatal(err)
	}

	if err := CheckSources(ctx, db, id); err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 0 {
		t.Fatalf("source check connected %d times", hits.Load())
	}
	if err := CheckSources(ctx, db, "00000000-0000-0000-0000-000000000099"); err != nil {
		t.Fatal(err)
	}

	var blocked, refused string
	err = db.QueryRowContext(ctx, `SELECT meta->'source_status'->>$2, meta->'source_status'->>'https://bit.ly/x' FROM suggestions WHERE id = $1`, id, raw).
		Scan(&blocked, &refused)
	if err != nil {
		t.Fatal(err)
	}
	if blocked != "blocked" || refused != "refused" {
		t.Fatalf("source status blocked=%q refused=%q", blocked, refused)
	}
}
