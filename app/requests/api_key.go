package requests

import "github.com/AlexGalhardo/vukem/z"

// APIKey names a new API key.
type APIKey struct {
	Name string `json:"name" form:"name"`
}

// APIKeySchema validates APIKey (1 to 40 characters, SDD 05 §5).
var APIKeySchema = z.Struct[APIKey](z.Shape{
	"name": z.String().Trim().Min(1).Max(40),
})
