package normalize

import (
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"strings"
)

// Cover is a cover candidate named by a source record; the cover job
// downloads it and sets the storage keys.
type Cover struct {
	URL    string
	Source string
	// Attribution is nil for sources that need none.
	Attribution map[string]string
}

var markup = regexp.MustCompile(`<[^>]+>`)

// ParseCommonsCover reads a Commons imageinfo response (prop=imageinfo with
// extmetadata Artist and LicenseShortName). It returns nil, nil when the file
// has no image info.
func ParseCommonsCover(raw []byte) (*Cover, error) {
	var r struct {
		Query struct {
			Pages []struct {
				ImageInfo []struct {
					URL            string `json:"url"`
					ThumbURL       string `json:"thumburl"`
					DescriptionURL string `json:"descriptionurl"`
					ExtMetadata    map[string]struct {
						Value string `json:"value"`
					} `json:"extmetadata"`
				} `json:"imageinfo"`
			} `json:"pages"`
		} `json:"query"`
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, fmt.Errorf("decode commons response: %w", err)
	}

	if len(r.Query.Pages) == 0 || len(r.Query.Pages[0].ImageInfo) == 0 {
		return nil, nil //nolint:nilnil // no image is not an error
	}

	info := r.Query.Pages[0].ImageInfo[0]
	author := strings.TrimSpace(html.UnescapeString(markup.ReplaceAllString(info.ExtMetadata["Artist"].Value, "")))
	url := info.ThumbURL

	if url == "" {
		url = info.URL
	}

	return &Cover{URL: url, Source: SourceCommons, Attribution: map[string]string{
		"source": "Wikimedia Commons", "license": info.ExtMetadata["LicenseShortName"].Value,
		"author": author, "url": info.DescriptionURL,
	}}, nil
}

// pickCover is the cover a source record names directly: the Open Library
// cover for books, then the official poster or box art (TMDB, IGDB), then the
// Steam and GOG images.
func pickCover(itemType string, s *sources) *Cover {
	if itemType == "book" && len(s.openlibrary.Covers) > 0 {
		return &Cover{
			URL:    fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-L.jpg", s.openlibrary.Covers[0]),
			Source: SourceOpenLibrary,
		}
	}

	official := map[string]struct{ src, name string }{
		"movie": {SourceTMDB, "TMDB"}, "series": {SourceTMDB, "TMDB"}, "game": {SourceIGDB, "IGDB"},
	}[itemType]

	if official.src != "" && s.extraSource == official.src && s.extra.Cover != "" {
		att := map[string]string{"source": official.name}
		if s.extra.URL != "" {
			att["url"] = s.extra.URL
		}

		return &Cover{URL: s.extra.Cover, Source: official.src, Attribution: att}
	}

	if itemType == "game" {
		if s.steam.HeaderImage != "" {
			return &Cover{URL: s.steam.HeaderImage, Source: SourceSteam}
		}

		if s.gog.Cover != "" {
			return &Cover{URL: s.gog.Cover, Source: SourceGOG}
		}
	}

	return nil
}
