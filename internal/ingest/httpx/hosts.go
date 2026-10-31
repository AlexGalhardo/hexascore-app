package httpx

// DefaultHosts is the SDD 06 §7.2 table for the connectors built so far;
// later connectors add their hosts.
func DefaultHosts() map[string]HostLimit {
	return map[string]HostLimit{
		"query.wikidata.org":     {PerSecond: 1, InFlight: 1},
		"www.wikidata.org":       {PerSecond: 5, InFlight: 2},
		"en.wikipedia.org":       {PerSecond: 5, InFlight: 4},
		"pt.wikipedia.org":       {PerSecond: 5, InFlight: 4},
		"es.wikipedia.org":       {PerSecond: 5, InFlight: 4},
		"commons.wikimedia.org":  {PerSecond: 5, InFlight: 2},
		"upload.wikimedia.org":   {PerSecond: 2, InFlight: 2},
		"openlibrary.org":        {PerSecond: 2.5, InFlight: 2},
		"covers.openlibrary.org": {PerSecond: 10, InFlight: 4},
	}
}
