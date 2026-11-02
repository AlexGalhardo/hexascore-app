package catalog

import (
	"regexp"
	"strings"
)

// GenreName is the display name of a canonical genre in the three locales.
type GenreName struct{ En, PtBR, Es string }

// Genres are the canonical genres (owner 2026-11-02): one short, translated name per code, shared by the
// categories that use it. Items link to these entities; the Wikidata genres they came from are kept in
// item_genre_sources and never shown. The code is also the slug, so ?genre=comedy filters every category.
var Genres = map[string]GenreName{
	"action":             {"Action", "Ação", "Acción"},
	"adventure":          {"Adventure", "Aventura", "Aventura"},
	"animation":          {"Animation", "Animação", "Animación"},
	"biography":          {"Biography", "Biografia", "Biografía"},
	"card-board":         {"Card and board", "Cartas e tabuleiro", "Cartas y tablero"},
	"children":           {"Children's", "Infantil", "Infantil"},
	"comedy":             {"Comedy", "Comédia", "Comedia"},
	"comics":             {"Comics", "Quadrinhos", "Cómics"},
	"crime":              {"Crime", "Crime", "Crimen"},
	"documentary":        {"Documentary", "Documentário", "Documental"},
	"drama":              {"Drama", "Drama", "Drama"},
	"family":             {"Family", "Família", "Familiar"},
	"fantasy":            {"Fantasy", "Fantasia", "Fantasía"},
	"fiction":            {"Fiction", "Ficção", "Ficción"},
	"fighting":           {"Fighting", "Luta", "Lucha"},
	"historical-fiction": {"Historical fiction", "Ficção histórica", "Ficción histórica"},
	"history":            {"History", "História", "Historia"},
	"horror":             {"Horror", "Terror", "Terror"},
	"humor":              {"Humor", "Humor", "Humor"},
	"mmo":                {"MMO", "MMO", "MMO"},
	"music":              {"Music and rhythm", "Música e ritmo", "Música y ritmo"},
	"musical":            {"Musical", "Musical", "Musical"},
	"mystery":            {"Mystery", "Mistério", "Misterio"},
	"non-fiction":        {"Non-fiction", "Não ficção", "No ficción"},
	"party":              {"Party", "Party", "Party"},
	"philosophy":         {"Philosophy", "Filosofia", "Filosofía"},
	"platform":           {"Platform", "Plataforma", "Plataformas"},
	"poetry":             {"Poetry", "Poesia", "Poesía"},
	"puzzle":             {"Puzzle", "Quebra-cabeça", "Puzles"},
	"racing":             {"Racing", "Corrida", "Carreras"},
	"reality":            {"Reality", "Reality show", "Telerrealidad"},
	"roguelike":          {"Roguelike", "Roguelike", "Roguelike"},
	"romance":            {"Romance", "Romance", "Romance"},
	"rpg":                {"RPG", "RPG", "Rol"},
	"sandbox":            {"Sandbox", "Sandbox", "Sandbox"},
	"science-fiction":    {"Science fiction", "Ficção científica", "Ciencia ficción"},
	"self-help":          {"Self-help", "Autoajuda", "Autoayuda"},
	"shooter":            {"Shooter", "Tiro", "Disparos"},
	"short-stories":      {"Short stories", "Contos", "Cuentos"},
	"simulation":         {"Simulation", "Simulação", "Simulación"},
	"sports":             {"Sports", "Esportes", "Deportes"},
	"stealth":            {"Stealth", "Furtividade", "Sigilo"},
	"strategy":           {"Strategy", "Estratégia", "Estrategia"},
	"theatre":            {"Theatre", "Teatro", "Teatro"},
	"thriller":           {"Thriller", "Suspense", "Suspense"},
	"visual-novel":       {"Visual novel", "Visual novel", "Novela visual"},
	"war":                {"War", "Guerra", "Bélico"},
	"western":            {"Western", "Faroeste", "Western"},
	"young-adult":        {"Young adult", "Jovem adulto", "Juvenil"},
}

// genreRule maps labels holding any of its keywords (whole words, lower case) to a canonical code.
type genreRule struct {
	code string
	re   *regexp.Regexp
}

func rules(pairs ...string) []genreRule {
	out := make([]genreRule, 0, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		kws := strings.Split(pairs[i+1], "|")
		for j, k := range kws {
			kws[j] = regexp.QuoteMeta(k)
		}
		out = append(out, genreRule{pairs[i], regexp.MustCompile(`(^|[^a-z])(` + strings.Join(kws, "|") + `)([^a-z]|$)`)})
	}

	return out
}

