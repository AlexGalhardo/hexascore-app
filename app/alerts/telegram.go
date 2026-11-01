package alerts

import (
	"cmp"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/netip"
	"regexp"
	"strings"
	"time"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/config"
	"github.com/AlexGalhardo/vukem/httpclient"
	"github.com/AlexGalhardo/vukem/logging/telegram"
)

// DedupWindow is how long one fingerprint (alert code, check or kind, route, error type, message) stays quiet on
// Telegram after it was sent. Railway logs keep every line; the daily digest keeps the counts.
const DedupWindow = 30 * time.Minute

// defaultRateLimit caps the log records sent per minute per chat when TELEGRAM_RATE_LIMIT is unset (SDD 12 §7.5).
const defaultRateLimit = 20

// pageCodes ring the page chat (Sev "page" in SDD 12 §7.5). A1 and A2 come from Railway and the smoke test; they are
// listed so a line carrying them would page too.
var pageCodes = map[string]bool{"A1": true, "A2": true, "A3": true, "A5": true, "A8": true, "A10": true, "A12": true}

// alertCode matches the codes of the SDD 12 §7.5 table; informational keys (scrape, trending_counter_dropped) do not.
var alertCode = regexp.MustCompile(`^A\d+$`)

// saoPaulo is the on-call zone; Brazil has had no DST since 2019, so the fixed fallback is exact without tzdata.
var saoPaulo = func() *time.Location {
	if loc, err := time.LoadLocation("America/Sao_Paulo"); err == nil {
		return loc
	}

	return time.FixedZone("BRT", -3*60*60)
}()

// Location is the on-call zone the digest and the quiet hours use.
func Location() *time.Location { return saoPaulo }

// Counter is the part of *cache.Cache the dedup needs: Increment starts a missing key at 0 with ttl.
type Counter interface {
	Increment(ctx context.Context, key string, delta int64, ttl time.Duration) (int64, error)
}

// TelegramOptions configure NewTelegram.
type TelegramOptions struct {
	// Config is vukem's TELEGRAM_* settings (bot token, page chat, minimum level, rate limit, API URL).
	Config vukem.TelegramConfig
	// ErrorChatID is TELEGRAM_ERROR_CHAT_ID, the silent chat; empty sends silent messages to Config.ChatID.
	ErrorChatID string
	AppName     string
	Env         string
	// Production blocks loopback and private addresses for the Bot API URL (a fake server is local).
	Production bool
	// Dedup is the cache; nil sends every record.
	Dedup Counter
	// OnError is told about delivery failures; the errors never carry the bot token.
	OnError func(error)
	Now     func() time.Time
}

// Notifier is an slog.Handler that sends alert lines and errors to Telegram (SDD 12 §7.5): page-severity alerts and
// LevelFatal records ring the page chat; other alert lines (WARN and up) and records at TELEGRAM_MIN_LEVEL go to the
// error chat without sound. A nil *Notifier sends nothing.
type Notifier struct {
	// pageTG and silentTG send and close; page and silent carry the WithAttrs and WithGroup state.
	pageTG, silentTG *telegram.Handler
	page, silent     slog.Handler
	minLevel         slog.Level
	dedup            Counter
	now              func() time.Time
	attrs            []slog.Attr
}

// NewTelegram returns the notifier, or nil when TELEGRAM_BOT_TOKEN or TELEGRAM_CHAT_ID is unset (no sends, no errors).
func NewTelegram(o TelegramOptions) (*Notifier, error) {
	tc := o.Config
	if tc.BotToken.IsZero() || tc.ChatID == "" {
		return nil, nil //nolint:nilnil // off when unset is the documented state, not an error
	}

	var minLevel slog.Level
	if err := minLevel.UnmarshalText([]byte(strings.ToUpper(cmp.Or(tc.MinLevel, "error")))); err != nil {
		return nil, fmt.Errorf("alerts: TELEGRAM_MIN_LEVEL: %w", err)
	}

	opts := httpclient.Options{
		Timeout: 5 * time.Second,
		// The breaker stops a dead Bot API from holding the sender queue; telegram retries a 429 itself.
		WrapTransport: httpclient.Breaker(httpclient.BreakerOptions{Failures: 5, Cooldown: time.Minute}),
	}
	if !o.Production {
		opts.Allow = []netip.Prefix{netip.MustParsePrefix("127.0.0.0/8"), netip.MustParsePrefix("::1/128")}
	}

	base := telegram.Config{
		BotToken: tc.BotToken, AppName: o.AppName, Env: o.Env, ThreadID: tc.ThreadID,
		// The Notifier filters; the handlers only format, rate-limit and deliver.
		MinLevel: slog.LevelWarn, MinLevelSet: true,
		RateLimit: cmp.Or(tc.RateLimit, defaultRateLimit), RatePeriod: time.Minute,
		BaseURL: strings.TrimSuffix(cmp.Or(tc.APIURL, "https://api.telegram.org"), "/"),
		HTTP:    httpclient.Safe(opts), OnError: o.OnError,
	}

	page, silent := base, base
	page.ChatID = tc.ChatID
	silent.ChatID = cmp.Or(o.ErrorChatID, tc.ChatID)
	silent.Silent = true

	now := o.Now
	if now == nil {
		now = time.Now
	}

	pageTG, silentTG := telegram.NewHandler(page), telegram.NewHandler(silent)

	return &Notifier{
		pageTG: pageTG, silentTG: silentTG, page: pageTG, silent: silentTG,
		minLevel: minLevel, dedup: o.Dedup, now: now,
	}, nil
}

