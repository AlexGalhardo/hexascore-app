package alerts

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/config"
	"github.com/AlexGalhardo/vukem/logging/telegram"

	"github.com/AlexGalhardo/hexascore-app/app/logging"
)

const (
	testToken     = "123456789:AAH-secret_token_value_for_tests_0123"
	pageChat      = "-100page"
	errorChat     = "-100error"
	quietHourUTC  = 5  // 02:00 America/Sao_Paulo
	awakeHourUTC  = 13 // 10:00 America/Sao_Paulo
	telegramReply = `{"ok":true,"result":{}}`
)

type sent struct {
	Path   string
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
	Silent bool   `json:"disable_notification"`
}

// fakeTelegram is a Bot API server that records every sendMessage call.
type fakeTelegram struct {
	srv    *httptest.Server
	mu     sync.Mutex
	got    []sent
	status int
}

func newFakeTelegram(t *testing.T) *fakeTelegram {
	t.Helper()
	f := &fakeTelegram{status: http.StatusOK}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var m sent
		_ = json.NewDecoder(r.Body).Decode(&m)
		m.Path = r.URL.Path
		f.mu.Lock()
		f.got = append(f.got, m)
		status := f.status
		f.mu.Unlock()
		w.WriteHeader(status)
		if status == http.StatusOK {
			_, _ = w.Write([]byte(telegramReply))
			return
		}
		_, _ = w.Write([]byte(`{"ok":false,"description":"Internal Server Error"}`))
	}))
	t.Cleanup(f.srv.Close)

	return f
}

func (f *fakeTelegram) messages() []sent {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]sent(nil), f.got...)
}

// fakeCounter is a cache.Increment with a movable clock: the ttl starts on the first increment.
type fakeCounter struct {
	now  time.Time
	vals map[string]int64
	exp  map[string]time.Time
	err  error
}

func (c *fakeCounter) Increment(_ context.Context, key string, delta int64, ttl time.Duration) (int64, error) {
	if c.err != nil {
		return 0, c.err
	}
	if c.vals == nil {
		c.vals, c.exp = map[string]int64{}, map[string]time.Time{}
	}
	if e, ok := c.exp[key]; !ok || !c.now.Before(e) {
		c.vals[key], c.exp[key] = 0, c.now.Add(ttl)
	}
	c.vals[key] += delta

	return c.vals[key], nil
}

func at(hourUTC int) time.Time { return time.Date(2026, 11, 1, hourUTC, 0, 0, 0, time.UTC) }

func newNotifier(t *testing.T, f *fakeTelegram, errChat string, dedup Counter, now time.Time) *Notifier {
	t.Helper()
	n, err := NewTelegram(TelegramOptions{
		Config: vukem.TelegramConfig{
			BotToken: config.NewSecret(testToken), ChatID: pageChat, MinLevel: "error", APIURL: f.srv.URL,
		},
		ErrorChatID: errChat,
		AppName:     "HexaScore",
		Env:         "test",
		Dedup:       dedup,
		Now:         func() time.Time { return now },
	})
	if err != nil {
		t.Fatal(err)
	}
	if n == nil {
		t.Fatal("configured notifier is nil")
	}

	return n
}

