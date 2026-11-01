// Package trailers checks which candidate trailers YouTube still lets a page embed and stores the verdicts as a
// "youtube" source record per item, which the normaliser reads (SDD 06 §7.5, port of legacy trailers.py).
package trailers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"regexp"
	"slices"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

const (
	// Source is the source_records.source of the verdicts.
	Source = "youtube"
	// OEmbedURL is the YouTube oEmbed endpoint.
	OEmbedURL = "https://www.youtube.com/oembed"
	// MaxTrailers is how many embeddable trailers an item needs; checking stops there.
	MaxTrailers = 2
)

// ErrNoVerdict means YouTube said nothing about the video (429, 5xx, network): it is neither available nor not.
var ErrNoVerdict = errors.New("trailers: no verdict")

var youtubeID = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)

// Claim is one Wikidata P1651 statement of the trimmed payload.
type Claim struct {
	Value string `json:"value"`
	Rank  string `json:"rank"`
}

// Sources are the trailer ids each source lists for one item.
type Sources struct {
	Wikidata []Claim  `json:"wikidata,omitempty"`
	TMDB     []string `json:"tmdb,omitempty"`
	IGDB     []string `json:"igdb,omitempty"`
}

// Checked is the verdict record: the payload of the "youtube" source.
type Checked struct {
	Available   []string `json:"available"`
	Unavailable []string `json:"unavailable"`
}

// Candidates lists the well-formed ids in display order: Wikidata (preferred rank first), then TMDB, then IGDB.
func Candidates(s Sources) []string {
	claims := slices.SortedStableFunc(slices.Values(s.Wikidata), func(a, b Claim) int {
		return b2i(a.Rank != "preferred") - b2i(b.Rank != "preferred")
	})

	var ids []string
	for _, c := range claims {
		ids = append(ids, c.Value)
	}

	ids = slices.Concat(ids, s.TMDB, s.IGDB)

	var out []string

	for _, id := range ids {
		if youtubeID.MatchString(id) && !slices.Contains(out, id) {
			out = append(out, id)
		}
	}

	return out
}

func b2i(b bool) int {
	if b {
		return 1
	}

	return 0
}

// Asker says whether a video can be embedded; ErrNoVerdict when it cannot tell.
type Asker interface {
	Available(ctx context.Context, id string) (bool, error)
}

// Check asks about the ids in order until MaxTrailers are available. Verdicts in known are reused (pass the zero
// value to ask again about every id). A missing verdict aborts: a rate limit says nothing about the video.
func Check(ctx context.Context, a Asker, ids []string, known Checked) (Checked, error) {
	out := Checked{Available: []string{}, Unavailable: []string{}}

	for _, id := range ids {
		if len(out.Available) == MaxTrailers {
			break
		}

		ok := slices.Contains(known.Available, id)
		if !ok && !slices.Contains(known.Unavailable, id) {
			var err error
			if ok, err = a.Available(ctx, id); err != nil {
				return Checked{}, err
			}
		}

		if ok {
			out.Available = append(out.Available, id)
		} else {
			out.Unavailable = append(out.Unavailable, id)
		}
	}

	return out, nil
}

// OEmbed is the Asker over YouTube's oEmbed endpoint.
type OEmbed struct {
	Client *ingest.Client
	// URL overrides OEmbedURL (tests).
	URL string
}

// Available maps 200 to true, 400/401/403/404 (removed, private, embedding off) to false, anything else to ErrNoVerdict.
func (o *OEmbed) Available(ctx context.Context, id string) (bool, error) {
	endpoint := o.URL
	if endpoint == "" {
		endpoint = OEmbedURL
	}

	q := url.Values{"url": {"https://www.youtube.com/watch?v=" + id}, "format": {"json"}}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+q.Encode(), nil)
	if err != nil {
		return false, fmt.Errorf("build oembed request: %w", err)
	}

	resp, err := o.Client.JSON.Do(req)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrNoVerdict, err)
	}

	_ = resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
		return false, nil
	}

	return false, fmt.Errorf("%w: oembed status %d for %s", ErrNoVerdict, resp.StatusCode, id)
}

// Input is what Targets hands to Fetch in Target.Type: the candidates and, unless the run re-checks, the stored verdicts.
type Input struct {
	QID     string   `json:"qid"`
	Sources Sources  `json:"sources"`
	Known   *Checked `json:"known,omitempty"`
}

// Connector is the ingest.Connector of the trailer check, behind the "youtube" flag.
type Connector struct {
	// OEmbedURL overrides the endpoint (tests).
	OEmbedURL string
}

