package scoring

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"net/http"
	"slices"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/events"
	"github.com/AlexGalhardo/vukem/uuid"

	"github.com/AlexGalhardo/hexascore-app/app/api/changes"
)

// CategoryKeys are the stable category keys per item type, in column order
// c1..c6 (A-06); they double as API filter names.
var CategoryKeys = map[string][Categories]string{
	"game":   {"gameplay", "fun", "sound", "technical", "visual", "storytelling"},
	"movie":  {"screenplay", "acting", "direction", "cinematography", "soundtrack", "entertainment"},
	"series": {"screenplay", "acting", "characters", "pacing", "production", "entertainment"},
	"book":   {"plot", "characters", "writing", "pacing", "originality", "impact"},
}

// defaultPrior is C_{t,k} while score_priors has no row for the type (SDD 02 §7.1).
const defaultPrior = 60

// Problem codes (SDD 02 §6).
const (
	CodeNotFound    = "not-found"
	CodeNotReleased = "rating-not-released"
)

// Events dispatched after commit (SDD 02 §6).
type (
	// RatingCreated follows a first rating of an item by a user.
	RatingCreated struct {
		UserID int64
		ItemID string
		At     time.Time
	}
	// RatingUpdated follows an edit.
	RatingUpdated struct {
		UserID int64
		ItemID string
		At     time.Time
	}
	// RatingDeleted follows a removal.
	RatingDeleted struct {
		UserID int64
		ItemID string
		At     time.Time
	}
	// ItemScoreChanged reports the stored score (0–100 scale) before and
	// after a write; nil means no score.
	ItemScoreChanged struct {
		ItemID         string
		Before, After  *float64
		FormulaVersion int
	}
)

// Summary is the public aggregate of an item (SDD 02 §6); score and
// categories are null below 5 ratings.
type Summary struct {
	Score           *float64           `json:"score"`
	ScoreByCategory map[string]float64 `json:"scoreByCategory"`
	RatingCount     int                `json:"ratingCount"`
	Confidence      *string            `json:"confidence"`
	FormulaVersion  int                `json:"formulaVersion"`
}

// Writer is the rating write path (SDD 02 §7.2).
type Writer struct {
	DB     *database.DB
	Events *events.Bus
	Now    func() time.Time
}

type lockedItem struct {
	typ       string
	release   sql.NullTime
	precision string
	status    string
	before    sql.NullFloat64
	limited   sql.NullTime
}

// Save upserts userID's rating of itemID (stars per category key, already
// range-checked) and recomputes the item in the same transaction.
func (w *Writer) Save(ctx context.Context, userID int64, itemID string, stars map[string]float64) (Summary, error) {
	return w.write(ctx, userID, itemID, stars)
}

// Delete removes userID's rating of itemID; a missing rating is a no-op.
func (w *Writer) Delete(ctx context.Context, userID int64, itemID string) (Summary, error) {
	return w.write(ctx, userID, itemID, nil)
}

func (w *Writer) write(ctx context.Context, userID int64, itemID string, stars map[string]float64) (Summary, error) {
	var out Summary

	if _, err := uuid.Parse(itemID); err != nil {
		return out, errNotFound()
	}

	err := w.DB.Tx(ctx, func(tx *database.Tx) error {
		now := w.Now().UTC()

		var it lockedItem
		err := tx.QueryRowContext(ctx, `SELECT type, release_date, release_precision, release_status, score, ratings_limited_until
FROM items WHERE id = $1 AND visibility = 'public' FOR UPDATE`, itemID).
			Scan(&it.typ, &it.release, &it.precision, &it.status, &it.before, &it.limited)
		if errors.Is(err, sql.ErrNoRows) {
			return errNotFound()
		}
		if err != nil {
			return err
		}

		var wrote string
		var burstAction int64
		if stars != nil {
			wrote, burstAction, err = w.upsert(ctx, tx, userID, itemID, it, stars, now)
		} else {
			wrote, err = deleteRating(ctx, tx, userID, itemID, it.typ, now)
		}
		if err != nil || wrote == "" {
			if err != nil {
				return err
			}

			out, err = summarize(ctx, tx, itemID)

			return err
		}

		version, err := recomputeItem(ctx, tx, itemID, it.typ, now)
		if err != nil {
			return err
		}
		if burstAction != 0 {
			if err := setActionAfter(ctx, tx, burstAction, itemID); err != nil {
				return err
			}
		}
		if out, err = summarize(ctx, tx, itemID); err != nil {
			return err
		}
		out.FormulaVersion = version

		var after sql.NullFloat64
		if err := tx.QueryRowContext(ctx, `SELECT score FROM items WHERE id = $1`, itemID).Scan(&after); err != nil {
			return err
		}

		w.afterCommit(tx, wrote, userID, itemID, now, it.before, after, version)

		return nil
	})

	return out, err
}

