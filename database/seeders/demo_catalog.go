package seeders

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/AlexGalhardo/vukem/auth/hash"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/factory"
	"github.com/AlexGalhardo/vukem/orm"
	"github.com/AlexGalhardo/vukem/uuid"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/database/factories"
)

// Demo data sizes (SDD 11 §7.7; P1.D.2.1 adds ratings to this seeder).
const (
	demoSeed        = 42
	demoItemsPerTyp = 75
	demoUsers       = 50
	// DemoPassword signs in the demo admin and every demo user.
	DemoPassword = "Demo-password-1"
	// DemoAdminEmail is the demo administrator.
	DemoAdminEmail = "admin@example.test"
)

var (
	demoLocales   = []string{"en", "pt-BR", "es"}
	demoPrefixes  = map[string]string{"en": "", "pt-BR": "O ", "es": "El "}
	demoPlatforms = []string{"PC", "PlayStation", "Xbox", "Nintendo Switch", "Android", "iOS"}
	demoGenres    = []string{
		"Action",
		"Adventure",
		"Comedy",
		"Drama",
		"Fantasy",
		"Horror",
		"Mystery",
		"Romance",
		"Science Fiction",
		"Thriller",
	}
	demoModes = []string{"Single-player", "Multiplayer", "Co-op"}
)

// DemoCatalog fills a Postgres database with a deterministic demo catalogue:
// 300 items (75 per type) with translations in en, pt-BR and es, entities,
// book editions, one admin and 50 users. The same seed gives the same slugs;
// a second run on the same database does nothing.
type DemoCatalog struct{}

// Seed implements vukem.Seeder.
func (DemoCatalog) Seed(ctx context.Context, db *database.DB) error {
	if db.Dialect().Name() != "postgres" {
		return errors.New("DemoCatalog needs PostgreSQL: the catalogue tables are Postgres-only (VF-033)")
	}

	return db.Tx(ctx, func(tx *database.Tx) error { return seedDemoCatalog(ctx, tx) })
}

