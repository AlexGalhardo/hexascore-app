package requests

import "github.com/AlexGalhardo/vukem/z"

// SearchClick is the body of POST /search/events, sent by navigator.sendBeacon.
type SearchClick struct {
	SearchID string `json:"searchId" form:"searchId"`
	RefID    string `json:"refId"    form:"refId"`
	Position int    `json:"position" form:"position"`
	Group    string `json:"group"    form:"group"`
}

// SearchClickSchema validates SearchClick; position is 1-based within its group.
var SearchClickSchema = z.Struct[SearchClick](z.Shape{
	"searchId": z.String().UUID(),
	"refId":    z.String().UUID(),
	"position": z.Int().Min(1).Max(200),
	"group":    z.String().Max(20).Default(""),
})