func drain(t *testing.T, n *Notifier) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := n.Close(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestTelegram_OffWhenUnset(t *testing.T) {
	for _, tc := range []vukem.TelegramConfig{
		{},
		{BotToken: config.NewSecret(testToken)},
		{ChatID: pageChat},
	} {
		n, err := NewTelegram(TelegramOptions{Config: tc})
		if err != nil || n != nil {
			t.Fatalf("%+v: want nil notifier and no error, got %v, %v", tc.ChatID, n, err)
		}
	}

	// A nil notifier is safe to use: no sends, no errors.
	var n *Notifier
	if err := n.SendPage(context.Background(), "x"); !errors.Is(err, telegram.ErrDisabled) {
		t.Fatalf("SendPage on nil: %v", err)
	}
	if err := n.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestTelegram_Routing(t *testing.T) {
	f := newFakeTelegram(t)
	n := newNotifier(t, f, errorChat, nil, at(awakeHourUTC))
	log := slog.New(n)
	ctx := context.Background()

	Error(ctx, log, A5, "inertia ssr fallback", "public", true)                    // page
	Error(ctx, log, A6, "job failed twice in a row", "kind", "k")                  // warn -> silent
	log.Warn("no incremental Wikidata run in 26 h", "alert", "A7")                 // warn alert -> silent
	log.Warn("trending counter dropped keys", "alert", "trending_counter_dropped") // informational: not sent
	log.Warn("plain warning")                                                      // below TELEGRAM_MIN_LEVEL: not sent
	log.Error("plain error")                                                       // silent
	log.Log(ctx, telegram.LevelFatal, "fatal test record")                         // page
	Error(ctx, log, "A8", "backup missing")                                        // page (awake)
	drain(t, n)

	want := map[string]struct {
		chat   string
		silent bool
	}{
		"inertia ssr fallback":                {pageChat, false},
		"job failed twice in a row":           {errorChat, true},
		"no incremental Wikidata run in 26 h": {errorChat, true},
		"plain error":                         {errorChat, true},
		"fatal test record":                   {pageChat, false},
		"backup missing":                      {pageChat, false},
	}
	got := f.messages()
	if len(got) != len(want) {
		t.Fatalf("want %d messages, got %d: %+v", len(want), len(got), got)
	}
	for _, m := range got {
		if m.Path != "/bot"+testToken+"/sendMessage" {
			t.Errorf("path %q", m.Path)
		}
		matched := false
		for msg, w := range want {
			if strings.Contains(m.Text, msg) {
				matched = true
				if m.ChatID != w.chat || m.Silent != w.silent {
					t.Errorf("%q: chat %s silent %v, want %s %v", msg, m.ChatID, m.Silent, w.chat, w.silent)
				}
			}
		}
		if !matched {
			t.Errorf("unexpected message %q", m.Text)
		}
	}
}

func TestTelegram_A8SilentInQuietHours(t *testing.T) {
	f := newFakeTelegram(t)
	n := newNotifier(t, f, errorChat, nil, at(quietHourUTC))
	Error(context.Background(), slog.New(n), "A8", "backup missing")
	drain(t, n)

	got := f.messages()
	if len(got) != 1 || got[0].ChatID != errorChat || !got[0].Silent {
		t.Fatalf("A8 during quiet hours must go to the silent chat: %+v", got)
	}
}

func TestTelegram_OneChatFallback(t *testing.T) {
	f := newFakeTelegram(t)
	n := newNotifier(t, f, "", nil, at(awakeHourUTC))
	log := slog.New(n)
	Error(context.Background(), log, A5, "inertia ssr fallback")
	log.Error("plain error")
	drain(t, n)

	got := f.messages()
	if len(got) != 2 {
		t.Fatalf("want 2 messages, got %+v", got)
	}
	for _, m := range got {
		if m.ChatID != pageChat {
			t.Errorf("without TELEGRAM_ERROR_CHAT_ID every message goes to TELEGRAM_CHAT_ID: %+v", m)
		}
		if strings.Contains(m.Text, "plain error") != m.Silent {
			t.Errorf("only the plain error is silent: %+v", m)
		}
	}
}

func TestTelegram_Dedup(t *testing.T) {
	f := newFakeTelegram(t)
	c := &fakeCounter{now: at(awakeHourUTC)}
	n := newNotifier(t, f, errorChat, c, at(awakeHourUTC))
	log := slog.New(n)
	ctx := context.Background()

	log.Warn("TMDB records older than 170 days", "alert", "A7", "check", "tmdb_age", "count", 3)
	log.Warn("TMDB records older than 170 days", "alert", "A7", "check", "tmdb_age", "count", 4) // same code and key
	log.Warn("ingestion run failed in the last 24 h", "alert", "A7", "check", "failed_run")      // other key
	log.With("alert", "A7").Warn("TMDB records older than 170 days", "check", "tmdb_age")        // same, via With

	c.now = c.now.Add(DedupWindow - time.Second)
	log.Warn("TMDB records older than 170 days", "alert", "A7", "check", "tmdb_age")
	c.now = c.now.Add(time.Second)
	log.Warn("TMDB records older than 170 days", "alert", "A7", "check", "tmdb_age") // window over: resent

	c.err = errors.New("cache down")
	Error(ctx, log, A5, "inertia ssr fallback") // a broken cache never hides a page
	drain(t, n)

	got := f.messages()
	if len(got) != 4 {
		t.Fatalf("want 4 messages (tmdb_age, failed_run, tmdb_age after window, A5), got %d: %+v", len(got), got)
	}
}

func TestFingerprint(t *testing.T) {
	base := map[string]string{"alert": "A7", "check": "tmdb_age", "route": "GET /x"}
	fp := Fingerprint("m", base)
	if len(fp) != 16 ||
		fp != Fingerprint("m", map[string]string{"route": "GET /x", "check": "tmdb_age", "alert": "A7"}) {
		t.Fatalf("fingerprint must be stable 16 hex chars: %q", fp)
	}
	for _, other := range []map[string]string{
		{"alert": "A6", "check": "tmdb_age", "route": "GET /x"},
		{"alert": "A7", "check": "failed_run", "route": "GET /x"},
		{"alert": "A7", "check": "tmdb_age", "route": "GET /y"},
		{"alert": "A7", "check": "tmdb_age", "route": "GET /x", "error.type": "*net.OpError"},
	} {
		if Fingerprint("m", other) == fp {
			t.Errorf("%v shares the fingerprint", other)
		}
	}
	if Fingerprint("other", base) == fp {
		t.Error("message is part of the fingerprint")
	}
	// The count or the request id must not split a fingerprint.
	if Fingerprint(
		"m",
		map[string]string{"alert": "A7", "check": "tmdb_age", "route": "GET /x", "request_id": "r1"},
	) != fp {
		t.Error("request_id must not change the fingerprint")
	}
}

func TestSecurity_A09_TelegramTokenNeverLogged(t *testing.T) {
	f := newFakeTelegram(t)
	f.status = http.StatusInternalServerError

	var reported []error
	var mu sync.Mutex
	var logs bytes.Buffer
	n, err := NewTelegram(TelegramOptions{
		Config: vukem.TelegramConfig{
			BotToken: config.NewSecret(testToken), ChatID: pageChat, MinLevel: "error", APIURL: f.srv.URL,
		},
		OnError: func(err error) {
			mu.Lock()
			reported = append(reported, err)
			mu.Unlock()
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	log := slog.New(slog.NewMultiHandler(slog.NewJSONHandler(&logs, nil), n))
	log.Error("delivery fails")

	sendErr := n.SendPage(context.Background(), "x")
	drain(t, n)

	// A dead server: the transport error (a url.Error) carries the URL, so the token, unless redacted.
	f.srv.Close()
	dead, err := NewTelegram(TelegramOptions{Config: vukem.TelegramConfig{
		BotToken: config.NewSecret(testToken), ChatID: pageChat, APIURL: f.srv.URL,
	}})
	if err != nil {
		t.Fatal(err)
	}
	deadErr := dead.SendPage(context.Background(), "x")
	drain(t, dead)

	if sendErr == nil || deadErr == nil {
		t.Fatalf("both sends must fail: %v, %v", sendErr, deadErr)
	}
	mu.Lock()
	all := append([]error{sendErr, deadErr}, reported...)
	mu.Unlock()
	for _, e := range all {
		if strings.Contains(e.Error(), testToken) || strings.Contains(e.Error(), "AAH-secret") {
			t.Errorf("error leaks the bot token: %q", e)
		}
	}
	if strings.Contains(logs.String(), testToken) {
		t.Errorf("log leaks the bot token: %s", logs.String())
	}

	// The app log handler scrubs a token that reaches a message or an error by any other path.
	for _, s := range []string{
		"Post \"https://api.telegram.org/bot" + testToken + "/sendMessage\": dial tcp: timeout",
		"token " + testToken,
	} {
		if got := logging.Scrub(s); strings.Contains(got, testToken) || strings.Contains(got, "AAH-secret") {
			t.Errorf("Scrub(%q) = %q", s, got)
		}
	}
}