func seedDemoCatalog(ctx context.Context, tx *database.Tx) error {
	var existing int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM items WHERE provenance @> '{"demo":"true"}'`).
		Scan(&existing); err != nil {
		return err
	}

	if existing > 0 {
		return nil
	}

	if _, err := tx.ExecContext(ctx, `INSERT INTO data_sources (code, name, licence, url)
VALUES ('hexascore', 'HexaScore demo data', 'proprietary', 'https://hexascore.app') ON CONFLICT (code) DO NOTHING`); err != nil {
		return err
	}

	if err := seedDemoUsers(ctx, tx); err != nil {
		return fmt.Errorf("seed users: %w", err)
	}

	entities, err := seedDemoEntities(ctx, tx)
	if err != nil {
		return fmt.Errorf("seed entities: %w", err)
	}

	rng := factory.NewFaker(demoSeed)

	for _, typ := range []string{"game", "movie", "series", "book"} {
		if err := seedDemoItems(ctx, tx, rng, typ, entities); err != nil {
			return fmt.Errorf("seed %s items: %w", typ, err)
		}
	}

	return nil
}

func seedDemoUsers(ctx context.Context, tx *database.Tx) error {
	passwordHash, err := hash.Default().Make(DemoPassword)
	if err != nil {
		return err
	}

	users := factories.User(passwordHash).Seed(demoSeed).Count(demoUsers + 1).Make()

	for i := range users {
		role, username := "user", "demo"+letters(i)

		if i == 0 {
			users[i].Email, users[i].Name, role, username = DemoAdminEmail, "Demo Admin", "admin", "demoadmin"
		}

		if err := orm.Create(ctx, tx, &users[i]); err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, `UPDATE users SET role = $2, username = $3, display_username = $3, locale = $4,
email_verified_at = now(), password_set_at = now() WHERE id = $1`,
			users[i].ID, role, username, demoLocales[i%len(demoLocales)]); err != nil {
			return err
		}
	}

	return nil
}

// letters encodes n as four lowercase letters, so usernames match ^[a-z]{4,16}$.
func letters(n int) string {
	b := []byte("aaaa")
	for i := len(b) - 1; i >= 0; i-- {
		b[i] = byte('a' + n%26)
		n /= 26
	}

	return string(b)
}

// demoEntities holds the ids of the seeded entities by kind.
type demoEntities map[string][]string

func seedDemoEntities(ctx context.Context, tx *database.Tx) (demoEntities, error) {
	out := demoEntities{}

	insert := func(kind, name string) error {
		slug, err := catalog.UniqueSlug(ctx, tx, catalog.ScopeEntity, kind, name, 0)
		if err != nil {
			return err
		}

		id := uuid.NewV7().String()
		if _, err := tx.ExecContext(ctx, `INSERT INTO entities (id, kind, slug, default_name, provenance)
VALUES ($1, $2, $3, $4, '{"demo":"true"}')`, id, kind, slug, name); err != nil {
			return err
		}

		for _, locale := range demoLocales {
			if _, err := tx.ExecContext(ctx, `INSERT INTO entity_translations (entity_id, locale, name, name_source)
VALUES ($1, $2, $3, 'hexascore')`, id, locale, name); err != nil {
				return err
			}
		}

		out[kind] = append(out[kind], id)

		return nil
	}

	fixed := map[string][]string{"platform": demoPlatforms, "genre": demoGenres, "game_mode": demoModes}
	for _, kind := range []string{"platform", "genre", "game_mode"} {
		for _, name := range fixed[kind] {
			if err := insert(kind, name); err != nil {
				return nil, err
			}
		}
	}

	for kind, n := range map[string]int{"person": 40, "organization": 15, "franchise": 10} {
		for _, e := range factories.Entity(kind).Seed(demoSeed).Count(n).Make() {
			if err := insert(kind, e.DefaultName); err != nil {
				return nil, err
			}
		}
	}

	return out, nil
}

// links lists the (role, entity kind, how many) relations each type gets; every
// pair is compatible per catalog.RoleCompatible.
var links = map[string][]struct {
	role, kind string
	n          int
}{
	"game": {
		{"developer", "organization", 1},
		{"publisher", "organization", 1},
		{"platform", "platform", 2},
		{"genre", "genre", 2},
		{"game_mode", "game_mode", 1},
	},
	"movie": {
		{"director", "person", 1},
		{"cast", "person", 3},
		{"genre", "genre", 2},
		{"production_company", "organization", 1},
	},
	"series": {{"creator", "person", 1}, {"cast", "person", 3}, {"genre", "genre", 2}, {"network", "organization", 1}},
	"book":   {{"author", "person", 1}, {"publisher", "organization", 1}, {"genre", "genre", 2}},
}

func seedDemoItems(ctx context.Context, tx *database.Tx, rng *factory.Faker, typ string, entities demoEntities) error {
	for n, item := range factories.Item(typ).Seed(demoSeed).Count(demoItemsPerTyp).Make() {
		title := item.OriginalTitle.String

		slug, err := catalog.UniqueSlug(ctx, tx, catalog.ScopeItem, typ, title, item.ReleaseDate.V.Year())
		if err != nil {
			return err
		}

		item.ID, item.Slug = uuid.NewV7().String(), slug
		if err := orm.Create(ctx, tx, &item); err != nil {
			return err
		}

		for _, locale := range demoLocales {
			if _, err := tx.ExecContext(ctx, `INSERT INTO item_translations
(item_id, locale, title, title_source, synopsis, synopsis_source, synopsis_licence)
VALUES ($1, $2, $3, 'hexascore', $4, 'hexascore', 'proprietary')`,
				item.ID, locale, demoPrefixes[locale]+title, "Demo synopsis for "+title+"."); err != nil {
				return err
			}
		}

		for _, l := range links[typ] {
			ids := entities[l.kind]
			// l.n consecutive ids from a random start: distinct and deterministic.
			start := rng.Int(0, len(ids)-1)

			for position := range l.n {
				i := (start + position) % len(ids)

				if !catalog.RoleCompatible(l.role, l.kind, typ) {
					return fmt.Errorf("incompatible demo link %s/%s/%s", l.role, l.kind, typ)
				}

				if _, err := tx.ExecContext(ctx, `INSERT INTO item_entities (item_id, entity_id, role, position, source)
VALUES ($1, $2, $3, $4, 'hexascore')`, item.ID, ids[i], l.role, position); err != nil {
					return err
				}
			}
		}

		if typ == "book" {
			if err := seedDemoEditions(
				ctx,
				tx,
				item.ID,
				title,
				slug,
				n,
				entities["organization"][rng.Int(0, len(entities["organization"])-1)],
			); err != nil {
				return err
			}
		}
	}

	return nil
}

// seedDemoEditions adds one or two editions and makes the first one primary;
// every third book also gets a Brazilian edition.
func seedDemoEditions(
	ctx context.Context,
	tx *database.Tx,
	bookID, title, slug string,
	n int,
	publisherID string,
) error {
	editions := []string{"en"}
	if n%3 == 0 {
		editions = append(editions, "pt-BR")
	}

	for i, language := range editions {
		id := uuid.NewV7().String()
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO editions
(id, work_id, openlibrary_id, title, language, market_country, publisher_id, publication_year, format, source)
VALUES ($1, $2, $3, $4, $5, $6, $7, 2000 + $8::int, 'paperback', 'hexascore')`,
			id,
			bookID,
			"OLDEMO-"+strings.ToUpper(slug)+"-"+language,
			title,
			language,
			map[string]string{"en": "US", "pt-BR": "BR"}[language],
			publisherID,
			n%25,
		); err != nil {
			return err
		}

		if i == 0 {
			if _, err := tx.ExecContext(
				ctx,
				`UPDATE items SET primary_edition_id = $1 WHERE id = $2`,
				id,
				bookID,
			); err != nil {
				return err
			}
		}
	}

	return nil
}
