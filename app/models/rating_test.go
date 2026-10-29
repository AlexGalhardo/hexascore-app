package models_test

import (
	"slices"
	"testing"

	"github.com/AlexGalhardo/vukem/orm"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

func TestRatingCompositeKey(t *testing.T) {
	t.Parallel()

	want := []string{"user_id", "item_id"}
	if got := orm.PrimaryKeysOf[models.Rating](); !slices.Equal(got, want) {
		t.Fatalf("Rating keys = %v, want %v", got, want)
	}

	if got := orm.PrimaryKeysOf[models.RatingRead](); !slices.Equal(got, want) {
		t.Fatalf("RatingRead keys = %v, want %v", got, want)
	}
}
