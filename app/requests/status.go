package requests

import "github.com/AlexGalhardo/vukem/z"

// Status is the item status form.
type Status struct {
	Status string `json:"status" form:"status"`
}

// StatusSchema validates Status.
var StatusSchema = z.Struct[Status](z.Shape{
	"status": z.Enum("want", "in_progress", "completed", "dropped"),
})

// TimeToBeat is the time-to-beat form: hours per style, 0.5 steps within 0.5–1,000; at least one is required.
type TimeToBeat struct {
	Main          *float64 `json:"mainHours"          form:"mainHours"`
	Extras        *float64 `json:"extrasHours"        form:"extrasHours"`
	Completionist *float64 `json:"completionistHours" form:"completionistHours"`
	Speedrun      *float64 `json:"speedrunHours"      form:"speedrunHours"`
}

var hours = z.Float().Min(0.5).Max(1000).MultipleOf(0.5).Nullable().Optional()

// TimeToBeatSchema validates TimeToBeat; the "at least one style" rule is checked by the controller.
var TimeToBeatSchema = z.Struct[TimeToBeat](z.Shape{
	"mainHours":          hours,
	"extrasHours":        hours,
	"completionistHours": hours,
	"speedrunHours":      hours,
})