// upsert applies the release gate and writes the rating and its event; it
// returns the event kind.
func (w *Writer) upsert(
	ctx context.Context, tx *database.Tx, userID int64, itemID string, it lockedItem,
	stars map[string]float64, now time.Time,
) (kind string, burstAction int64, err error) {
	var release *time.Time
	if it.release.Valid {
		release = &it.release.Time
	}
	if RatedBeforeRelease(now, release, it.precision, it.status) {
		return "", 0, errs.NewKey(http.StatusConflict, "ratings.not_released", "This title is not released yet.", nil).
			WithExtension("code", CodeNotReleased)
	}

	keys := CategoryKeys[it.typ]
	v := errs.NewValidation()
	var c [Categories]int16
	for i, k := range keys {
		s, ok := stars[k]
		if !ok {
			v.Append("scores."+k, "required")
			continue
		}
		c[i] = int16(math.Round(s * 10))
	}
	for k := range stars {
		if !slices.Contains(keys[:], k) {
			v.Append("scores."+k, "unknown-category")
		}
	}
	if !v.Empty() {
		return "", 0, v
	}

	var before sql.NullFloat64
	err = tx.QueryRowContext(ctx, `SELECT overall FROM ratings WHERE user_id = $1 AND item_id = $2`, userID, itemID).
		Scan(&before)
	kind = "rating_updated"
	isNew := errors.Is(err, sql.ErrNoRows)
	if isNew {
		kind = "rating_created"
	} else if err != nil {
		return "", 0, err
	}

	if err := checkDailyCap(ctx, tx, userID, isNew, now); err != nil {
		return "", 0, err
	}
	if isNew {
		if err := checkItemLimiter(ctx, tx, userID, itemID, it.limited, now); err != nil {
			return "", 0, err
		}
	}

	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO ratings (user_id, item_id, c1, c2, c3, c4, c5, c6, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
ON CONFLICT (user_id, item_id) DO UPDATE SET c1 = $3, c2 = $4, c3 = $5, c4 = $6, c5 = $7, c6 = $8, updated_at = $9`,
		userID,
		itemID,
		c[0],
		c[1],
		c[2],
		c[3],
		c[4],
		c[5],
		now,
	); err != nil {
		return "", 0, err
	}

	if err := appendEvent(ctx, tx, kind, userID, itemID, it.typ, now, before); err != nil {
		return "", 0, err
	}

	if isNew {
		burstAction, err = detectBurst(ctx, tx, itemID, now)
		if err != nil {
			return "", 0, err
		}
	}

	return kind, burstAction, nil
}

func deleteRating(
	ctx context.Context,
	tx *database.Tx,
	userID int64,
	itemID, typ string,
	now time.Time,
) (string, error) {
	var before sql.NullFloat64
	err := tx.QueryRowContext(ctx, `DELETE FROM ratings WHERE user_id = $1 AND item_id = $2 RETURNING overall`, userID, itemID).
		Scan(&before)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO rating_events (at, kind, user_id, item_id, item_type, overall_before)
VALUES ($1, 'rating_deleted', $2, $3, $4, $5)`, now, userID, itemID, typ, before)

	return "rating_deleted", err
}

