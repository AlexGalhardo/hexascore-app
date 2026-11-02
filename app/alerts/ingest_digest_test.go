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
		Created: 12, Updated: 340, Failed: 3, Covers: 25, RateLimited: 2, Queued: 4, Running: 1,
		Errors:      []IngestError{{"wikidata", "http_5xx", 3}, {"tmdb", "timeout", 1}, {"a<b", "x&y", 1}},
		LLMMonthUSD: 1.23, LLMProjectedUSD: 9.8,
	}

	want, err := os.ReadFile("testdata/ingest_digest.golden")
	if err != nil {
		t.Fatal(err)
	}

	if got := FormatIngest(s); got != string(want) {
		t.Errorf("digest differs from the golden fixture:\n%s\nwant:\n%s", got, want)
	}

	s.LLMProjectedUSD = 18
	if got := FormatIngest(s); !strings.Contains(got, "<b>over the alert</b>") {
		t.Errorf("projected spend at the alert is not flagged: %s", got)
	}
}

func TestIngestSummaryQuietAndSchedule(t *testing.T) {
	if !(IngestSummary{}).Quiet() {
		t.Error("an empty window is quiet")
	}

	for _, s := range []IngestSummary{{Created: 1}, {Failed: 1}, {Queued: 1}, {Running: 1}, {Errors: []IngestError{{}}}} {
		if s.Quiet() {
			t.Errorf("%+v is not quiet", s)
		}
	}

	sp := Location()
	for at, want := range map[time.Time]bool{
		time.Date(2026, 11, 2, 8, 0, 0, 0, sp):   true,
		time.Date(2026, 11, 2, 8, 30, 0, 0, sp):  false,
		time.Date(2026, 11, 2, 20, 0, 0, 0, sp):  true,
		time.Date(2026, 11, 2, 14, 0, 0, 0, sp):  false,
		time.Date(2026, 11, 2, 20, 29, 0, 0, sp): true,
	} {
		if scheduledSend(at) != want {
			t.Errorf("scheduledSend(%v) = %v, want %v", at, !want, want)
		}
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
