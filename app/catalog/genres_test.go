package catalog

import (
	"regexp"
	"testing"
)

// The entities.slug CHECK: canonical codes are also slugs.
var slugRe = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func TestGenreFor(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct{ typ, label, want string }{
		// The owner's examples (2026-11-02).
		{"movie", "comedy film", "comedy"},
		{"movie", "buddy film", "comedy"},
		{"movie", "children's film", "family"},
		{"movie", "adventure film", "adventure"},
		{"movie", "survival film", "adventure"},
		{"movie", "coming-of-age film", "drama"},
		{"movie", "drama film", "drama"},
		{"movie", "historical drama", "history"},
		{"movie", "romantic comedy film", "comedy"},
		{"movie", "tragicomedy", "comedy"},
		{"movie", "American football film", "sports"},
		{"movie", "gangster film", "crime"},
		{"movie", "crime thriller film", "thriller"},
		{"movie", "police procedural", "crime"},
		{"movie", "science fiction film", "science-fiction"},
		{"movie", "contemporary Western film", "western"},
		{"series", "crime television series", "crime"},
		{"series", "drama television series", "drama"},
		{"series", "thriller television series", "thriller"},
		{"series", "black comedy television series", "comedy"},
		{"series", "reality television", "reality"},
		{"game", "role-playing video game", "rpg"},
		{"game", "massively multiplayer online role-playing game", "mmo"},
		{"game", "first-person shooter", "shooter"},
		{"game", "action-adventure game", "adventure"},
		{"game", "Hack and slash/Beat 'em up", "fighting"},
		{"game", "survival horror", "horror"},
		{"game", "Real Time Strategy (RTS)", "strategy"},
		{"game", "platform game", "platform"},
		{"book", "novel", "fiction"},
		{"book", "fantasy novel", "fantasy"},
		{"book", "mystery fiction", "mystery"},
		{"book", "historical novel", "historical-fiction"},
		{"book", "graphic novel", "comics"},
		{"book", "comic novel", "humor"},
		{"book", "children's literature", "children"},
		{"book", "autobiography", "biography"},
		{"book", "Science Fiction", "science-fiction"},
	} {
		if got, _ := GenreFor(tc.typ, tc.label); got != tc.want {
			t.Errorf("GenreFor(%q, %q) = %q, want %q", tc.typ, tc.label, got, tc.want)
		}
	}

	// Labels that fit no canonical genre stay hidden.
	for _, tc := range []struct{ typ, label string }{
		{"movie", "flashback film"}, {"movie", "film"}, {"movie", "epic film"}, {"game", "video game"}, {"book", ""},
	} {
		if code, ok := GenreFor(tc.typ, tc.label); ok {
			t.Errorf("GenreFor(%q, %q) = %q, want no match", tc.typ, tc.label, code)
		}
	}
}

func TestEveryGenreRuleHasTranslatedNames(t *testing.T) {
	t.Parallel()

	for _, rs := range [][]genreRule{screenRules, gameRules, bookRules} {
		for _, r := range rs {
			n, ok := Genres[r.code]
			if !ok || n.En == "" || n.PtBR == "" || n.Es == "" {
				t.Errorf("rule code %q has no complete name in Genres", r.code)
			}
			if !slugRe.MatchString(r.code) {
				t.Errorf("code %q is not a valid slug", r.code)
			}
		}
	}
}
