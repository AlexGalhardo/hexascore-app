package requests

import "github.com/AlexGalhardo/vukem/z"

// Login is the login form; Identifier is an email or a nickname.
type Login struct {
	Identifier string `json:"identifier" form:"identifier"`
	Password   string `json:"password"   form:"password"`
	Remember   bool   `json:"remember"   form:"remember"`
}

// LoginSchema validates Login.
var LoginSchema = z.Struct[Login](z.Shape{
	"identifier": z.String().Trim().Min(1).Max(254),
	"password":   z.String().Min(1),
	"remember":   z.Bool().Optional(),
})

// Register is the registration form.
type Register struct {
	Name                 string `json:"name"                  form:"name"`
	Email                string `json:"email"                 form:"email"`
	Password             string `json:"password"              form:"password"`
	PasswordConfirmation string `json:"password_confirmation" form:"password_confirmation"`
	// Website is the honeypot: hidden from people, filled by bots (SDD 03 section 7.3).
	Website string `json:"website" form:"website"`
	// FormToken is the signed render time of the sign-up page (minimum fill time, SDD 13 §7.2).
	FormToken string `json:"form_token" form:"form_token"`
}

// RegisterSchema validates Register (password rules live here).
var RegisterSchema = z.Struct[Register](z.Shape{
	"name":                  z.String().Trim().Min(2).Max(100),
	"email":                 z.String().Trim().Email().Max(254),
	"password":              PasswordSchema,
	"password_confirmation": z.String(),
	"website":               z.String().Max(256).Default(""),
	"form_token":            z.String().Max(256).Default(""),
}).Refine(func(r Register) error {
	if r.Password != r.PasswordConfirmation {
		return z.FieldError("password_confirmation", "The password confirmation does not match.")
	}

	return nil
})

// ForgotPassword is the reset link request form.
type ForgotPassword struct {
	Email string `json:"email" form:"email"`
}

// ForgotPasswordSchema validates ForgotPassword.
var ForgotPasswordSchema = z.Struct[ForgotPassword](z.Shape{"email": z.String().Trim().Email()})

// ResetPassword is the new password form.
type ResetPassword struct {
	Token                string `json:"token"                 form:"token"`
	Email                string `json:"email"                 form:"email"`
	Password             string `json:"password"              form:"password"`
	PasswordConfirmation string `json:"password_confirmation" form:"password_confirmation"`
}

// ResetPasswordSchema validates ResetPassword.
var ResetPasswordSchema = z.Struct[ResetPassword](z.Shape{
	"token":                 z.String().Min(1),
	"email":                 z.String().Trim().Email(),
	"password":              PasswordSchema,
	"password_confirmation": z.String(),
}).Refine(func(r ResetPassword) error {
	if r.Password != r.PasswordConfirmation {
		return z.FieldError("password_confirmation", "The password confirmation does not match.")
	}

	return nil
})

// Onboarding is the nickname form.
type Onboarding struct {
	Nickname string `json:"nickname" form:"nickname"`
	// FeedOptIn is honoured only while the feature flag `social` is on (SDD 14 FR-14-18).
	FeedOptIn bool `json:"feedOptIn" form:"feedOptIn"`
}

// OnboardingSchema validates Onboarding; the nickname rule and its error
// codes live in the handler, which answers `invalid-nickname` and `nickname-taken`.
var OnboardingSchema = z.Struct[Onboarding](z.Shape{
	"nickname":  z.String().Trim(),
	"feedOptIn": z.Bool().Default(false),
})
