package models

import (
	"database/sql"
	"time"

	"github.com/AlexGalhardo/vukem/orm"
)

// Entity is a person, organisation, franchise or shared taxonomy term
// (SDD 01 §5.2). WikidataQID is null for HexaScore-native rows.
type Entity struct {
	ID          string
	Kind        string
	Slug        string
	WikidataQID sql.NullString `db:"wikidata_qid"` // VF-036
	Code        sql.NullString
	DefaultName string
	ExternalIDs orm.JSON[map[string]string] `db:"external_ids"` // VF-036
	ImageKey    sql.NullString
	ImageSource sql.NullString
	BornOn      orm.NullTime
	DiedOn      orm.NullTime
	FoundedYear sql.NullInt64
	Provenance  orm.JSON[map[string]string]
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
