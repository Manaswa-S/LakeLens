package utils

// TODO: should validate password for structural correctness
// password cannot be longer than 72 chars due to bcrypt limitation
func ValidatePassword(pass string) (string, bool) {
	return "", true
}