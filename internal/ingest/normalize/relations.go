package normalize

import (
	"slices"
	"strings"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/store"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/wikidata"
)

const (
	// MaxCast is the number of cast credits kept per item (SDD 06 §7.4).
	MaxCast = 15
	// MaxAltTitles caps item_alt_titles per item.
	MaxAltTitles = 20

	wikidataSource = "wikidata"
	relationBased  = "based_on"
)

// role is one SDD 06 §7.4 row: the property, the entity kind it points to and
// the item types it applies to (SDD 01 §5.2 role compatibility, DQ-08).
type role struct {
	name, prop, kind string
	types            []string
	limit            int
}

var (
	screen = []string{"movie", "series"}
	all    = []string{"movie", "series", "game", "book"}
)

// roles is ordered: credits come out role by role in this order.
var roles = []role{
	{"director", "P57", "person", screen, 0},
	{"writer", "P58", "person", screen, 0},
	{"cast", "P161", "person", screen, MaxCast},
	{"creator", "P170", "person", screen, 0},
	{"author", "P50", "person", []string{"book"}, 0},
	{"developer", "P178", "organization", []string{"game"}, 0},
	{"publisher", "P123", "organization", []string{"game", "book"}, 0},
	{"production_company", "P272", "organization", screen, 0},
	{"network", "P449", "organization", []string{"series"}, 0},
	{"platform", "P400", "platform", []string{"game"}, 0},
	{"game_mode", "P404", "game_mode", []string{"game"}, 0},
	{"franchise", "P8345", "franchise", all, 0},
	{"series", "P179", "franchise", all, 0},
	{"genre", "P136", "genre", all, 0},
	{"country", "P495", "country", all, 0},
	{"original_language", "P364", "language", []string{"movie", "series", "game"}, 0},
	{"original_language", "P407", "language", []string{"book"}, 0},
}

// Credit ties an item to a referenced Wikidata entity by QID. Matching it to
// an entities row is P6.B.3.2.
type Credit struct {
	Role     string
	Kind     string
	QID      string
	Name     string
	Position int
}

// AltTitle is one item_alt_titles row.
type AltTitle struct {
	Title  string
	Lang   string
	Kind   string
	Source string
}

// Link is one item_links row, by QID.
type Link struct {
	FromQID  string
	ToQID    string
	Relation string
	Source   string
}

// Relations are the rows derived from one item's Wikidata record.
type Relations struct {
	Credits   []Credit
	AltTitles []AltTitle
	Links     []Link
}

// ExtractRelations derives credits, alternative titles and based_on links of
// qid from its records (pure, order independent). known holds the QIDs of the
// catalogue items, so a based_on link is only emitted between two of them.
func ExtractRelations(qid string, recs []store.Record, known map[string]bool) (*Relations, error) {
	wd, summaries, err := decode(qid, recs)
	if err != nil {
		return nil, err
	}

	return &Relations{
		Credits:   credits(wd),
		AltTitles: altTitles(wd, buildTitles(wd, summaries)),
		Links:     basedOn(qid, wd, known),
	}, nil
}

// values lists the QIDs of a property's usable statements, preferred first,
// without deprecated ones or duplicates.
func values(wd wikidata.Payload, prop string) []string {
	var out []string

	for _, c := range ranked(wd.Claims[prop]) {
		id, _ := c.Value.(string)
		if c.Rank != "deprecated" && id != "" && !slices.Contains(out, id) {
			out = append(out, id)
		}
	}

	return out
}

func credits(wd wikidata.Payload) []Credit {
	var out []Credit

	for _, r := range roles {
		if !slices.Contains(r.types, wd.Type) {
			continue
		}

		for i, id := range values(wd, r.prop) {
			if r.limit > 0 && i >= r.limit {
				break
			}

			out = append(out, Credit{Role: r.name, Kind: r.kind, QID: id, Name: wd.Refs[id], Position: i})
		}
	}

	return out
}

// altTitles lists transliterations (aliases ending in -latn), the original
// title (P1476), former names (P1448, P2561), then the other aliases in label-language order, skipping any text equal to
// a locale title or already listed, up to MaxAltTitles.
func altTitles(wd wikidata.Payload, titles map[string]string) []AltTitle {
	seen := map[string]bool{}
	for _, t := range titles {
		seen[strings.ToLower(strings.TrimSpace(t))] = true
	}

	var (
		out   []AltTitle
		trans []string
	)

	for lang := range wd.Aliases {
		if strings.HasSuffix(lang, "-latn") {
			trans = append(trans, lang)
		}
	}

	slices.Sort(trans)

	addText := func(text, lang, kind string) {
		text = strings.TrimSpace(text)
		key := strings.ToLower(text)

		if text == "" || len(out) >= MaxAltTitles || seen[key] {
			return
		}

		seen[key] = true
		out = append(out, AltTitle{Title: text, Lang: lang, Kind: kind, Source: wikidataSource})
	}

	add := func(lang, kind string) {
		for _, a := range wd.Aliases[lang] {
			addText(a, bcp47(lang), kind)
		}
	}

	for _, lang := range trans {
		add(lang, "transliteration")
	}

	for _, c := range []struct{ prop, kind string }{{"P1476", "original"}, {"P1448", "former"}, {"P2561", "former"}} {
		for _, t := range ranked(wd.Claims[c.prop]) {
			if text, _ := t.Value.(string); t.Rank != "deprecated" {
				addText(text, "", c.kind)
			}
		}
	}

	for _, lang := range wikidata.LabelLangs {
		add(lang, "alias")
	}

	return out
}

// bcp47 formats a Wikidata language code: "mul" (multiple) is no language,
// and a -latn suffix becomes the -Latn script subtag.
func bcp47(lang string) string {
	switch {
	case lang == "mul":
		return ""
	case strings.HasSuffix(lang, "-latn"):
		return strings.TrimSuffix(lang, "-latn") + "-Latn"
	}

	return lang
}

func basedOn(qid string, wd wikidata.Payload, known map[string]bool) []Link {
	if !known[qid] {
		return nil
	}

	var out []Link

	for _, to := range values(wd, "P144") {
		if to != qid && known[to] {
			out = append(out, Link{FromQID: qid, ToQID: to, Relation: relationBased, Source: wikidataSource})
		}
	}

	return out
}
