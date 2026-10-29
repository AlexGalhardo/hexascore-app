package requests

import "github.com/AlexGalhardo/vukem/z"

// Contact is the contact form (SDD 14 §7.3). Website is the honeypot: people never see it. Token is the signed
// form token. Name and Email are ignored for signed-in users.
type Contact struct {
	Name    string `json:"name"    form:"name"`
	Email   string `json:"email"   form:"email"`
	Subject string `json:"subject" form:"subject"`
	Message string `json:"message" form:"message"`
	Website string `json:"website" form:"website"`
	Token   string `json:"token"   form:"token"`
}

// ContactSchema validates Contact (legacy contactSchema bounds: message 32 to 512 characters).
var ContactSchema = z.Struct[Contact](z.Shape{
	"name":    z.String().Trim().Max(100).Default(""),
	"email":   z.String().Trim().Max(254).Default(""),
	"subject": z.Enum("question", "feedback", "bug"),
	"message": z.String().Trim().Min(32).Max(512),
	"website": z.String().Max(200).Default(""),
	"token":   z.String().Max(200).Default(""),
})
