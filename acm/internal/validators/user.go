package validators

import (
	"strings"
)

func ValidateEmail(email string) bool {
	return strings.Contains(email, "@") &&
		strings.Contains(email, ".")
}
func ValidatePassword(password string) bool {
	return len(password) >= 8
}
