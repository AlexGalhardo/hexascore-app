package models

import "time"

// Rating is one user's six-category rating of an item (SDD 02 §5). Values are
// 5–100 in steps of 5 (half stars). The generated `overall` column is not
// mapped: the ORM writes every mapped column and PostgreSQL rejects writes to
// generated columns (P0.F.5.1, SQLSTATE 428C9); read it with RatingRead.
type Rating struct {
	UserID    int64
	ItemID    string
	C1        int16
	C2        int16
	C3        int16
	C4        int16
	C5        int16
	C6        int16
	CreatedAt time.Time
	UpdatedAt time.Time
}

// PrimaryKeys declares the composite key, so orm.Update and orm.Delete filter on
// both columns and refuse a write matching several rows (VF-035).
func (Rating) PrimaryKeys() []string { return []string{"user_id", "item_id"} }

// RatingRead adds the generated overall for reads only.
type RatingRead struct {
	Rating

	Overall float64
}

// Table reads from ratings.
func (RatingRead) Table() string { return "ratings" }
