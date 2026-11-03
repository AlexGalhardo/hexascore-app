package offers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/httpx"
)

// The verify job's budget (RFC-003 section 3): a chunk fits a queue job well inside its timeout, a day stops at the
// daily limit, and an offer must answer 404 or 410 on two runs in a row before it is deactivated.
const (
	VerifyChunk      = 250
	VerifyDailyLimit = 5000
	brokenRunsToGone = 2
	probeTimeout     = 10 * time.Second
	// maxGETBody caps the body read when a store refuses HEAD.
	maxGETBody = 64 << 10
)

// VerifyHosts are the only hosts the verify job probes. Amazon blocks bots (its links are checked by ISBN shape only)
// and TMDB-sourced offers are refreshed from the API, so neither appears. An admin URL on any other host stays
// unverified.
var VerifyHosts = []string{
	"store.steampowered.com", "www.gog.com", "store.epicgames.com", "play.google.com", "apps.apple.com",
	"archive.org", "www.mercadolivre.com.br", "produto.mercadolivre.com.br",
}

// HostLimits is the allowlist and rate table of the verify client: one request per second per host.
func HostLimits() map[string]httpx.HostLimit {
	m := make(map[string]httpx.HostLimit, len(VerifyHosts))
	for _, h := range VerifyHosts {
		m[h] = httpx.HostLimit{PerSecond: 1, InFlight: 1}
	}

	return m
}

// Verifiable reports whether the job may request rawURL: https, no credentials, no port, a host of VerifyHosts.
func Verifiable(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme != "https" || u.User != nil || u.Port() != "" {
		return false
	}

	return slices.Contains(VerifyHosts, u.Host)
}

// hostPattern matches the URLs Verifiable accepts, for the SQL that picks the due offers.
var hostPattern = func() string {
	quoted := make([]string, len(VerifyHosts))
	for i, h := range VerifyHosts {
		quoted[i] = regexp.QuoteMeta(h)
	}

	return "^https://(" + strings.Join(quoted, "|") + ")/"
}()

// Probe is what asking a store about one URL returned. Err is set when no HTTP answer came back.
type Probe struct {
	Status   int
	FinalURL string
	Err      error
}

// Prober asks a store about a URL.
type Prober interface {
	Probe(ctx context.Context, rawURL string) Probe
}

// HTTPProber probes with HEAD and falls back to a capped GET for stores that refuse HEAD. The client is the
// ingestion client (identified User-Agent, per-host throttle, SSRF-safe dialer, at most a few redirects followed).
type HTTPProber struct {
	Client *http.Client
}

// Probe implements Prober.
func (p HTTPProber) Probe(ctx context.Context, rawURL string) Probe {
	got := p.do(ctx, http.MethodHead, rawURL)
	if got.Err == nil && (got.Status == http.StatusForbidden || got.Status == http.StatusMethodNotAllowed ||
		got.Status == http.StatusNotImplemented) {
		got = p.do(ctx, http.MethodGet, rawURL)
	}

	return got
}

func (p HTTPProber) do(ctx context.Context, method, rawURL string) Probe {
	ctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, rawURL, nil)
	if err != nil {
		return Probe{Err: err}
	}

	resp, err := p.Client.Do(req)
	if err != nil {
		return Probe{Err: err}
	}

	defer func() { _ = resp.Body.Close() }()

	if method == http.MethodGet {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxGETBody))
	}

	return Probe{Status: resp.StatusCode, FinalURL: resp.Request.URL.String()}
}

// Check is the stored state of one offer under verification.
type Check struct {
	ID, URL, Store, Status string
	BrokenRuns             int
}

// Verdict is the state machine's answer. A zero Verdict (not Definitive) means the probe proved nothing and the
// offer keeps its state until the next run.
type Verdict struct {
	Definitive bool
	Status     string
	HTTP       int
	BrokenRuns int
	Active     bool
}

