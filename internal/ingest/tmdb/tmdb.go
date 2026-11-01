// Package tmdb is the TMDB connector (SDD 06 §7.3): display-only poster and facts for movies and series. Ids are
// resolved only through /find/{imdb_id}, never by title. TMDB data is never commercial and never leaves the site
// audience (ADR 0006, D-03). Port of legacy connectors/tmdb.py.
package tmdb

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
	"strconv"
	"strings"
	"sync"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

// APIURL is the v3 API; the v4 read access token is sent as a Bearer token.
const APIURL = "https://api.themoviedb.org/3"

// PosterBase is the poster URL prefix; w780 is the smallest poster size above the 600 px the cover pipeline keeps.
const PosterBase = "https://image.tmdb.org/t/p/w780"

// MaxTrailers caps the trailer keys kept per item.
const MaxTrailers = 3

// Workers caps concurrent targets inside one Fetch; the host limits of httpx still bound the rate.
const Workers = 8

var (
	_ ingest.Connector = (*Connector)(nil)

	imdbID = regexp.MustCompile(`^tt\d+$`)
	date   = regexp.MustCompile(`^\d{4}-\d\d-\d\d$`)
)

// Connector fetches TMDB details for stored items. A Target carries the Wikidata QID in ID and Ref(...) in Type.
type Connector struct {
	token string
	// APIURL overrides APIURL; tests only.
	APIURL string
}

// New returns a Connector that authenticates with token.
func New(token string) *Connector { return &Connector{token: token, APIURL: APIURL} }

// Source names the connector in source_records.
func (*Connector) Source() string { return "tmdb" }

// Ref packs what Fetch needs besides the QID into Target.Type: the item type and the ids Wikidata knows.
func Ref(itemType, tmdbID, imdb string) string { return itemType + "|" + tmdbID + "|" + imdb }

func unpack(ref string) (itemType, tmdbID, imdb string) {
	parts := strings.SplitN(ref, "|", 3)
	for len(parts) < 3 {
		parts = append(parts, "")
	}

	return parts[0], parts[1], parts[2]
}

func mediaOf(itemType string) string {
	if itemType == "series" {
		return "tv"
	}

	return "movie"
}

// Payload is the stored record body. Overview is the licensed input of the synopsis step and is never copied
// into item_translations (FR-06-3).
type Payload struct {
	Qid         string   `json:"qid"`
	ID          int      `json:"id"`
	URL         string   `json:"url"`
	Cover       string   `json:"cover"`
	Overview    string   `json:"overview"`
	Runtime     int      `json:"runtime"`
	Genres      []string `json:"genres"`
	ReleaseDate string   `json:"release_date"`
	Seasons     int      `json:"seasons"`
	Episodes    int      `json:"episodes"`
	Creators    []string `json:"creators"`
	Countries   []string `json:"countries"`
	Trailers    []string `json:"trailers"`
}

type named struct {
	Name string `json:"name"`
}

type video struct {
	Key      string `json:"key"`
	Site     string `json:"site"`
	Type     string `json:"type"`
	Official bool   `json:"official"`
}

