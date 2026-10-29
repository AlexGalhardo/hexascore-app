package admin

import (
	"context"
	"time"

	vadmin "github.com/AlexGalhardo/vukem/admin"
	"github.com/AlexGalhardo/vukem/z"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// noInput is the empty edit schema of the read-only resources: with no keys
// the panel has no writable field, whatever a crafted POST carries.
type noInput struct{}

var noInputSchema = z.Struct[noInput](z.Shape{})

// ScoreBurstsResource lists detected rating bursts; resolving one goes through
// the moderation routes (SDD 02 §6), never the generic form.
func ScoreBurstsResource() *vadmin.Resource[models.ScoreBurst] {
	return vadmin.NewResource[models.ScoreBurst](noInputSchema).
		Slug("score-bursts").
		Search("item_id").
		Filter("trigger", "volume", "skew").
		List("id", "item_id", "trigger", "started_at", "limited_until", "resolution", "resolved_at")
}

// RatingFlagsResource lists shadow-weighted ratings. The key is composite
// (user_id, item_id) and the panel resolves a record by one column, so only
// the list is exposed (VF-NEW-P1.D.3.1b).
func RatingFlagsResource() *vadmin.Resource[models.RatingFlag] {
	return vadmin.NewResource[models.RatingFlag](noInputSchema).
		Slug("rating-flags").
		Search("item_id").
		Filter("reason", "burst_hold", "admin_exclude", "admin_downweight").
		List("user_id", "item_id", "shadow_weight", "reason", "burst_id", "created_at")
}

// ScoreActionsResource lists the scoring audit trail.
func ScoreActionsResource() *vadmin.Resource[models.ScoreAction] {
	return vadmin.NewResource[models.ScoreAction](noInputSchema).
		Slug("score-actions").
		Search("kind").
		List("id", "at", "kind", "actor", "item_id", "user_id", "burst_id", "note")
}

// ScoreFormulaVersionsResource lists the activated scoring formulas.
func ScoreFormulaVersionsResource() *vadmin.Resource[models.ScoreFormulaVersion] {
	return vadmin.NewResource[models.ScoreFormulaVersion](noInputSchema).
		Slug("score-formula-versions").
		Search("changelog").
		List("version", "m_game", "m_movie", "m_series", "m_book", "activated_at", "changelog")
}

// readOnlyPolicy lets admins list and view; every write is denied. The
// generic type keeps one policy for the four scoring models.
type readOnlyPolicy[T any] struct {
	now func() time.Time
	// detail is false when a record cannot be addressed by one key.
	detail bool
}

func (p readOnlyPolicy[T]) ViewAny(_ context.Context, u *models.User) bool { return IsAdmin(u, p.now()) }

func (p readOnlyPolicy[T]) View(_ context.Context, u *models.User, _ *T) bool {
	return p.detail && IsAdmin(u, p.now())
}

func (readOnlyPolicy[T]) Create(context.Context, *models.User) bool { return false }

func (readOnlyPolicy[T]) Update(context.Context, *models.User, *T) bool { return false }

func (readOnlyPolicy[T]) Delete(context.Context, *models.User, *T) bool { return false }
