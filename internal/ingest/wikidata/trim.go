package wikidata

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
)

// Trimming is load-bearing (SDD 06 §7.3): adding a property here changes
// every stored hash and needs a full Wikidata run (about 1 h).

// LabelLangs are the label and alias languages kept; "mul" holds labels
// valid in every language and is the fallback for names.
var LabelLangs = []string{"en", "mul", "pt-br", "pt", "es"}

// SitelinkSites are the sitelinks kept by name; the total is kept as a count.
var SitelinkSites = []string{"enwiki", "ptwiki", "eswiki"}

// RefProps hold entity ids whose labels are resolved into Payload.Refs and
// whose entities are stored as wikidata:entity records. Beyond the legacy
// set: P144 (based on, FR-06-11) and the SDD 06 §7.4 role properties.
var RefProps = []string{
	"P136", "P400", "P57", "P170", "P178", "P50", "P495",
	"P144", "P58", "P161", "P272", "P449", "P404", "P179", "P8345", "P364", "P407", "P123",
}

// ScalarProps hold values read as they are. P444 (review scores) is not kept:
// D-09 has no critic scores.
var ScalarProps = []string{
	"P577", "P580", "P571", // dates
	"P2047", "P1104", "P2437", "P1113", // runtime, pages, seasons, episodes
	"P1651", "P18", "P3383", // trailers, images
	"P345", "P5794", "P1733", "P4947", "P4983", "P212", "P648", "P2816",
	"P8383", "P1712", "P1258", "P6278", "P2725",
}

// EntityProps are the external ids kept on a wikidata:entity record, used
// to match credits (SDD 06 §7.4). P4985 and P8084 are [NOT VERIFIED] until
// fixtures pin them.
var EntityProps = []string{"P345", "P648", "P4985", "P8084"}

// KeptProps is every claim property of a trimmed item entity.
var KeptProps = func() map[string]bool {
	m := map[string]bool{}
	for _, p := range slices.Concat(RefProps, ScalarProps) {
		m[p] = true
	}

	return m
}()

// Entity is a wbgetentities entity.
type Entity struct {
	ID       string                 `json:"id"`
	Missing  *string                `json:"missing"`
	Labels   map[string]term        `json:"labels"`
	Aliases  map[string][]term      `json:"aliases"`
	Claims   map[string][]Statement `json:"claims"`
	Sitelink map[string]struct {
		Title string `json:"title"`
	} `json:"sitelinks"`
}

type term struct {
	Value string `json:"value"`
}

// Statement is one claim of an entity.
type Statement struct {
	Mainsnak *struct {
		Snaktype  string `json:"snaktype"`
		Datavalue *struct {
			Type  string          `json:"type"`
			Value json.RawMessage `json:"value"`
		} `json:"datavalue"`
	} `json:"mainsnak"`
	Rank string `json:"rank"`
}

// ParseEntity decodes one entity and rejects a statement without a main
// snak or with an unknown rank.
func ParseEntity(raw []byte) (Entity, error) {
	var e Entity
	if err := json.Unmarshal(raw, &e); err != nil {
		return Entity{}, fmt.Errorf("decode wikidata entity: %w", err)
	}

	if e.ID == "" {
		return Entity{}, errors.New("wikidata entity has no id")
	}

	for prop, statements := range e.Claims {
		for _, st := range statements {
			if st.Mainsnak == nil || st.Mainsnak.Snaktype == "" {
				return Entity{}, fmt.Errorf("%s %s: statement without a main snak", e.ID, prop)
			}

			if st.Rank != "preferred" && st.Rank != "normal" && st.Rank != "deprecated" {
				return Entity{}, fmt.Errorf("%s %s: invalid rank %q", e.ID, prop, st.Rank)
			}
		}
	}

	return e, nil
}

// Claim is a kept statement value.
type Claim struct {
	Value any    `json:"value"`
	Rank  string `json:"rank"`
}

// Payload is the trimmed entity stored in source_records.payload for
// source "wikidata".
type Payload struct {
	Type           string              `json:"type"`
	ID             string              `json:"id"`
	Labels         map[string]string   `json:"labels"`
	Aliases        map[string][]string `json:"aliases"`
	Sitelinks      map[string]string   `json:"sitelinks"`
	SitelinksCount int                 `json:"sitelinks_count"`
	Claims         map[string][]Claim  `json:"claims"`
	Refs           map[string]string   `json:"refs"`
}