// Decide applies the state machine of RFC-003 section 3: 2xx is ok (redirected when the final page is another place),
// 404 and 410 are broken once and gone, with the offer deactivated, on the second run in a row; everything else
// (5xx, 429, a bot wall, a timeout) says nothing about the offer.
func Decide(c Check, p Probe) Verdict {
	switch {
	case p.Err != nil:
		return Verdict{}
	case p.Status >= 200 && p.Status < 300:
		status := "ok"
		if !samePlace(c.URL, p.FinalURL) {
			status = "redirected"
		}

		return Verdict{Definitive: true, Status: status, HTTP: p.Status, Active: true}
	case p.Status == http.StatusNotFound || p.Status == http.StatusGone:
		runs := c.BrokenRuns + 1
		if runs >= brokenRunsToGone {
			return Verdict{Definitive: true, Status: "gone", HTTP: p.Status, BrokenRuns: runs}
		}

		return Verdict{Definitive: true, Status: "broken", HTTP: p.Status, BrokenRuns: runs, Active: true}
	default:
		return Verdict{}
	}
}

// samePlace is true when the final URL is on the same host and at the same path or below it (a store appending the
// game's slug is not a different page).
func samePlace(orig, final string) bool {
	a, err := url.Parse(orig)
	b, err2 := url.Parse(final)

	if final == "" || err != nil || err2 != nil || a.Host != b.Host {
		return final == ""
	}

	ap, bp := strings.TrimRight(a.Path, "/"), strings.TrimRight(b.Path, "/")

	return bp == ap || strings.HasPrefix(bp, ap+"/")
}

// Stats counts what one Verify call did.
type Stats struct {
	Checked, OK, Redirected, Broken, Gone, Transient int
}

const dueSQL = `SELECT id::text, url, store, verify_status, broken_runs FROM offers
WHERE active AND source IN ('wikidata', 'admin') AND store NOT LIKE 'amazon%' AND url ~ $1
  AND (last_checked_at IS NULL OR last_checked_at < $2)
ORDER BY last_checked_at NULLS FIRST, id LIMIT $3`

const updateSQL = `UPDATE offers SET last_checked_at = $2,
  last_verified_at = CASE WHEN $3 THEN $2 ELSE last_verified_at END,
  verify_status = CASE WHEN $3 THEN $4 ELSE verify_status END,
  verify_http_status = CASE WHEN $3 THEN $5 ELSE verify_http_status END,
  broken_runs = CASE WHEN $3 THEN $6 ELSE broken_runs END,
  active = CASE WHEN $3 THEN $7 ELSE active END
WHERE id = $1`

// Verify probes up to limit due offers: the least recently checked first, only those not yet checked since the start
// of the run. Every attempt moves last_checked_at, so a timeout never makes the same offer due again inside the run;
// only a definitive answer moves last_verified_at and the status.
func Verify(ctx context.Context, db *database.DB, pr Prober, since, now time.Time, limit int) (Stats, error) {
	rows, err := db.QueryContext(ctx, dueSQL, hostPattern, since, limit)
	if err != nil {
		return Stats{}, fmt.Errorf("offers verify: list due: %w", err)
	}

	var due []Check

	for rows.Next() {
		var c Check
		if err := rows.Scan(&c.ID, &c.URL, &c.Store, &c.Status, &c.BrokenRuns); err != nil {
			_ = rows.Close()

			return Stats{}, err
		}

		due = append(due, c)
	}

	if err := rows.Err(); err != nil {
		_ = rows.Close()

		return Stats{}, err
	}

	if err := rows.Close(); err != nil {
		return Stats{}, err
	}

	var st Stats

	for _, c := range due {
		if err := ctx.Err(); err != nil {
			return st, err
		}

		v := Decide(c, pr.Probe(ctx, c.URL))

		if _, err := db.ExecContext(
			ctx,
			updateSQL,
			c.ID,
			now,
			v.Definitive,
			v.Status,
			v.HTTP,
			v.BrokenRuns,
			v.Active,
		); err != nil {
			return st, fmt.Errorf("offers verify: update %s: %w", c.ID, err)
		}

		st.Checked++

		switch v.Status {
		case "ok":
			st.OK++
		case "redirected":
			st.Redirected++
		case "broken":
			st.Broken++
		case "gone":
			st.Gone++
		default:
			st.Transient++
		}
	}

	return st, nil
}
