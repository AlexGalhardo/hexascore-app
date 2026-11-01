// Command capture freezes the parity corpus (SDD 06 section 7.9): it selects 200 QIDs (50 per type) from
// Wikidata and writes qids.txt plus one records/<QID>.json per QID with the source records the normaliser reads.
// Run it once and commit the output; the parity test never touches the network.
//
//	go run ./scripts/parity/capture -out testdata/parity
package main

import (
	"cmp"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/httpx"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/wikidata"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/wikipedia"
)

const perType = 50

type pick struct {
	qid, typ, tier string
}

// special is a pool of candidates with a property the corpus must cover, SDD 06 section 7.9.
type special struct {
	tier    string
	want    int
	pattern string
}

type recordSet map[string]map[string]json.RawMessage

var (
	// Labels in a non-Latin script (cyrillic, arabic, kana, han).
	nonLatin = special{
		"non-latin",
		3,
		`?item rdfs:label ?l . FILTER(lang(?l) IN ("en","mul") && REGEX(?l, "[Ѐ-ӿ؀-ۿ぀-ヿ一-鿿]"))`,
	}
	// Release date known to the year only (time precision 9).
	yearOnly = special{"year-precision", 3, `?item p:P577/psv:P577 ?t . ?t wikibase:timePrecision 9 .`}

	specials = map[string][]special{
		"movie":  {nonLatin, yearOnly},
		"series": {nonLatin, yearOnly, {"p580", 4, `?item wdt:P580 [] .`}},
		"game": {
			nonLatin,
			yearOnly,
			{"steam-only", 4, `?item wdt:P1733 [] . FILTER NOT EXISTS { ?item wdt:P5794 [] }`},
		},
		"book": {
			nonLatin,
			yearOnly,
			{"isbn10-only", 4, `?item wdt:P957 [] . FILTER NOT EXISTS { ?item wdt:P212 [] }`},
		},
	}
)

func main() {
	out := flag.String("out", "testdata/parity", "output directory")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Minute)
	defer cancel()

	c := httpx.New(httpx.Config{ContactURL: "https://hexascore.app", Hosts: httpx.DefaultHosts()})
	wd := wikidata.New(c)

	var all []pick

	records := recordSet{}

	for _, typ := range wikidata.Types {
		picks, err := selectType(ctx, c, wd, typ, records)
		if err != nil {
			log.Fatalf("%s: %v", typ, err)
		}

		log.Printf("%s: %d picked", typ, len(picks))
		all = append(all, picks...)
	}

	if err := fetchWikipedia(ctx, c, all, records); err != nil {
		log.Fatal(err)
	}

	if err := write(*out, all, records); err != nil {
		log.Fatal(err)
	}
}

func values(typ string) string {
	var b strings.Builder
	for _, cl := range wikidata.TypeClasses[typ] {
		b.WriteString("wd:" + cl + " ")
	}

	return b.String()
}

func sparql(ctx context.Context, c *ingest.Client, q string) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		wikidata.SparqlURL+"?"+url.Values{"query": {q}, "format": {"json"}}.Encode(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/sparql-results+json")

	resp, err := c.JSON.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sparql status %d for %.80q", resp.StatusCode, q)
	}

	var doc struct {
		Results struct {
			Bindings []struct {
				Item struct{ Value string }
			}
		}
	}
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, err
	}

	var qids []string

	for _, b := range doc.Results.Bindings {
		qids = append(qids, b.Item.Value[strings.LastIndex(b.Item.Value, "/")+1:])
	}

	return qids, nil
}

func query(typ, extra, filter, tail string) string {
	return fmt.Sprintf(`SELECT DISTINCT ?item ?links WHERE {
  VALUES ?class { %s }
  ?item wdt:P31 ?class ; wikibase:sitelinks ?links .
  %s
  %s
} %s`, values(typ), extra, filter, tail)
}

