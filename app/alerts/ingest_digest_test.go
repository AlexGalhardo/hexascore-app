package alerts

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

var digestAt = time.Date(2026, 11, 2, 17, 30, 0, 0, time.UTC) // 14:30 America/Sao_Paulo

func TestFormatIngestGolden(t *testing.T) {
	s := IngestSummary{
		At: digestAt, Since: digestAt.Add(-IngestWindow),
		Created: 12, Updated: 340, Failed: 3, Covers: 25, RateLimited: 2, Queued: 4, Running: 2,
		Books: 100, Movies: 200, Series: 50, Games: 30, HeartbeatOK: true, HeartbeatAge: 42 * time.Second,
		NewByType: map[string]int{"movie": 9, "series": 3}, CreatedQuiet: 12,
		Active: []ActiveRun{
			{
				Job:     "fetch",
				Source:  "wikidata",
				Age:     12 * time.Minute,
				BeatAge: time.Minute,
				Seen:    4210,
				Changed: 180,
				Failed:  2,
			},
			{Job: "normalize", Age: 40 * time.Minute, BeatAge: 9 * time.Minute, Seen: 1020, Changed: 64},
		},
		Fetched:     []SourceCount{{"wikidata", 4200}, {"tmdb", 1020}},
		Errors:      []IngestError{{"wikidata", "http_5xx", 3}, {"tmdb", "timeout", 1}, {"a<b", "x&y", 1}},
		ErrorSample: "wikidata Q4521: save alt title: duplicate key <x>",
		LLMMonthUSD: 1.23, LLMProjectedUSD: 9.8,
	}

	want, err := os.ReadFile("testdata/ingest_digest.golden")
	if err != nil {
		t.Fatal(err)
	}

	if got := FormatIngest(s); got != string(want) {
		t.Errorf("digest differs from the golden fixture:\n%s\nwant:\n%s", got, want)
	}

	quiet := IngestSummary{At: digestAt, Active: s.Active}
	if got := FormatIngest(quiet); !strings.Contains(got, "⚠️ <b>0 new items in 45 min</b> (2 runs active)") {
		t.Errorf("a quiet catalogue is not flagged: %s", got)
	}

	s.LLMProjectedUSD = 18
	if got := FormatIngest(s); !strings.Contains(got, "<b>over the alert</b>") {
		t.Errorf("projected spend at the alert is not flagged: %s", got)
	}
}

// SendIngestDigest with a nil notifier sends nothing and needs no database.
func TestSendIngestDigestWithoutTelegram(t *testing.T) {
	if sent, err := SendIngestDigest(context.Background(), nil, nil, digestAt, true); sent || err != nil {
		t.Errorf("sent %v, err %v", sent, err)
	}
}

// TELEGRAM_OPS_CHAT_ID wins; unset it falls back to TELEGRAM_ERROR_CHAT_ID, then to the page chat. Always silent.
func TestTelegram_OpsChatRouting(t *testing.T) {
	for _, c := range []struct{ ops, errChat, want string }{
		{"ops-chat", errorChat, "ops-chat"},
		{"", errorChat, errorChat},
		{"", "", pageChat},
	} {
		f := newFakeTelegram(t)

		n, err := NewTelegram(TelegramOptions{
			Config:      testConfig(f),
			ErrorChatID: c.errChat, OpsChatID: c.ops, AppName: "HexaScore", Env: "test",
			Now: func() time.Time { return at(awakeHourUTC) },
		})
		if err != nil {
			t.Fatal(err)
		}

		if err := n.SendOps(context.Background(), "digest"); err != nil {
			t.Fatal(err)
		}

		drain(t, n)

		got := f.messages()
		if len(got) != 1 || got[0].ChatID != c.want || !got[0].Silent {
			t.Errorf("ops %q error %q: %+v, want one silent message to %s", c.ops, c.errChat, got, c.want)
		}
	}
}
