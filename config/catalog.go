// Package config holds typed application settings read from the environment.
package config

import "strings"

// DefaultCatalogAPISources are the sources whose fields reach paid API
// customers (SDD 01 §6.1, ADR 0006): open licences plus HexaScore's own data,
// which includes approved community edits (A-07, SDD 05 §5).
var DefaultCatalogAPISources = []string{"wikidata", "wikipedia", "openlibrary", "hexascore", "community"}

// Catalog configures the catalogue read models.
type Catalog struct {
	// APISources are the data_sources codes allowed for the "api" audience
	// (CATALOG_API_SOURCES, comma-separated).
	APISources []string
}

// LoadCatalog reads CATALOG_API_SOURCES through get (app.Source.Get).
func LoadCatalog(get func(string) string) Catalog {
	raw := strings.TrimSpace(get("CATALOG_API_SOURCES"))
	if raw == "" {
		return Catalog{APISources: append([]string(nil), DefaultCatalogAPISources...)}
	}
	var sources []string
	for s := range strings.SplitSeq(raw, ",") {
		if s = strings.TrimSpace(strings.ToLower(s)); s != "" {
			sources = append(sources, s)
		}
	}
	return Catalog{APISources: sources}
}
