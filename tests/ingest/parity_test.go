//go:build parity

package ingest_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"golang.org/x/text/unicode/norm"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/normalize"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/store"
)

const parityDir = "../../testdata/parity"

// allowed lists every deliberate Go-versus-legacy difference (SDD 06 section 7.9). Keys are
// "<field>" for a whole field or "<field>@<QID>" for one item; the value is the reason.
// A diff with no entry here fails the test.
var allowed = map[string]string{}

type pyItem struct {
	QID         string            `json:"qid"`
	Type        string            `json:"type"`
	ReleaseDate *string           `json:"release_date"`
	ExternalIDs map[string]string `json:"external_ids"`
	Meta        struct {
		Genres         []string `json:"genres"`
		Platforms      []string `json:"platforms"`
		Creators       []string `json:"creators"`
		Countries      []string `json:"countries"`
		RuntimeMinutes int      `json:"runtimeMinutes"`
		Pages          int      `json:"pages"`
		Seasons        int      `json:"seasons"`
		Episodes       int      `json:"episodes"`
		Trailers       []struct {
			YouTubeID string `json:"youtubeId"`
		} `json:"trailers"`
	} `json:"meta"`
	Translations []struct {
		Locale   string  `json:"locale"`
		Title    string  `json:"title"`
		Synopsis *string `json:"synopsis"`
	} `json:"translations"`
}

type pyOut struct {
	QID  string  `json:"qid"`
	Item *pyItem `json:"item"`
	Slug string  `json:"slug"`
}

type frozen struct {
	QID     string                     `json:"qid"`
	Records map[string]json.RawMessage `json:"records"`
}

func readJSON(t *testing.T, path string, v any) {
	t.Helper()

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal(raw, v); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
}

// synopsis is the comparison form of an extract: trimmed, whitespace collapsed, NFC.
func synopsis(s string) string {
	return norm.NFC.String(strings.Join(strings.Fields(s), " "))
}

func diffStrings(field string, goV, pyV []string) string {
	if slices.Equal(goV, pyV) {
		return ""
	}

	return fmt.Sprintf("%s: go %q py %q", field, goV, pyV)
}

func diffInt(field string, goV, pyV int) string {
	if goV == pyV {
		return ""
	}

	return fmt.Sprintf("%s: go %d py %d", field, goV, pyV)
}

// compare returns one "<field>: ..." line per semantic difference of one QID.
func compare(qid string, goItem *normalize.Item, goErr error, py pyOut) []string {
	if py.Item == nil {
		if goErr == nil {
			return []string{"build: go built an item, legacy skipped"}
		}

		return nil
	}

	if goErr != nil {
		return []string{"build: go skipped (" + goErr.Error() + "), legacy built an item"}
	}

	var out []string

	add := func(d string) {
		if d != "" {
			out = append(out, d)
		}
	}

	if goItem.Type != py.Item.Type {
		add(fmt.Sprintf("type: go %q py %q", goItem.Type, py.Item.Type))
	}

	goDate := ""
	if goItem.ReleaseDate != nil {
		goDate = goItem.ReleaseDate.Format(time.DateOnly)
	}

	pyDate := ""
	if py.Item.ReleaseDate != nil {
		pyDate = *py.Item.ReleaseDate
	}

	if goDate != pyDate {
		add(fmt.Sprintf("release_date: go %q py %q", goDate, pyDate))
	}

	year := 0
	if goItem.ReleaseDate != nil {
		year = goItem.ReleaseDate.Year()
	}

	if got := normalize.SlugCandidates(goItem.Titles["en"], year, qid)[0]; got != py.Slug {
		add(fmt.Sprintf("slug: go %q py %q", got, py.Slug))
	}

	if !mapsEqual(goItem.ExternalIDs, py.Item.ExternalIDs) {
		add(fmt.Sprintf("external_ids: go %v py %v", goItem.ExternalIDs, py.Item.ExternalIDs))
	}

	m, pm := goItem.Meta, py.Item.Meta

	add(diffStrings("genres", m.Genres, pm.Genres))
	add(diffStrings("platforms", m.Platforms, pm.Platforms))
	add(diffStrings("creators", m.Creators, pm.Creators))
	add(diffStrings("countries", m.Countries, pm.Countries))
	add(diffInt("runtimeMinutes", m.RuntimeMinutes, pm.RuntimeMinutes))
	add(diffInt("pages", m.Pages, pm.Pages))
	add(diffInt("seasons", m.Seasons, pm.Seasons))
	add(diffInt("episodes", m.Episodes, pm.Episodes))

	pyTrailers := make([]string, len(pm.Trailers))
	for i, tr := range pm.Trailers {
		pyTrailers[i] = tr.YouTubeID
	}

	add(diffStrings("trailers", m.Trailers, pyTrailers))

	pyTitles := map[string]string{}
	pySyn := map[string]string{}

	for _, tr := range py.Item.Translations {
		pyTitles[tr.Locale] = tr.Title
		if tr.Synopsis != nil {
			pySyn[tr.Locale] = synopsis(*tr.Synopsis)
		}
	}

	if !mapsEqual(goItem.Titles, pyTitles) {
		add(fmt.Sprintf("titles: go %v py %v", goItem.Titles, pyTitles))
	}

	goSyn := map[string]string{}

	for loc, s := range goItem.Synopses {
		goSyn[loc] = synopsis(s.Text)
	}

	if !mapsEqual(goSyn, pySyn) {
		add(fmt.Sprintf("synopsis: go %v py %v", goSyn, pySyn))
	}

	return out
}

func mapsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}

	for k, v := range a {
		if bv, ok := b[k]; !ok || bv != v {
			return false
		}
	}

	return true
}

func TestParityLegacy(t *testing.T) {
	files, err := filepath.Glob(filepath.Join(parityDir, "records", "*.json"))
	if err != nil || len(files) != 200 {
		t.Fatalf("want 200 frozen records, got %d (%v)", len(files), err)
	}

	now := time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC)
	diffs, accepted := 0, 0

	for _, file := range files {
		var fr frozen
		readJSON(t, file, &fr)

		var py pyOut
		readJSON(t, filepath.Join(parityDir, "python", fr.QID+".json"), &py)

		recs := make([]store.Record, 0, len(fr.Records))
		for src, payload := range fr.Records {
			recs = append(recs, store.Record{Source: src, ExternalID: fr.QID, Payload: payload})
		}

		item, buildErr := normalize.BuildItem(fr.QID, recs, now)

		for _, d := range compare(fr.QID, item, buildErr, py) {
			field, _, _ := strings.Cut(d, ":")
			if reason, ok := allowed[field+"@"+fr.QID]; ok {
				accepted++
				t.Logf("allowed %s %s (%s)", fr.QID, d, reason)

				continue
			}

			if reason, ok := allowed[field]; ok {
				accepted++
				t.Logf("allowed %s %s (%s)", fr.QID, d, reason)

				continue
			}

			diffs++

			t.Errorf("%s %s", fr.QID, d)
		}
	}

	t.Logf("%d semantic diffs / %d (%d allowed)", diffs, len(files), accepted)
}
