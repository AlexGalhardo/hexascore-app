package catalog

import "testing"

func TestValidYoutubeID(t *testing.T) {
	t.Parallel()
	if !validYoutubeID("aaaaaaaaaaa") {
		t.Fatal("11 url-safe chars")
	}
	for _, id := range []string{"short", "aaaaaaaaaa!", "aaaaaaaaaaá", ""} {
		if validYoutubeID(id) {
			t.Fatalf("accepted %q", id)
		}
	}
}

func TestAssignTrailersDropsInvalidIDs(t *testing.T) {
	t.Parallel()
	got, err := assignTrailers(
		[]byte(`[{"youtubeId":"aaaaaaaaaaa","title":"Demo trailer"},{"youtubeId":"bad","title":"no"}]`),
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].YoutubeID != "aaaaaaaaaaa" || got[0].Title != "Demo trailer" {
		t.Fatalf("got %#v", got)
	}
	empty, err := assignTrailers(nil)
	if err != nil || len(empty) != 0 || empty == nil {
		t.Fatalf("empty %#v %v", empty, err)
	}
}