// FromApp builds the notifier from the app's TELEGRAM_* variables and cache; nil when they are unset.
func FromApp(app *vukem.App) (*Notifier, error) {
	tc, err := config.Load[vukem.TelegramConfig](app.Source)
	if err != nil {
		return nil, err
	}

	if tc.BotToken.IsZero() || tc.ChatID == "" {
		return nil, nil //nolint:nilnil // off when unset
	}

	c, err := app.Cache()
	if err != nil {
		return nil, err
	}

	return NewTelegram(TelegramOptions{
		Config: tc, ErrorChatID: app.Source.Get("TELEGRAM_ERROR_CHAT_ID"),
		AppName: app.Config.Name, Env: string(app.Config.Env), Production: app.Config.IsProduction(),
		Dedup: c, Now: app.Now,
		// Logged at WARN without an alert key, so it never comes back to Telegram.
		OnError: func(err error) { app.Log.Warn("telegram delivery failed", "err", err) },
	})
}

// SendPage delivers htmlText to the page chat now (alerts:test).
func (n *Notifier) SendPage(ctx context.Context, htmlText string) error {
	if n == nil {
		return telegram.ErrDisabled
	}

	return n.pageTG.SendSync(ctx, htmlText)
}

// SendSilent delivers htmlText to the error chat now, without sound (the daily digest).
func (n *Notifier) SendSilent(ctx context.Context, htmlText string) error {
	if n == nil {
		return telegram.ErrDisabled
	}

	return n.silentTG.SendSync(ctx, htmlText)
}

// Close drains both send queues.
func (n *Notifier) Close(ctx context.Context) error {
	if n == nil {
		return nil
	}

	return errors.Join(n.pageTG.Close(ctx), n.silentTG.Close(ctx))
}

// Enabled implements slog.Handler.
func (n *Notifier) Enabled(_ context.Context, l slog.Level) bool {
	return l >= min(n.minLevel, slog.LevelWarn)
}

// Handle implements slog.Handler.
func (n *Notifier) Handle(ctx context.Context, r slog.Record) error {
	f := Fields(n.attrs, r)
	code := f["alert"]

	isAlert := alertCode.MatchString(code)
	if (isAlert && r.Level < slog.LevelWarn) || (!isAlert && r.Level < n.minLevel) {
		return nil
	}

	if !n.first(ctx, Fingerprint(r.Message, f)) {
		return nil
	}

	h := n.silent
	if r.Level >= telegram.LevelFatal || (pageCodes[code] && (code != "A8" || !quiet(n.now()))) {
		h = n.page
	}

	return h.Handle(ctx, r)
}

// WithAttrs implements slog.Handler.
func (n *Notifier) WithAttrs(attrs []slog.Attr) slog.Handler {
	c := *n
	c.attrs = append(append([]slog.Attr(nil), n.attrs...), attrs...)
	c.page, c.silent = n.page.WithAttrs(attrs), n.silent.WithAttrs(attrs)

	return &c
}

// WithGroup implements slog.Handler.
func (n *Notifier) WithGroup(name string) slog.Handler {
	c := *n
	c.page, c.silent = n.page.WithGroup(name), n.silent.WithGroup(name)

	return &c
}

// first reports whether fp is new in the dedup window. A cache error sends: a broken cache must not hide a page.
func (n *Notifier) first(ctx context.Context, fp string) bool {
	if n.dedup == nil {
		return true
	}

	c, err := n.dedup.Increment(ctx, "alerts:telegram:"+fp, 1, DedupWindow)

	return err != nil || c == 1
}

// quiet reports the 23:00-08:00 America/Sao_Paulo quiet hours, when A8 does not ring (SDD 12 §7.5).
func quiet(t time.Time) bool {
	h := t.In(saoPaulo).Hour()
	return h >= 23 || h < 8
}

// fingerprintKeys identify an error; request_id is kept as a sample, never part of the fingerprint.
var fingerprintKeys = []string{"alert", "check", "kind", "route", "error.type"}

// Fields returns the top-level attributes that identify a record (alert, check, kind, route, request_id) and the
// Go type of its err or error attribute as error.type.
func Fields(handlerAttrs []slog.Attr, r slog.Record) map[string]string {
	f := map[string]string{}
	add := func(a slog.Attr) bool {
		switch a.Key {
		case "alert", "check", "kind", "route", "request_id":
			f[a.Key] = a.Value.String()
		case "err", "error":
			if e, ok := a.Value.Any().(error); ok {
				f["error.type"] = fmt.Sprintf("%T", e)
			}
		}

		return true
	}

	for _, a := range handlerAttrs {
		add(a)
	}

	r.Attrs(add)

	return f
}

// Fingerprint is the stable 16-hex id of an error (SDD 12 §7.2): alert code, check, kind, route, error type, message.
func Fingerprint(msg string, f map[string]string) string {
	h := sha256.New()
	for _, k := range fingerprintKeys {
		h.Write([]byte(f[k]))
		h.Write([]byte{0})
	}

	h.Write([]byte(msg))

	return hex.EncodeToString(h.Sum(nil))[:16]
}
