package requests

import "github.com/AlexGalhardo/vukem/z"

// PasskeyOptions is the name and password sent before a registration ceremony.
// The credential JSON itself is not a form: the verify route reads the raw body.
type PasskeyOptions struct {
	Name     string `json:"name"     form:"name"`
	Password string `json:"password" form:"password"`
}

// PasskeyOptionsSchema validates PasskeyOptions. An empty name becomes "Passkey" in the handler.
var PasskeyOptionsSchema = z.Struct[PasskeyOptions](z.Shape{
	"name":     z.String().Trim().Max(64).Default(""),
	"password": z.String().Max(256).Default(""),
})
