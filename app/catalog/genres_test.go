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
		// The 2026-11-02 production dry run's most used unmapped labels.
		{"series", "telenovela", "drama"},
		{"movie", "LGBTQ-related film", "lgbtq"},
		{"series", "LGBT-related television series", "lgbtq"},
		{"game", "science fiction video game", "science-fiction"},
		{"game", "post-apocalyptic video game", "science-fiction"},
		{"game", "eroge", "visual-novel"},
		{"movie", "Christmas film", "family"},
		{"book", "fable", "short-stories"},
		{"movie", "ghost film", "horror"},
		{"game", "cyberpunk video game", "science-fiction"},
		{"game", "maze video game", "puzzle"},
		{"series", "tokusatsu", "action"},
		{"game", "vehicular combat game", "action"},
		{"movie", "speculative fiction film", "science-fiction"},
		{"game", "Arcade", "action"},
		{"game", "video game with LGBT character", "lgbtq"},
		{"game", "fantasy video game", "fantasy"},
		{"movie", "trial film", "crime"},
		{"game", "educational video game", "educational"},
		{"book", "LGBT-related literature", "lgbtq"},
		{"book", "chronicle", "history"},
		{"movie", "sword-and-sandal film", "history"},
		{"movie", "pirate film", "adventure"},
		{"movie", "medieval film", "history"},
		{"game", "mecha video game", "action"},
		{"movie", "time-travel film", "science-fiction"},
		{"movie", "vigilante film", "action"},
		{"game", "action role-playing game", "rpg"},
	} {
		if got, _ := GenreFor(tc.typ, tc.label); got != tc.want {
			t.Errorf("GenreFor(%q, %q) = %q, want %q", tc.typ, tc.label, got, tc.want)
		}
	}

	// Labels that fit no canonical genre stay hidden.
	for _, tc := range []struct{ typ, label string }{
		{"movie", "flashback film"},
		{"movie", "film"},
		{"movie", "epic film"},
		{"game", "video game"},
		{"book", ""},
		// Formats and production labels, not genres.
		{"movie", "silent film"},
		{"movie", "independent film"},
		{"movie", "art film"},
		{"movie", "propaganda film"},
		{"series", "anthology series"},
		{"game", "crossover fiction"},
		{"game", "casual game"},
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
