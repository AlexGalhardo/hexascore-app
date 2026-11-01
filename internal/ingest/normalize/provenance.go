package normalize

// data_sources codes (SDD 01 §5.2). Every provenance value and every
// *_source column must exist as a data_sources row (DQ-11).
const (
	SourceWikidata    = "wikidata"
	SourceWikipedia   = "wikipedia"
	SourceTMDB        = "tmdb"
	SourceIGDB        = "igdb"
	SourceSteam       = "steam"
	SourceGOG         = "gog"
	SourceOpenLibrary = "openlibrary"
	SourceOwn         = "hexascore"
	SourceCommons     = "commons"
)

// DataSource is one data_sources row.
type DataSource struct {
	Code, Name, Licence, Attribution, URL string
}

// DataSources lists the rows the merge labels fields with. Licences name the
// terms each source publishes (owner cleared the legal reads in Q-A16,
// 2026-11-01); they only decide what is displayed. Terms pages:
//   - Open Library: https://openlibrary.org/developers/licensing (no rights asserted by the Internet Archive)
var DataSources = []DataSource{
	{SourceWikidata, "Wikidata", "CC0-1.0", "", "https://www.wikidata.org"},
	{SourceWikipedia, "Wikipedia", "CC-BY-SA-4.0", "Wikipedia (CC BY-SA 4.0)", "https://www.wikipedia.org"},
	{SourceTMDB, "TMDB", "TMDB-API-TERMS", "TMDB", "https://www.themoviedb.org"},
	{SourceIGDB, "IGDB", "IGDB-API-TERMS", "IGDB", "https://www.igdb.com"},
	{SourceSteam, "Steam", "STEAM-API-TERMS", "", "https://store.steampowered.com"},
	{SourceGOG, "GOG", "GOG-API-TERMS", "", "https://www.gog.com"},
	{SourceOpenLibrary, "Open Library", "OPEN-LIBRARY-NO-RIGHTS-ASSERTED", "", "https://openlibrary.org"},
	{SourceOwn, "HexaScore", "proprietary", "", "https://hexascore.app"},
	{SourceCommons, "Wikimedia Commons", "PER-FILE", "", "https://commons.wikimedia.org"},
}

// licencePlaceholders are the values seeded before the terms were read; SeedDataSources replaces them.
var licencePlaceholders = []string{"IGDB-API-TERMS", "STEAM-API-TERMS", "GOG-API-TERMS", "OPEN-LIBRARY-PENDING-A05"}

// Licence returns the data_sources licence of code, or "" for an unknown source.
func Licence(code string) string {
	for _, s := range DataSources {
		if s.Code == code {
			return s.Licence
		}
	}

	return ""
}

// Provenance maps a field name (SDD 04 §7.2 names where one exists, else the
// items column or "meta.<key>") to the data_sources code that supplied it.
type Provenance map[string]string
