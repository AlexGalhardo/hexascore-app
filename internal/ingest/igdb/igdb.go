// Package igdb is the IGDB connector (SDD 06 §7.3): official cover and fallback facts for games with an IGDB slug
// (Wikidata P5794). It authenticates with a Twitch app access token (client credentials grant). Port of legacy
// connectors/igdb.py.
//
// Terms: Twitch Developer Services Agreement (https://www.twitch.tv/p/en/legal/developer-agreement/) and the IGDB
// API docs (https://api-docs.igdb.com/, read 2026-11-01: commercial use through a partnership, visible user-facing
// attribution to IGDB.com, 4 requests/s). Owner cleared the legal read in Q-A16; the D-36 interim retention (refresh
// after 24 h, nothing kept past 7 days) still applies and IGDB data stays site-only.
package igdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"iter"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

// Endpoints.
const (
	TokenURL = "https://id.twitch.tv/oauth2/token" //nolint:gosec // an endpoint URL, not a credential
	APIURL   = "https://api.igdb.com/v4"
	// CoverURL is the 528x748 portrait preset, the largest one.
	CoverURL = "https://images.igdb.com/igdb/image/upload/t_cover_big_2x/%s.jpg"
)

// Fields is the games query projection.
const Fields = "slug,summary,first_release_date,cover.image_id,genres.name,platforms.name," +
	"involved_companies.developer,involved_companies.company.name,videos.video_id"

// Batch is the slugs per games query (one request returns at most 500 games).
const Batch = 200

// MaxTrailers caps the trailer ids kept per game.
const MaxTrailers = 3

var (
	_ ingest.Connector = (*Connector)(nil)

	// Slugs go into the query body, so anything outside IGDB's slug alphabet is dropped instead of escaped.
	safeSlug = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
)

// Payload is the stored record body. Overview is only the input of the synopsis step (FR-06-3).
type Payload struct {
	Qid         string   `json:"qid"`
	ID          int      `json:"id"`
	Slug        string   `json:"slug"`
	URL         string   `json:"url"`
	Cover       string   `json:"cover,omitempty"`
	Overview    string   `json:"overview,omitempty"`
	ReleaseDate string   `json:"release_date,omitempty"`
	Genres      []string `json:"genres"`
	Platforms   []string `json:"platforms"`
	Creators    []string `json:"creators"`
	Trailers    []string `json:"trailers"`
}

type named struct {
	Name string `json:"name"`
}

type game struct {
	ID               int    `json:"id"`
	Slug             string `json:"slug"`
	Summary          string `json:"summary"`
	FirstReleaseDate *int64 `json:"first_release_date"`
	Cover            *struct {
		ImageID string `json:"image_id"`
	} `json:"cover"`
	Genres            []named `json:"genres"`
	Platforms         []named `json:"platforms"`
	InvolvedCompanies []struct {
		Developer bool   `json:"developer"`
		Company   *named `json:"company"`
	} `json:"involved_companies"`
	Videos []struct {
		VideoID string `json:"video_id"`
	} `json:"videos"`
}

// GamesQuery is the Apicalypse body for a batch of safe slugs.
func GamesQuery(slugs []string) string {
	quoted := make([]string, len(slugs))
	for i, s := range slugs {
		quoted[i] = `"` + s + `"`
	}

	return fmt.Sprintf("fields %s; where slug = (%s); limit %d;", Fields, strings.Join(quoted, ","), len(slugs))
}

func names(in []named) []string {
	out := make([]string, 0, len(in))
	for _, n := range in {
		out = append(out, n.Name)
	}

	return out
}

// ParseGames turns a games response into payloads; qids maps slug to QID for the batch and games outside it are
// ignored.
func ParseGames(raw []byte, qids map[string]string) ([]Payload, error) {
	var games []game
	if err := json.Unmarshal(raw, &games); err != nil {
		return nil, fmt.Errorf("decode igdb games: %w", err)
	}

	var out []Payload

	for _, g := range games {
		qid, ok := qids[g.Slug]
		if !ok {
			continue
		}

		p := Payload{
			Qid: qid, ID: g.ID, Slug: g.Slug, URL: "https://www.igdb.com/games/" + g.Slug,
			Overview: g.Summary, Genres: names(g.Genres), Platforms: names(g.Platforms),
			Creators: []string{}, Trailers: []string{},
		}

		if g.Cover != nil && g.Cover.ImageID != "" {
			p.Cover = fmt.Sprintf(CoverURL, g.Cover.ImageID)
		}

		// Unix seconds may be negative for pre-1970 releases.
		if g.FirstReleaseDate != nil {
			p.ReleaseDate = time.Unix(*g.FirstReleaseDate, 0).UTC().Format(time.DateOnly)
		}

		for _, c := range g.InvolvedCompanies {
			if c.Developer && c.Company != nil && !slices.Contains(p.Creators, c.Company.Name) {
				p.Creators = append(p.Creators, c.Company.Name)
			}
		}

		for _, v := range g.Videos[:min(MaxTrailers, len(g.Videos))] {
			p.Trailers = append(p.Trailers, v.VideoID)
		}

		out = append(out, p)
	}

	return out, nil
}

// Connector fetches IGDB games for stored items. A Target carries the QID in ID and the IGDB slug in Type.
type Connector struct {
	clientID, secret string
	// TokenURL and APIURL override the endpoints; tests only.
	TokenURL, APIURL string
}

// New returns a Connector that authenticates with the Twitch app credentials.
func New(clientID, secret string) *Connector {
	return &Connector{clientID: clientID, secret: secret, TokenURL: TokenURL, APIURL: APIURL}
}

// Source names the connector in source_records.
func (*Connector) Source() string { return "igdb" }

