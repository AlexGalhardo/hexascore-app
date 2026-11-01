// Package steam is the Steam storefront connector (SDD 06 §7.3): header image, OS platforms and the store link for
// games with a Steam app id (Wikidata P1733). Port of legacy connectors/steam.py.
//
// Terms: Steam Web API Terms of Use (https://steamcommunity.com/dev/apiterms, read 2026-11-01: data "as is", no
// implied endorsement by Valve). Owner cleared the read in Q-A16. Images stay site-only (SDD 06 §7.1); the app id
// comes from Wikidata and is used for links.
package steam

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/http"
	"net/url"
	"regexp"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

// APIURL is the public storefront appdetails endpoint.
const APIURL = "https://store.steampowered.com/api/appdetails"

var (
	_ ingest.Connector = (*Connector)(nil)

	appID = regexp.MustCompile(`^\d+$`)
)

// Payload is the stored record body.
type Payload struct {
	Qid         string   `json:"qid"`
	AppID       string   `json:"appid"`
	Name        string   `json:"name"`
	HeaderImage string   `json:"header_image,omitempty"`
	Platforms   []string `json:"platforms"`
	StoreURL    string   `json:"store_url"`
}

// ParseAppdetails reads the envelope of appid; nil without error when Steam reports no success.
func ParseAppdetails(raw []byte, appid, qid string) (*Payload, error) {
	var doc map[string]struct {
		Success bool `json:"success"`
		Data    *struct {
			Name        string `json:"name"`
			HeaderImage string `json:"header_image"`
			Platforms   struct {
				Windows bool `json:"windows"`
				Mac     bool `json:"mac"`
				Linux   bool `json:"linux"`
			} `json:"platforms"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("decode steam appdetails %s: %w", appid, err)
	}

	env, ok := doc[appid]
	if !ok || !env.Success || env.Data == nil {
		return nil, nil
	}

	d := env.Data
	p := &Payload{
		Qid: qid, AppID: appid, Name: d.Name, HeaderImage: d.HeaderImage, Platforms: []string{},
		StoreURL: "https://store.steampowered.com/app/" + appid + "/",
	}

	for _, os := range []struct {
		on   bool
		name string
	}{{d.Platforms.Windows, "Windows"}, {d.Platforms.Mac, "macOS"}, {d.Platforms.Linux, "Linux"}} {
		if os.on {
			p.Platforms = append(p.Platforms, os.name)
		}
	}

	return p, nil
}

// Connector fetches Steam appdetails for stored games. A Target carries the QID in ID and the app id in Type.
type Connector struct {
	// APIURL overrides APIURL; tests only.
	APIURL string
}

// Source names the connector in source_records.
func (*Connector) Source() string { return "steam" }

// targetsSQL: $1 after, $2 since (null = any), $3 only missing, $4 limit (0 = none).
const targetsSQL = `
SELECT i.wikidata_qid, i.external_ids->>'steam'
FROM items i
WHERE i.type = 'game' AND i.external_ids ? 'steam'
  AND i.wikidata_qid > $1
  AND ($2::timestamptz IS NULL OR i.updated_at >= $2)
  AND (NOT $3::boolean OR NOT EXISTS (
    SELECT 1 FROM source_records s WHERE s.source = 'steam' AND s.external_id = i.wikidata_qid))
ORDER BY i.wikidata_qid
LIMIT CASE WHEN $4::int > 0 THEN $4::int END`

// Targets yields the games with a numeric Steam app id, after the cursor.
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
			yield(ingest.Target{}, fmt.Errorf("steam targets: %w", err))

			return
		}

		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var qid, id string
			if err := rows.Scan(&qid, &id); err != nil {
				yield(ingest.Target{}, fmt.Errorf("scan steam target: %w", err))

				return
			}

			if !appID.MatchString(id) {
				continue
			}

			if !yield(ingest.Target{ID: qid, Type: id}, nil) {
				return
			}
		}

		if err := rows.Err(); err != nil {
			yield(ingest.Target{}, fmt.Errorf("steam targets: %w", err))
		}
	}
}

func (c *Connector) one(ctx context.Context, client *ingest.Client, t ingest.Target) (*ingest.Record, error) {
	endpoint := c.APIURL
	if endpoint == "" {
		endpoint = APIURL
	}

	q := url.Values{"appids": {t.Type}, "l": {"english"}, "cc": {"us"}}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+q.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("build steam request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := client.JSON.Do(req)
	if err != nil {
		return nil, fmt.Errorf("steam %s: %w", t.Type, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("steam %s: status %d", t.Type, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read steam %s: %w", t.Type, err)
	}

	p, err := ParseAppdetails(body, t.Type, t.ID)
	if err != nil || p == nil {
		return nil, err
	}

	raw, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("encode steam %s: %w", t.ID, err)
	}

	return &ingest.Record{Source: c.Source(), ExternalID: t.ID, Payload: raw}, nil
}

// Fetch resolves the batch one app at a time (1 request per 1.6 s, SDD 06 §7.2). An app Steam reports no success
// for is skipped; a failed request is kept for the dead-letter table and the rest of the batch goes on.
func (c *Connector) Fetch(
	ctx context.Context, client *ingest.Client, batch []ingest.Target,
) ([]ingest.Record, []ingest.TargetError) {
	var (
		records []ingest.Record
		errs    []ingest.TargetError
	)

	for _, t := range batch {
		rec, err := c.one(ctx, client, t)

		switch {
		case err != nil:
			errs = append(errs, ingest.TargetError{Target: t, Err: err})
		case rec != nil:
			records = append(records, *rec)
		}
	}

	return records, errs
}
