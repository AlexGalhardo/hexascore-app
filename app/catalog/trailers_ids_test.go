package catalog

import "testing"

// The normaliser writes trailers as a list of YouTube ids; the page used to read only objects and returned a 500 for
// any item with a trailer from Wikidata.
func TestAssignTrailersReadsIdsAndObjects(t *testing.T) {
	for name, c := range map[string]struct {
		raw  string
		want []string
	}{
		"ids from the normaliser": {`["QdBZY2fkU-0","VQRLujxTm3c"]`, []string{"QdBZY2fkU-0", "VQRLujxTm3c"}},
		"objects":                 {`[{"youtubeId":"QdBZY2fkU-0","title":"Trailer 1"}]`, []string{"QdBZY2fkU-0"}},
		"mixed":                   {`["QdBZY2fkU-0",{"youtubeId":"VQRLujxTm3c"}]`, []string{"QdBZY2fkU-0", "VQRLujxTm3c"}},
		"invalid id dropped":      {`["not a youtube id", "QdBZY2fkU-0"]`, []string{"QdBZY2fkU-0"}},
		"odd entries skipped":     {`[42, null, "QdBZY2fkU-0"]`, []string{"QdBZY2fkU-0"}},
		"empty":                   {`[]`, nil},
		"null":                    {`null`, nil},
	} {
		got, err := assignTrailers([]byte(c.raw))
		if err != nil {
			t.Errorf("%s: %v", name, err)

			continue
		}

		if len(got) != len(c.want) {
			t.Errorf("%s: %d trailers, want %d: %+v", name, len(got), len(c.want), got)

			continue
		}

		for i, w := range c.want {
			if got[i].YoutubeID != w {
				t.Errorf("%s: trailer %d = %q, want %q", name, i, got[i].YoutubeID, w)
			}
		}
	}
}