// appendEvent writes a created or updated event; overall_after is the
// stored overall read back from the row just written.
func appendEvent(
	ctx context.Context, tx *database.Tx, kind string, userID int64, itemID, typ string, now time.Time,
	before sql.NullFloat64,
) error {
	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO rating_events (at, kind, user_id, item_id, item_type, overall_before, overall_after)
VALUES ($1, $2::rating_event_kind, $3, $4, $5, $6,
  (SELECT overall FROM ratings WHERE user_id = $3 AND item_id = $4))`,
		now,
		kind,
		userID,
		itemID,
		typ,
		before,
	)

	return err
}

func (w *Writer) afterCommit(
	tx *database.Tx, kind string, userID int64, itemID string, at time.Time,
	before, after sql.NullFloat64, version int,
) {
	if w.Events == nil {
		return
	}

	var ev any
	switch kind {
	case "rating_created":
		ev = RatingCreated{UserID: userID, ItemID: itemID, At: at}
	case "rating_updated":
		ev = RatingUpdated{UserID: userID, ItemID: itemID, At: at}
	default:
		ev = RatingDeleted{UserID: userID, ItemID: itemID, At: at}
	}
	changed := ItemScoreChanged{
		ItemID:         itemID,
		Before:         nullable(before),
		After:          nullable(after),
		FormulaVersion: version,
	}

	tx.AfterCommit(func(ctx context.Context) {
		// A failing listener must not undo a committed rating.
		_ = w.Events.Dispatch(ctx, ev)
		_ = w.Events.Dispatch(ctx, changed)
	})
}

// loadParams returns the active formula version's m_t and the priors of typ.
func loadParams(ctx context.Context, tx *database.Tx, typ string, at time.Time) (Params, int, error) {
	var p Params
	var version int
	err := tx.QueryRowContext(ctx, `SELECT version, CASE $2 WHEN 'game' THEN m_game WHEN 'movie' THEN m_movie
WHEN 'series' THEN m_series ELSE m_book END FROM score_formula_versions
WHERE activated_at <= $1 ORDER BY version DESC LIMIT 1`, at, typ).Scan(&version, &p.M)
	if err != nil {
		return p, 0, err
	}
	p.Version = version

	for i := range p.Prior {
		p.Prior[i] = defaultPrior
	}

	var c [Categories]float64
	err = tx.QueryRowContext(ctx, `SELECT prior_c[1], prior_c[2], prior_c[3], prior_c[4], prior_c[5], prior_c[6]
FROM score_priors WHERE item_type = $1 AND formula_version = $2`, typ, version).
		Scan(&c[0], &c[1], &c[2], &c[3], &c[4], &c[5])
	if err == nil {
		p.Prior = c
	} else if !errors.Is(err, sql.ErrNoRows) {
		return p, 0, err
	}

	return p, version, nil
}

// summarize reads the public aggregate; the score shows only from 5 ratings.
func summarize(ctx context.Context, tx *database.Tx, itemID string) (Summary, error) {
	var (
		s    Summary
		typ  string
		sc   sql.NullFloat64
		cats [Categories]sql.NullFloat64
		conf sql.NullString
	)
	err := tx.QueryRowContext(ctx, `SELECT type, rating_count, score, score_c1, score_c2, score_c3, score_c4, score_c5, score_c6,
score_confidence FROM items WHERE id = $1`, itemID).
		Scan(&typ, &s.RatingCount, &sc, &cats[0], &cats[1], &cats[2], &cats[3], &cats[4], &cats[5], &conf)
	if err != nil {
		return s, err
	}
	if s.RatingCount < minVisible {
		return s, nil
	}

	stars := func(v float64) float64 { return math.Round(v/10*10) / 10 }
	if sc.Valid {
		v := stars(sc.Float64)
		s.Score = &v
	}
	s.ScoreByCategory = map[string]float64{}
	for i, k := range CategoryKeys[typ] {
		if cats[i].Valid {
			s.ScoreByCategory[k] = stars(cats[i].Float64)
		}
	}
	if conf.Valid {
		s.Confidence = &conf.String
	}

	return s, nil
}

func errNotFound() error {
	return errs.NewKey(
		http.StatusNotFound,
		"ratings.not_found",
		"Item not found.", nil).WithExtension("code", CodeNotFound)
}

func nullable(n sql.NullFloat64) *float64 {
	if !n.Valid {
		return nil
	}

	return &n.Float64
}

// recomputeItem stores formula v1 for the item with the active version's inputs and returns the version.
func recomputeItem(ctx context.Context, tx *database.Tx, itemID, typ string, now time.Time) (int, error) {
	p, version, err := loadParams(ctx, tx, typ, now)
	if err != nil {
		return 0, err
	}

	before, err := changes.LoadScore(ctx, tx, itemID)
	if err != nil {
		return 0, err
	}
	if err := Recompute(ctx, tx, itemID, now, p); err != nil {
		return 0, err
	}
	after, err := changes.LoadScore(ctx, tx, itemID)
	if err != nil {
		return 0, err
	}
	if err := changes.Record(ctx, tx, changes.Change{
		ItemID: itemID, ItemType: typ, Action: "updated", Fields: changes.ScoreFieldNames(before, after),
	}); err != nil {
		return 0, err
	}

	return version, nil
}