// targetsSQL: $1 after, $2 since (null = any), $3 only missing, $4 limit (0 = none).
const targetsSQL = `
SELECT i.wikidata_qid, i.external_ids->>'igdb'
FROM items i
WHERE i.type = 'game' AND i.external_ids ? 'igdb'
  AND i.wikidata_qid > $1
  AND ($2::timestamptz IS NULL OR i.updated_at >= $2)
  AND (NOT $3::boolean OR NOT EXISTS (
    SELECT 1 FROM source_records s WHERE s.source = 'igdb' AND s.external_id = i.wikidata_qid))
ORDER BY i.wikidata_qid
LIMIT CASE WHEN $4::int > 0 THEN $4::int END`

// Targets yields the games that carry an IGDB slug, after the cursor.
func (*Connector) Targets(
	ctx context.Context, db database.Executor, p ingest.RunParams, after string,
) iter.Seq2[ingest.Target, error] {
	return func(yield func(ingest.Target, error) bool) {
		var since any
		if p.Mode == run.ModeIncremental && !p.Since.IsZero() {
			since = p.Since
		}

		rows, err := db.QueryContext(ctx, targetsSQL, after, since, p.Mode == run.ModeBootstrap, p.Limit)
		if err != nil {
			yield(ingest.Target{}, fmt.Errorf("igdb targets: %w", err))

			return
		}

		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var qid, slug string
			if err := rows.Scan(&qid, &slug); err != nil {
				yield(ingest.Target{}, fmt.Errorf("scan igdb target: %w", err))

				return
			}

			if !yield(ingest.Target{ID: qid, Type: slug}, nil) {
				return
			}
		}

		if err := rows.Err(); err != nil {
			yield(ingest.Target{}, fmt.Errorf("igdb targets: %w", err))
		}
	}
}

// token runs the client credentials grant. A 4xx answer means the credentials are wrong, which is permanent.
func (c *Connector) token(ctx context.Context, client *ingest.Client) (string, error) {
	if c.clientID == "" || c.secret == "" {
		return "", fmt.Errorf(
			"igdb: IGDB_TWITCH_CLIENT_ID and IGDB_TWITCH_CLIENT_SECRET must be set: %w",
			queue.ErrPermanent,
		)
	}

	q := url.Values{"client_id": {c.clientID}, "client_secret": {c.secret}, "grant_type": {"client_credentials"}}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.TokenURL+"?"+q.Encode(), nil)
	if err != nil {
		return "", fmt.Errorf("build twitch token request: %w", err)
	}

	resp, err := client.JSON.Do(req)
	if err != nil {
		// The URL carries the secret; never wrap it into a stored error message.
		return "", errors.New("twitch token request failed")
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		return "", fmt.Errorf(
			"twitch rejected the igdb credentials (status %d): %w",
			resp.StatusCode,
			queue.ErrPermanent,
		)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("twitch token: status %d", resp.StatusCode)
	}

	var tok struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil || tok.AccessToken == "" {
		return "", errors.New("twitch token: no access token")
	}

	return tok.AccessToken, nil
}

func (c *Connector) games(
	ctx context.Context,
	client *ingest.Client,
	token string,
	qids map[string]string,
) ([]Payload, error) {
	slugs := make([]string, 0, len(qids))
	for s := range qids {
		slugs = append(slugs, s)
	}

	slices.Sort(slugs)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.APIURL+"/games",
		strings.NewReader(GamesQuery(slugs)),
	)
	if err != nil {
		return nil, fmt.Errorf("build igdb request: %w", err)
	}

	req.Header.Set("Client-ID", c.clientID)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := client.JSON.Do(req)
	if err != nil {
		return nil, fmt.Errorf("igdb games: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusUnauthorized, http.StatusForbidden:
		return nil, fmt.Errorf("igdb rejected the access token: %w", queue.ErrPermanent)
	default:
		return nil, fmt.Errorf("igdb games: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read igdb games: %w", err)
	}

	return ParseGames(body, qids)
}

// Fetch drops unsafe slugs, then queries the rest in batches of Batch with one token per call.
// shortcut: a new token per Fetch call (one per chunk of targets); cache it until expires_in if the token endpoint
// ever rate-limits us.
func (c *Connector) Fetch(
	ctx context.Context, client *ingest.Client, batch []ingest.Target,
) ([]ingest.Record, []ingest.TargetError) {
	safe := slices.DeleteFunc(slices.Clone(batch), func(t ingest.Target) bool { return !safeSlug.MatchString(t.Type) })
	if len(safe) == 0 {
		return nil, nil
	}

	failAll := func(ts []ingest.Target, err error) []ingest.TargetError {
		out := make([]ingest.TargetError, len(ts))
		for i, t := range ts {
			out[i] = ingest.TargetError{Target: t, Err: err}
		}

		return out
	}

	token, err := c.token(ctx, client)
	if err != nil {
		return nil, failAll(safe, err)
	}

	var (
		records []ingest.Record
		errs    []ingest.TargetError
	)

	for chunk := range slices.Chunk(safe, Batch) {
		qids := map[string]string{}
		for _, t := range chunk {
			qids[t.Type] = t.ID
		}

		payloads, err := c.games(ctx, client, token, qids)
		if err != nil {
			errs = append(errs, failAll(chunk, err)...)

			continue
		}

		for _, p := range payloads {
			body, err := json.Marshal(p)
			if err != nil {
				errs = append(errs, ingest.TargetError{Target: ingest.Target{ID: p.Qid, Type: p.Slug}, Err: err})

				continue
			}

			records = append(records, ingest.Record{Source: c.Source(), ExternalID: p.Qid, Payload: body})
		}
	}

	return records, errs
}
