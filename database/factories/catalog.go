package factories

import (
	"database/sql"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/AlexGalhardo/vukem/factory"
	"github.com/AlexGalhardo/vukem/orm"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// DemoMarker is stored in provenance so the demo seeder can tell its own rows
// apart and skip itself on a second run.
const DemoMarker = "demo"

// qidBase keeps the fake Wikidata ids of each type in their own range.
var qidBase = map[string]int{"game": 1_000_000, "movie": 2_000_000, "series": 3_000_000, "book": 4_000_000}

func title(f *factory.Faker, words int) string {
	s := f.Words(words)
	r, n := utf8.DecodeRuneInString(s)

	return string(unicode.ToUpper(r)) + s[n:]
}

func demoProvenance() orm.JSON[map[string]string] {
	return orm.JSON[map[string]string]{V: map[string]string{DemoMarker: "true"}}
}

// Item builds unsaved items of one type with a title in OriginalTitle. The
// slug is left empty: the seeder assigns it with catalog.UniqueSlug.
func Item(itemType string) factory.Factory[models.Item] {
	return factory.New(func(f *factory.Faker) models.Item {
		released := f.Time(time.Date(1985, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC))
		item := models.Item{
			Type:          itemType,
			WikidataQID:   "Q" + strconv.Itoa(qidBase[itemType]+f.Seq()),
			OriginalTitle: sql.NullString{String: title(f, f.Int(1, 3)), Valid: true},
			ReleaseDate: orm.NullTime{
				Null: sql.Null[time.Time]{V: released.Truncate(24 * time.Hour), Valid: true},
			},
			ReleasePrecision: "day",
			ReleaseStatus:    "released",
			Visibility:       "public",
			Trailers:         orm.JSON[[]models.Trailer]{V: []models.Trailer{}},
			ExternalIDs:      orm.JSON[map[string]string]{V: map[string]string{}},
			Provenance:       demoProvenance(),
			PopularityPrior:  float32(f.Float(0, 100)),
		}

		switch itemType {
		case "movie":
			item.RuntimeMinutes = sql.NullInt64{Int64: int64(f.Int(80, 180)), Valid: true}
		case "series":
			seasons := f.Int(1, 8)
			item.SeasonCount = sql.NullInt64{Int64: int64(seasons), Valid: true}
			item.EpisodeCount = sql.NullInt64{Int64: int64(seasons * f.Int(6, 24)), Valid: true}
		case "book":
			item.PageCount = sql.NullInt64{Int64: int64(f.Int(120, 900)), Valid: true}
		}

		return item
	})
}

// Entity builds unsaved person, organization or franchise entities. The slug
// is left empty: the seeder assigns it.
func Entity(kind string) factory.Factory[models.Entity] {
	return factory.New(func(f *factory.Faker) models.Entity {
		var name string

		switch kind {
		case "person":
			name = f.Name()
		case "organization":
			name = f.Company()
		default:
			name = title(f, f.Int(1, 2))
		}

		return models.Entity{
			Kind:        kind,
			DefaultName: strings.TrimSpace(name),
			ExternalIDs: orm.JSON[map[string]string]{V: map[string]string{}},
			Provenance:  demoProvenance(),
		}
	})
}

// User builds unsaved users sharing one password hash (hashing 51 passwords
// is slow and pointless for demo accounts). Role, username and locale are set
// by the seeder because the model does not map those columns.
func User(passwordHash string) factory.Factory[models.User] {
	return factory.New(func(f *factory.Faker) models.User {
		name := f.Name()

		return models.User{
			Name:     name,
			Email:    "demo-user-" + strconv.Itoa(f.Seq()) + "@example.test",
			Password: passwordHash,
		}
	})
}
