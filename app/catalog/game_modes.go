package catalog

import "strings"

// GameModes are the canonical game modes (owner 2026-11-04): short names with a capital first letter in the three
// locales, replacing the Wikidata labels ("single-player video game", "multiplayer online game"). The code is also
// the slug, so ?mode=co-op filters /games. The source labels are kept in item_genre_sources and never shown.
var GameModes = map[string]GenreName{
	"co-op": {
		"Co-op",
		"Cooperativo", //nolint:misspell // Portuguese, not "Cooperation"
		"Cooperativo", //nolint:misspell // Spanish, not "Cooperation"
	},
	"mmo":           {"MMO", "MMO", "MMO"},
	"multiplayer":   {"Multiplayer", "Multijogador", "Multijugador"},
	"online":        {"Online multiplayer", "Multijogador online", "Multijugador en línea"},
	"pvp":           {"PvP", "Jogador contra jogador", "Jugador contra jugador"},
	"single-player": {"Single player", "Um jogador", "Un jugador"},
	"split-screen":  {"Split screen", "Tela dividida", "Pantalla dividida"},
}

// gameModeRules serve the games. First match wins, so the specific rules come first: "shared/splitscreen
// multiplayer" is split screen, "multiplayer online game" is online multiplayer and only a plain "multiplayer video
// game" is multiplayer.
var gameModeRules = rules(
	"mmo", "massively multiplayer|mmo|mmorpg",
	"split-screen", "split-screen|splitscreen|split screen|shared/splitscreen|local multiplayer|hotseat|hot seat|couch",
	"pvp", "player versus player|player vs player|pvp|versus",
	"co-op", "co-op|coop|cooperative|cooperation",
	"online", "online|cross-platform",
	"single-player", "single-player|single player|singleplayer|solo",
	"multiplayer", "multiplayer|multi-player|multi player",
)

// GameModeFor maps a source game mode label (Wikidata P404, IGDB) to its canonical code; false means the label fits
// no canonical mode and stays hidden.
func GameModeFor(typ, label string) (string, bool) {
	if typ != "game" {
		return "", false
	}

	l := strings.ToLower(strings.TrimSpace(label))
	for _, r := range gameModeRules {
		if r.re.MatchString(l) {
			return r.code, true
		}
	}

	return "", false
}