// screenRules serve movies and series. First match wins, so the more specific rules come first:
// "romantic comedy" is comedy, "crime thriller" is thriller, "American football film" is sports.
var screenRules = rules(
	"documentary", "documentary|docudrama|mockumentary",
	"animation", "animated|animation|anime|cartoon|stop-motion",
	"reality", "reality|talk show|game show|variety show",
	"science-fiction", "science fiction|sci-fi|cyberpunk|space opera|dystopian|post-apocalyptic|time travel",
	"horror", "horror|slasher|zombie|vampire|splatter|monster",
	"comedy", "comedy|comedic|sitcom|parody|satire|spoof|tragicomedy|buddy",
	"thriller", "thriller|suspense|psychological|heist|spy|espionage|conspiracy|erotic thriller",
	"crime", "crime|gangster|police|detective|noir|neo-noir|legal|courtroom|procedural|mafia|prison",
	"mystery", "mystery|whodunit",
	"war", "war|military|anti-war",
	"western", "western",
	"musical", "musical|music|dance",
	"sports", "sport|sports|football|soccer|baseball|basketball|boxing|hockey|racing|wrestling|surfing",
	"biography", "biographical|biopic|biography",
	"history", "historical|history|period|costume drama",
	"family", "family|children's|kids",
	"fantasy", "fantasy|fairy tale|sword and sorcery|supernatural|magic",
	"action", "action|martial arts|superhero|disaster|kung fu|wuxia",
	"adventure", "adventure|swashbuckler|survival|road",
	"romance", "romance|romantic|love",
	"drama", "drama|melodrama|soap opera|tragedy|coming-of-age|teen",
)

var gameRules = rules(
	"mmo", "massively multiplayer|mmo|mmorpg",
	"roguelike", "roguelike|roguelite",
	"visual-novel", "visual novel|dating sim",
	"rpg", "role-playing|rpg|jrpg|dungeon crawler",
	"shooter", "shooter|first-person shooter|third-person shooter|shoot 'em up|shmup|battle royale|light gun",
	"fighting", "fighting|beat 'em up|brawler",
	"platform", "platform|platformer|metroidvania",
	"racing", "racing|driving|kart",
	"sports", "sports|sport|football|soccer|basketball|golf|tennis|baseball|hockey|skateboarding",
	"horror", "horror|survival horror",
	"stealth", "stealth",
	"strategy", "strategy|real-time strategy|4x|tower defense|tactics|tactical|moba|wargame|grand strategy",
	"simulation", "simulation|simulator|life simulation|city-building|management|construction|flight",
	"puzzle", "puzzle|tile-matching|logic",
	"music", "rhythm|music",
	"party", "party|trivia|quiz",
	"card-board", "card|board|deck-building|digital collectible card|tabletop",
	"sandbox", "sandbox|open world|survival|crafting",
	"adventure", "adventure|action-adventure|point-and-click|interactive fiction|graphic adventure|walking simulator",
	"action", "action|hack and slash|hack-and-slash|run and gun",
)

var bookRules = rules(
	"humor", "comic novel|humor|humour|satire|satirical|comedy",
	"comics", "graphic novel|comic|comics|manga|comic book",
	"children", "children's|picture book|children",
	"young-adult", "young adult|ya",
	"science-fiction", "science fiction|sci-fi|cyberpunk|space opera|dystopian|post-apocalyptic|speculative",
	"fantasy", "fantasy|fairy tale|sword and sorcery|mythology|myth",
	"horror", "horror|gothic|ghost story|vampire",
	"thriller", "thriller|suspense|spy|espionage",
	"mystery", "mystery|detective|whodunit",
	"crime", "crime|noir|legal",
	"romance", "romance|romantic|love story",
	"historical-fiction", "historical novel|historical fiction|historical",
	"biography", "biography|autobiography|memoir|biographical",
	"poetry", "poetry|poem|poems|verse|sonnet|epic poem",
	"short-stories", "short story|short stories|anthology|novella collection",
	"theatre", "play|drama|tragedy|theatre|theater",
	"self-help", "self-help",
	"philosophy", "philosophy|philosophical",
	"history", "history",
	"adventure", "adventure",
	"non-fiction", "non-fiction|nonfiction|essay|textbook|reference|popular science|travel literature|cookbook",
	"fiction", "novel|fiction|literary|novella|bildungsroman|prose",
)

// GenreFor maps a source genre label (Wikidata, IGDB, TMDB) of an item of typ to its canonical code; false means
// the label fits no canonical genre and stays hidden.
func GenreFor(typ, label string) (string, bool) {
	var rs []genreRule

	switch typ {
	case "movie", "series":
		rs = screenRules
	case "game":
		rs = gameRules
	case "book":
		rs = bookRules
	}

	l := strings.ToLower(strings.TrimSpace(label))
	for _, r := range rs {
		if r.re.MatchString(l) {
			return r.code, true
		}
	}

	return "", false
}
