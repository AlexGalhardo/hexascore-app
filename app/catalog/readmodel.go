package catalog

import (
	"slices"
	"time"
)

// Sourced is a value with the data source it came from.
type Sourced[T any] struct {
	Value  T
	Source string
}

// SynopsisIn is a synopsis with its licence and source URL.
type SynopsisIn struct {
	Text, Source, Licence, SourceURL string
}

// EntityRefIn is one item-entity relation as stored.
type EntityRefIn struct {
	ID, Kind, Role, Name, Source string
}

// ItemSource is everything the loader read for one item, before filtering.
type ItemSource struct {
	ID, Type, Slug, WikidataQID string
	Title                       Sourced[string]
	Synopsis                    *SynopsisIn
	ReleaseDate                 *Sourced[time.Time]
	ReleasePrecision            string
	RuntimeMinutes              *Sourced[int]
	PageCount                   *Sourced[int]
	Cover                       *Sourced[string] // public cover URL
	ExternalIDs                 map[string]string
	Entities                    []EntityRefIn
	AffiliateLinks              []string
}

// Synopsis is the exposed synopsis; licence and source travel with the text.
type Synopsis struct {
	Text      string `json:"text"`
	Licence   string `json:"licence"`
	SourceURL string `json:"sourceUrl,omitempty"`
}

// EntityRef is an exposed item-entity relation.
type EntityRef struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
	Role string `json:"role"`
	Name string `json:"name"`
}

// Meta lists what the audience did not get, so buyers see partial coverage.
type Meta struct {
	OmittedFields []string `json:"omittedFields,omitempty"`
}

// ItemDetail is the item read model. Excluded fields are omitted from JSON,
// never null (SDD 01 §6.1).
type ItemDetail struct {
	ID               string            `json:"id"`
	Type             string            `json:"type"`
	Slug             string            `json:"slug"`
	WikidataQID      string            `json:"wikidataQid"`
	Title            string            `json:"title,omitempty"`
	Synopsis         *Synopsis         `json:"synopsis,omitempty"`
	ReleaseDate      string            `json:"releaseDate,omitempty"`
	ReleasePrecision string            `json:"releasePrecision,omitempty"`
	RuntimeMinutes   *int              `json:"runtimeMinutes,omitempty"`
	PageCount        *int              `json:"pageCount,omitempty"`
	Cover            string            `json:"cover,omitempty"`
	ExternalIDs      map[string]string `json:"externalIds,omitempty"`
	Entities         []EntityRef       `json:"entities,omitempty"`
	AffiliateLinks   []string          `json:"affiliateLinks,omitempty"`
	Meta             *Meta             `json:"meta,omitempty"`
}

// identifierKeys are external ids always exposed alongside allowed sources.
var identifierKeys = []string{"wikidata", "isbn13"}

// Assemble applies the audience policy to src.
func Assemble(src ItemSource, p Policy) ItemDetail {
	d := ItemDetail{ID: src.ID, Type: src.Type, Slug: src.Slug, WikidataQID: src.WikidataQID}
	var omitted []string
	omit := func(field string) { omitted = append(omitted, field) }

	if p.Allows(src.Title.Source) {
		d.Title = src.Title.Value
	} else {
		omit("title")
	}
	if s := src.Synopsis; s != nil {
		if p.AllowsContent(s.Source) {
			d.Synopsis = &Synopsis{Text: s.Text, Licence: s.Licence, SourceURL: s.SourceURL}
		} else {
			omit("synopsis")
		}
	}
	if r := src.ReleaseDate; r != nil {
		if p.AllowsContent(r.Source) {
			d.ReleaseDate, d.ReleasePrecision = r.Value.Format(time.DateOnly), src.ReleasePrecision
		} else {
			omit("releaseDate")
		}
	}
	if r := src.RuntimeMinutes; r != nil {
		if p.AllowsContent(r.Source) {
			d.RuntimeMinutes = &r.Value
		} else {
			omit("runtimeMinutes")
		}
	}
	if r := src.PageCount; r != nil {
		if p.AllowsContent(r.Source) {
			d.PageCount = &r.Value
		} else {
			omit("pageCount")
		}
	}
	if c := src.Cover; c != nil {
		if p.AllowsContent(c.Source) {
			d.Cover = c.Value
		} else {
			omit("cover")
		}
	}

	for key, value := range src.ExternalIDs {
		if p.audience == AudienceSite || slices.Contains(identifierKeys, key) || p.Allows(key) {
			if d.ExternalIDs == nil {
				d.ExternalIDs = map[string]string{}
			}
			d.ExternalIDs[key] = value
		} else if !slices.Contains(omitted, "externalIds") {
			omit("externalIds")
		}
	}

	for _, e := range src.Entities {
		if p.Allows(e.Source) {
			d.Entities = append(d.Entities, EntityRef{ID: e.ID, Kind: e.Kind, Role: e.Role, Name: e.Name})
		} else if !slices.Contains(omitted, "entities") {
			omit("entities")
		}
	}

	if p.audience == AudienceSite {
		d.AffiliateLinks = src.AffiliateLinks
	}

	if len(omitted) > 0 {
		slices.Sort(omitted)
		d.Meta = &Meta{OmittedFields: omitted}
	}
	return d
}