// selectType returns the 50 picks of a type and stores their Wikidata records.
func selectType(
	ctx context.Context,
	c *ingest.Client,
	wd *wikidata.Connector,
	typ string,
	records recordSet,
) ([]pick, error) {
	var picks []pick

	taken := map[string]bool{}

	// take fetches candidates and keeps those the legacy normaliser can build (an English title) and not taken yet.
	take := func(tier string, want int, qids []string) error {
		got := 0

		for chunk := range slices.Chunk(qids, 50) {
			targets := make([]ingest.Target, len(chunk))
			for i, q := range chunk {
				targets[i] = ingest.Target{ID: q, Type: typ}
			}

			recs, _ := wd.Fetch(ctx, c, targets)

			for _, r := range recs {
				if r.Source != "wikidata" || taken[r.ExternalID] || got == want {
					continue
				}

				var p wikidata.Payload
				if json.Unmarshal(r.Payload, &p) != nil || !hasEnglish(p) {
					continue
				}

				taken[r.ExternalID] = true
				records[r.ExternalID] = map[string]json.RawMessage{"wikidata": r.Payload}
				picks = append(picks, pick{r.ExternalID, typ, tier})
				got++
			}

			if got == want {
				break
			}
		}

		if got < want {
			return fmt.Errorf("tier %s: %d of %d", tier, got, want)
		}

		return nil
	}

	for _, s := range specials[typ] {
		var qids []string

		var err error

		// The WDQS timeout hits the heavier special patterns now and then; a retry usually passes.
		for range 3 {
			if qids, err = sparql(ctx, c, query(typ, s.pattern, "FILTER(?links >= 8)", "LIMIT 60")); err == nil {
				break
			}
		}

		if err != nil {
			log.Printf("%s: special %s skipped: %v", typ, s.tier, err)

			continue
		}

		// A thin special pool (e.g. non-Latin game titles) is reported and the shortfall goes to the plain tiers.
		if err := take(s.tier, s.want, qids); err != nil {
			log.Printf("%s: %v", typ, err)
		}
	}

	rest := perType - len(picks)
	third := rest / 3

	for _, t := range []struct {
		tier   string
		want   int
		filter string
		tail   string
	}{
		{"top", third, "", "ORDER BY DESC(?links) LIMIT 120"},
		{"mid", third, "FILTER(?links >= 25 && ?links <= 35)", "LIMIT 150"},
		{"tail", rest - 2*third, "FILTER(?links >= 6 && ?links <= 9)", "LIMIT 200"},
	} {
		qids, err := sparql(ctx, c, query(typ, "", t.filter, t.tail))
		if err != nil {
			return nil, err
		}

		if err := take(t.tier, t.want, qids); err != nil {
			return nil, err
		}
	}

	return picks, nil
}

func hasEnglish(p wikidata.Payload) bool {
	return p.Labels["en"] != "" || p.Labels["mul"] != "" || p.Sitelinks["enwiki"] != ""
}

func fetchWikipedia(ctx context.Context, c *ingest.Client, all []pick, records recordSet) error {
	for _, lang := range wikipedia.Langs {
		w, err := wikipedia.New(lang)
		if err != nil {
			return err
		}

		var targets []ingest.Target

		for _, p := range all {
			var wp wikidata.Payload
			if err := json.Unmarshal(records[p.qid]["wikidata"], &wp); err != nil {
				return err
			}

			if title := wp.Sitelinks[lang+"wiki"]; title != "" {
				targets = append(targets, ingest.Target{ID: p.qid, Type: title})
			}
		}

		recs, errs := w.Fetch(ctx, c, targets)
		for _, e := range errs {
			log.Printf("wikipedia:%s %s: %v", lang, e.Target.ID, e.Err)
		}

		for _, r := range recs {
			records[r.ExternalID][w.Source()] = r.Payload
		}
	}

	return nil
}

func write(dir string, all []pick, records recordSet) error {
	slices.SortFunc(all, func(a, b pick) int {
		return cmp.Or(cmp.Compare(a.typ, b.typ), cmp.Compare(a.qid, b.qid))
	})

	if err := os.MkdirAll(filepath.Join(dir, "records"), 0o750); err != nil {
		return err
	}

	var list strings.Builder

	for _, p := range all {
		fmt.Fprintf(&list, "%s %s %s\n", p.qid, p.typ, p.tier)

		doc, err := json.MarshalIndent(map[string]any{
			"qid": p.qid, "type": p.typ, "tier": p.tier, "records": records[p.qid],
		}, "", "  ")
		if err != nil {
			return err
		}

		if err := os.WriteFile(filepath.Join(dir, "records", p.qid+".json"), append(doc, '\n'), 0o600); err != nil {
			return err
		}
	}

	return os.WriteFile(filepath.Join(dir, "qids.txt"), []byte(list.String()), 0o600)
}
