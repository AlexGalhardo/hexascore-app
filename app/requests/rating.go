package requests

import "github.com/AlexGalhardo/vukem/z"

// Rating is the rating form: stars (0.5–10, half steps) per category key.
type Rating struct {
	Scores map[string]float64 `json:"scores" form:"scores"`
}

// RatingSchema validates Rating; the category keys of the item's type are
// checked by scoring.Writer once the item is known.
var RatingSchema = z.Struct[Rating](z.Shape{
	"scores": z.Map(z.Float().Min(0.5).Max(10).MultipleOf(0.5)).Min(1).Max(6),
})