// Source implements ingest.Connector.
func (*Connector) Source() string { return Source }

// shortcut: the aggregate reads the trailer keys of every stored record of the item; fine at 100k items, add a
// partial index on source_records(source, external_id) if the plan shows a scan.
const targetsSQL = `SELECT external_id, jsonb_object_agg(source, CASE source
    WHEN 'wikidata' THEN payload #> '{claims,P1651}'
    WHEN 'youtube' THEN payload
    ELSE payload -> 'trailers' END)
FROM source_records WHERE source IN ('wikidata', 'tmdb', 'igdb', 'youtube') AND external_id > $1
GROUP BY external_id
HAVING $2::timestamptz IS NULL OR max(fetched_at) FILTER (WHERE source <> 'youtube') >= $2
ORDER BY external_id LIMIT NULLIF($3, 0)`

// Targets yields the items that have trailer candidates or an earlier verdict, by QID. A full run asks again about
// every id; the other modes only ask about ids never checked.
func (*Connector) Targets(
	ctx context.Context, db database.Executor, p ingest.RunParams, after string,
) iter.Seq2[ingest.Target, error] {
	return func(yield func(ingest.Target, error) bool) {
		var since any
		if p.Mode == run.ModeIncremental && !p.Since.IsZero() {
			since = p.Since
		}

		rows, err := db.QueryContext(ctx, targetsSQL, after, since, p.Limit)
		if err != nil {
			yield(ingest.Target{}, fmt.Errorf("trailers targets: %w", err))

			return
		}

		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var (
				qid string
				raw []byte
			)

			if err := rows.Scan(&qid, &raw); err != nil {
				yield(ingest.Target{}, fmt.Errorf("scan trailers target: %w", err))

				return
			}

			in, err := decodeInput(qid, raw, p.Mode != run.ModeFull)
			if err != nil {
				yield(ingest.Target{}, err)

				return
			}

			typ, _ := json.Marshal(in)

			if !yield(ingest.Target{ID: qid, Type: string(typ)}, nil) {
				return
			}
		}

		if err := rows.Err(); err != nil {
			yield(ingest.Target{}, fmt.Errorf("trailers targets: %w", err))
		}
	}
}

func decodeInput(qid string, raw []byte, reuse bool) (Input, error) {
	var by struct {
		Wikidata []Claim  `json:"wikidata"`
		TMDB     []string `json:"tmdb"`
		IGDB     []string `json:"igdb"`
		YouTube  *Checked `json:"youtube"`
	}

	if err := json.Unmarshal(raw, &by); err != nil {
		return Input{}, fmt.Errorf("decode trailer sources of %s: %w", qid, err)
	}

	in := Input{QID: qid, Sources: Sources{Wikidata: by.Wikidata, TMDB: by.TMDB, IGDB: by.IGDB}}
	if reuse {
		in.Known = by.YouTube
	}

	// A stored verdict keeps the item in the run even when its candidates vanished, so the record can be emptied.
	if by.YouTube != nil && in.Known == nil {
		in.Known = &Checked{}
	}

	return in, nil
}

// Fetch checks one item at a time (the host limit, not concurrency, bounds oEmbed). An item without candidates and
// without an earlier verdict has no record; one whose check lacks a verdict is reported and keeps its old record.
func (c *Connector) Fetch(
	ctx context.Context, client *ingest.Client, batch []ingest.Target,
) ([]ingest.Record, []ingest.TargetError) {
	var (
		records []ingest.Record
		errs    []ingest.TargetError
	)

	asker := &OEmbed{Client: client, URL: c.OEmbedURL}

	for _, t := range batch {
		if ctx.Err() != nil {
			break
		}

		var in Input
		if err := json.Unmarshal([]byte(t.Type), &in); err != nil {
			errs = append(errs, ingest.TargetError{Target: t, Err: fmt.Errorf("decode trailers input: %w", err)})

			continue
		}

		ids := Candidates(in.Sources)
		if len(ids) == 0 && in.Known == nil {
			continue
		}

		known := Checked{}
		if in.Known != nil {
			known = *in.Known
		}

		res, err := Check(ctx, asker, ids, known)
		if err != nil {
			errs = append(errs, ingest.TargetError{Target: t, Err: err})

			continue
		}

		payload, _ := json.Marshal(struct {
			QID string `json:"qid"`
			Checked
		}{in.QID, res})

		records = append(records, ingest.Record{Source: Source, ExternalID: in.QID, Payload: payload})
	}

	return records, errs
}
