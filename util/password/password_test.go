package password_test

import (
	"testing"

	. "github.com/ngavinsir/notification-service/util/password"
)

func TestHashPassword_IsMatch(t *testing.T) {
	tests := []string{
		"ajsbdhjasbdas",
		"98248u52jfjsSDSf",
		"ZKJsiasnzx9891:}{>?:",
	}

	for _, password := range(tests) {
		t.Run("Match password hash", func(t *testing.T) {
			hashedPassword, err := HashPassword(password)
			if err != nil {
				t.Error(err)
			}
			if !CheckPasswordHash(password, hashedPassword) {
				t.Errorf("not match for password %s and hash %s", password, hashedPassword)
			}
		})
	}
}