// ParseFind returns the TMDB id of the first result of the media type ("movie" or "tv"), or zero.
func ParseFind(raw []byte, media string) (int, error) {
	var doc struct {
		Movies []struct {
			ID int `json:"id"`
		} `json:"movie_results"`
		TV []struct {
			ID int `json:"id"`
		} `json:"tv_results"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return 0, fmt.Errorf("decode tmdb find: %w", err)
	}

	if media == "movie" && len(doc.Movies) > 0 {
		return doc.Movies[0].ID, nil
	}

	if media != "movie" && len(doc.TV) > 0 {
		return doc.TV[0].ID, nil
	}

	return 0, nil
}

// ParseDetails turns a movie or tv details response (with credits and videos appended) into the payload.
func ParseDetails(raw []byte, media, qid string) (Payload, error) {
	var d struct {
		ID               int     `json:"id"`
		Overview         string  `json:"overview"`
		PosterPath       string  `json:"poster_path"`
		Runtime          int     `json:"runtime"`
		EpisodeRunTime   []int   `json:"episode_run_time"`
		Genres           []named `json:"genres"`
		ReleaseDate      string  `json:"release_date"`
		FirstAirDate     string  `json:"first_air_date"`
		NumberOfSeasons  int     `json:"number_of_seasons"`
		NumberOfEpisodes int     `json:"number_of_episodes"`
		CreatedBy        []named `json:"created_by"`
		Countries        []named `json:"production_countries"`
		Credits          struct {
			Crew []struct {
				Name string `json:"name"`
				Job  string `json:"job"`
			} `json:"crew"`
		} `json:"credits"`
		Videos struct {
			Results []video `json:"results"`
		} `json:"videos"`
	}
	if err := json.Unmarshal(raw, &d); err != nil {
		return Payload{}, fmt.Errorf("decode tmdb details %s: %w", qid, err)
	}

	if d.ID == 0 {
		return Payload{}, fmt.Errorf("decode tmdb details %s: no id", qid)
	}

	p := Payload{
		Qid: qid, ID: d.ID, URL: fmt.Sprintf("https://www.themoviedb.org/%s/%d", media, d.ID),
		Overview: d.Overview, Genres: names(d.Genres), Countries: names(d.Countries),
		Seasons: d.NumberOfSeasons, Episodes: d.NumberOfEpisodes,
	}

	if d.PosterPath != "" {
		p.Cover = PosterBase + d.PosterPath
	}

	released := d.ReleaseDate
	people := []string{}

	if media == "movie" {
		p.Runtime = d.Runtime

		for _, c := range d.Credits.Crew {
			if c.Job == "Director" {
				people = append(people, c.Name)
			}
		}
	} else {
		released = d.FirstAirDate
		if len(d.EpisodeRunTime) > 0 {
			p.Runtime = d.EpisodeRunTime[0]
		}

		people = names(d.CreatedBy)
	}

	if date.MatchString(released) {
		p.ReleaseDate = released
	}

	p.Creators = unique(people)
	p.Trailers = trailers(d.Videos.Results)

	return p, nil
}

// trailers keeps the YouTube trailers, official first (TMDB's own order breaks ties), at most MaxTrailers.
func trailers(videos []video) []string {
	var keep []video

	for _, v := range videos {
		if v.Site == "YouTube" && v.Type == "Trailer" {
			keep = append(keep, v)
		}
	}

	slices.SortStableFunc(keep, func(a, b video) int {
		switch {
		case a.Official == b.Official:
			return 0
		case a.Official:
			return -1
		default:
			return 1
		}
	})

	keys := make([]string, len(keep))
	for i, v := range keep {
		keys[i] = v.Key
	}

	return unique(keys)[:min(MaxTrailers, len(unique(keys)))]
}

func names(in []named) []string {
	out := make([]string, len(in))
	for i, n := range in {
		out[i] = n.Name
	}

	return out
}

func unique(in []string) []string {
	out := make([]string, 0, len(in))

	for _, s := range in {
		if !slices.Contains(out, s) {
			out = append(out, s)
		}
	}

	return out
}

// targetsSQL: $1 after, $2 since (null = any), $3 only missing (skip QIDs that already have a tmdb record),
// $4 limit (0 = none).
const targetsSQL = `
SELECT i.wikidata_qid, i.type, coalesce(i.external_ids->>'tmdb', ''), coalesce(i.external_ids->>'imdb', '')
FROM items i
WHERE i.type IN ('movie', 'series')
  AND (i.external_ids ? 'tmdb' OR i.external_ids ? 'imdb')
  AND i.wikidata_qid > $1
  AND ($2::timestamptz IS NULL OR i.updated_at >= $2)
  AND (NOT $3::boolean OR NOT EXISTS (
    SELECT 1 FROM source_records s WHERE s.source = 'tmdb' AND s.external_id = i.wikidata_qid))
ORDER BY i.wikidata_qid
LIMIT CASE WHEN $4::int > 0 THEN $4::int END`

// Targets yields the movies and series that carry a TMDB or IMDb id, after the cursor. Bootstrap takes only the
// ones without a record, incremental those updated since p.Since, full all of them.
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
			yield(ingest.Target{}, fmt.Errorf("tmdb targets: %w", err))

			return
		}

		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var qid, itemType, tmdbID, imdb string
			if err := rows.Scan(&qid, &itemType, &tmdbID, &imdb); err != nil {
				yield(ingest.Target{}, fmt.Errorf("scan tmdb target: %w", err))

				return
			}

			if !yield(ingest.Target{ID: qid, Type: Ref(itemType, tmdbID, imdb)}, nil) {
				return
			}
		}

		if err := rows.Err(); err != nil {
			yield(ingest.Target{}, fmt.Errorf("tmdb targets: %w", err))
		}
	}
}

// get returns the body of a v3 GET. found is false for a 404. A 401 is permanent: a rejected token must stop the
// run, not log one failure per item.
func (c *Connector) get(
	ctx context.Context, client *ingest.Client, path string, q url.Values,
) (body []byte, found bool, err error) {
	if c.token == "" {
		return nil, false, fmt.Errorf("tmdb: THEMOVIEDATABASE_API_READ_ACCESS_TOKEN is not set: %w", queue.ErrPermanent)
	}

	endpoint := c.APIURL
	if endpoint == "" {
		endpoint = APIURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+path+"?"+q.Encode(), nil)
	if err != nil {
		return nil, false, fmt.Errorf("build tmdb request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := client.JSON.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("tmdb %s: %w", path, err)
	}

	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusUnauthorized:
		return nil, false, fmt.Errorf("tmdb rejected the access token: %w", queue.ErrPermanent)
	case http.StatusNotFound:
		return nil, false, nil
	default:
		return nil, false, fmt.Errorf("tmdb %s: status %d", path, resp.StatusCode)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, false, fmt.Errorf("read tmdb %s: %w", path, err)
	}

	return body, true, nil
}

// one resolves a single target; a nil record means TMDB has nothing for it.
func (c *Connector) one(ctx context.Context, client *ingest.Client, t ingest.Target) (*ingest.Record, error) {
	itemType, tmdbID, imdb := unpack(t.Type)
	m := mediaOf(itemType)

	if _, err := strconv.Atoi(tmdbID); err != nil {
		// Wikidata has no TMDB id: resolve it through the IMDb id (exact match, never by title).
		if !imdbID.MatchString(imdb) {
			return nil, nil
		}

		raw, found, err := c.get(ctx, client, "/find/"+imdb, url.Values{"external_source": {"imdb_id"}})
		if err != nil || !found {
			return nil, err
		}

		id, err := ParseFind(raw, m)
		if err != nil || id == 0 {
			return nil, err
		}

		tmdbID = strconv.Itoa(id)
	}

	appended := "videos"
	if m == "movie" {
		appended = "credits,videos"
	}

	raw, found, err := c.get(ctx, client, "/"+m+"/"+tmdbID, url.Values{"append_to_response": {appended}})
	if err != nil || !found {
		return nil, err
	}

	p, err := ParseDetails(raw, m, t.ID)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("encode tmdb %s: %w", t.ID, err)
	}

	return &ingest.Record{Source: c.Source(), ExternalID: t.ID, Payload: body}, nil
}

// Fetch resolves the batch on a small pool, keeping the order of batch. Targets TMDB does not know are skipped;
// a permanent error (rejected token) cancels the rest of the batch.
func (c *Connector) Fetch(
	ctx context.Context, client *ingest.Client, batch []ingest.Target,
) ([]ingest.Record, []ingest.TargetError) {
	type result struct {
		rec *ingest.Record
		err error
	}

	results := make([]result, len(batch))
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	var (
		wg  sync.WaitGroup
		sem = make(chan struct{}, Workers)
	)

	for i, t := range batch {
		sem <- struct{}{}

		wg.Add(1)

		go func() {
			defer func() { <-sem; wg.Done() }()

			rec, err := c.one(ctx, client, t)
			results[i] = result{rec, err}

			if errors.Is(err, queue.ErrPermanent) {
				cancel()
			}
		}()
	}

	wg.Wait()

	var (
		records []ingest.Record
		errs    []ingest.TargetError
	)

	for i, r := range results {
		switch {
		case r.err != nil:
			errs = append(errs, ingest.TargetError{Target: batch[i], Err: r.err})
		case r.rec != nil:
			records = append(records, *r.rec)
		}
	}

	return records, errs
}