// simplify reduces a data value to what the normaliser reads; nil drops it.
func simplify(typ string, raw json.RawMessage) any {
	switch typ {
	case "wikibase-entityid":
		var v struct {
			ID *string `json:"id"`
		}
		if json.Unmarshal(raw, &v) == nil && v.ID != nil {
			return *v.ID
		}
	case "string":
		var s string
		if json.Unmarshal(raw, &s) == nil {
			return s
		}
	case "monolingualtext":
		var v struct {
			Text *string `json:"text"`
		}
		if json.Unmarshal(raw, &v) == nil && v.Text != nil {
			return *v.Text
		}
	case "time":
		var v struct {
			Time      json.RawMessage `json:"time"`
			Precision json.RawMessage `json:"precision"`
		}
		if json.Unmarshal(raw, &v) == nil {
			return map[string]json.RawMessage{"time": nullable(v.Time), "precision": nullable(v.Precision)}
		}
	case "quantity":
		var v struct {
			Amount json.RawMessage `json:"amount"`
			Unit   *string         `json:"unit"`
		}
		if json.Unmarshal(raw, &v) == nil {
			unit := "1"
			if v.Unit != nil {
				unit = *v.Unit
			}

			u, _ := json.Marshal(unit[strings.LastIndex(unit, "/")+1:])

			return map[string]json.RawMessage{"amount": nullable(v.Amount), "unit": u}
		}
	}

	return nil
}

func nullable(r json.RawMessage) json.RawMessage {
	if len(r) == 0 {
		return json.RawMessage("null")
	}

	return r
}

func claims(e Entity, props []string) map[string][]Claim {
	out := map[string][]Claim{}

	for _, prop := range props {
		var kept []Claim

		for _, st := range e.Claims[prop] {
			if st.Rank == "deprecated" || st.Mainsnak.Datavalue == nil {
				continue
			}

			v := simplify(st.Mainsnak.Datavalue.Type, st.Mainsnak.Datavalue.Value)
			if v == nil {
				continue
			}

			kept = append(kept, Claim{Value: v, Rank: st.Rank})
		}

		if len(kept) > 0 {
			out[prop] = kept
		}
	}

	return out
}

func labels(e Entity) map[string]string {
	out := map[string]string{}

	for _, lang := range LabelLangs {
		if t, ok := e.Labels[lang]; ok {
			out[lang] = t.Value
		}
	}

	return out
}

// Trim keeps only the data the normaliser reads, so unrelated edits on
// Wikidata do not change the hash. Refs is filled by the caller.
func Trim(e Entity, itemType string) Payload {
	p := Payload{
		Type:           itemType,
		ID:             e.ID,
		Labels:         labels(e),
		Aliases:        map[string][]string{},
		Sitelinks:      map[string]string{},
		SitelinksCount: len(e.Sitelink),
		Claims:         claims(e, slices.Concat(RefProps, ScalarProps)),
		Refs:           map[string]string{},
	}

	for _, lang := range LabelLangs {
		for _, a := range e.Aliases[lang] {
			p.Aliases[lang] = append(p.Aliases[lang], a.Value)
		}
	}

	for _, site := range SitelinkSites {
		if s, ok := e.Sitelink[site]; ok {
			p.Sitelinks[site] = s.Title
		}
	}

	return p
}

// ReferencedQids returns the sorted entity ids named by the RefProps claims.
func ReferencedQids(p Payload) []string {
	set := map[string]bool{}

	for _, prop := range RefProps {
		for _, c := range p.Claims[prop] {
			if q, ok := c.Value.(string); ok {
				set[q] = true
			}
		}
	}

	out := make([]string, 0, len(set))
	for q := range set {
		out = append(out, q)
	}

	slices.Sort(out)

	return out
}

// EntityPayload is a referenced entity stored as source "wikidata:entity".
type EntityPayload struct {
	ID     string             `json:"id"`
	Labels map[string]string  `json:"labels"`
	Claims map[string][]Claim `json:"claims"`
}

// TrimEntity keeps the labels and external ids of a referenced entity.
func TrimEntity(e Entity) EntityPayload {
	return EntityPayload{ID: e.ID, Labels: labels(e), Claims: claims(e, EntityProps)}
}

// refLabel is the English label, else the "mul" one; empty when neither exists.
func refLabel(e Entity) string {
	if t, ok := e.Labels["en"]; ok && t.Value != "" {
		return t.Value
	}

	return e.Labels["mul"].Value
}
