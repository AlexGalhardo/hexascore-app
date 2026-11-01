package covers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/normalize"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/store"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/wikidata"
)

// CommonsAPI is the Wikimedia Commons API the P3383 posters are looked up on.
const CommonsAPI = "https://commons.wikimedia.org/w/api.php"

// CommonsFile is the Commons file of the film poster (P3383) of qid, or "". P18 is any image of the subject (an
// arcade cabinet, an author portrait), so it is never a cover.
func CommonsFile(qid string, recs []store.Record) string {
	for _, r := range recs {
		if r.Source != normalize.SourceWikidata || r.ExternalID != qid {
			continue
		}

		var wd wikidata.Payload
		if json.Unmarshal(r.Payload, &wd) != nil {
			return ""
		}

		for _, c := range wd.Claims["P3383"] {
			if v, ok := c.Value.(string); ok && v != "" {
				return v
			}
		}
	}

	return ""
}

// commonsCover asks Commons for the 600 px rendition, licence and author of file.
func (j *Job) commonsCover(ctx context.Context, file string) (*normalize.Cover, error) {
	q := url.Values{
		"action": {"query"}, "titles": {"File:" + file}, "prop": {"imageinfo"},
		"iiprop": {"url|extmetadata"}, "iiextmetadatafilter": {"Artist|LicenseShortName"},
		"iiurlwidth": {"600"}, "format": {"json"}, "formatversion": {"2"},
	}

	api := j.CommonsAPI
	if api == "" {
		api = CommonsAPI
	}

	body, err := j.get(ctx, api+"?"+q.Encode())
	if err != nil {
		return nil, fmt.Errorf("commons %s: %w", file, err)
	}

	return normalize.ParseCommonsCover(body)
}

// get downloads u; anything but 200 is an error. The client caps the body (MaxSourceBytes on the image client).
func (j *Job) get(ctx context.Context, u string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := j.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: status %d", req.URL.Redacted(), resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxSourceBytes+1))
	if err != nil {
		return nil, err
	}

	if len(body) > MaxSourceBytes {
		return nil, fmt.Errorf("%w: over %d bytes", ErrTooLarge, MaxSourceBytes)
	}

	return body, nil
}
