package models

import (
	"database/sql"
	"time"

	"github.com/AlexGalhardo/vukem/orm"
)

// ScoreBurst is one detected rating burst on an item (SDD 02 §5). The admin
// panel reads it; scoring writes it in SQL.
type ScoreBurst struct {
	ID           int64
	ItemID       string
	StartedAt    time.Time
	LimitedUntil time.Time
	Trigger      string
	Stats        orm.JSON[map[string]any]
	Resolution   sql.NullString
	ResolvedAt   orm.NullTime
	ResolvedBy   sql.NullInt64
}

// RatingFlag is a shadow weight on one (user, item) rating (SDD 02 §5).
// The real key is (user_id, item_id), but vadmin.NewResource panics on composite
// models and the admin panel only lists it, so it keeps a single-column
// PrimaryKey: never call orm.Update or orm.Delete on it (VF-035).
type RatingFlag struct {
	UserID       int64
	ItemID       string
	ShadowWeight float64
	Reason       string
	BurstID      sql.NullInt64
	CreatedAt    time.Time
}

// PrimaryKey names user_id only to satisfy the ORM's single-column key.
func (RatingFlag) PrimaryKey() string { return "user_id" }

// ScoreAction is one row of the scoring moderation audit trail (SDD 02 §5).
type ScoreAction struct {
	ID      int64
	At      time.Time
	Actor   sql.NullInt64
	Kind    string
	ItemID  sql.NullString
	UserID  sql.NullInt64
	BurstID sql.NullInt64
	Before  orm.JSON[map[string]any]
	After   orm.JSON[map[string]any]
	Note    sql.NullString
}

// ScoreFormulaVersion is one activated scoring formula (SDD 02 §5).
type ScoreFormulaVersion struct {
	Version     int16
	MGame       int16 `db:"m_game"`
	MMovie      int16 `db:"m_movie"`
	MSeries     int16 `db:"m_series"`
	MBook       int16 `db:"m_book"`
	Params      orm.JSON[map[string]any]
	ActivatedAt time.Time
	Changelog   string
}

// PrimaryKey is version, not id.
func (ScoreFormulaVersion) PrimaryKey() string { return "version" }
