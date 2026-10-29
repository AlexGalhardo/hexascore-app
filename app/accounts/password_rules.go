package accounts

import "github.com/AlexGalhardo/hexascore-app/app/requests"

// ValidatePassword applies the new-password rules (SDD 03 FR-03-2) outside a
// form request, for services that set or change a password. Never use it for login.
func ValidatePassword(password string) error {
	return requests.PasswordSchema.Check(password)
}
