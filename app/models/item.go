package models

import (
	"database/sql"
	"time"

	"github.com/AlexGalhardo/vukem/orm"
)

// Trailer is one entry of items.trailers.
type Trailer struct {
	YoutubeID string `json:"youtubeId"`
	Title     string `json:"title"`
}

// Item is a catalogue entry (SDD 01 §5.2). IDs are uuid v7 strings set in Go
// before Create (uuid.NewV7); score and search columns are owned by SDD 02
// and SDD 07 and are mapped by their own models.
type Item struct {
	ID               string
	Type             string
	Slug             string
	WikidataQID      string `db:"wikidata_qid"` // Snake would give wikidata_q_i_d (VF-036)
	OriginalTitle    sql.NullString
	ReleaseDate      orm.NullTime
	ReleasePrecision string
	ReleaseStatus    string
	Visibility       string
	RuntimeMinutes   sql.NullInt64
	SeasonCount      sql.NullInt64
	EpisodeCount     sql.NullInt64
	PageCount        sql.NullInt64
	PrimaryEditionID sql.NullString `db:"primary_edition_id"`
	Trailers         orm.JSON[[]Trailer]
	ExternalIDs      orm.JSON[map[string]string] `db:"external_ids"` // VF-036
	Provenance       orm.JSON[map[string]string]
	CoverKey         sql.NullString
	CoverThumbKey    sql.NullString
	CoverSource      sql.NullString
	PopularityPrior  float32
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
