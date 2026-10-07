package bcrypt

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// PasswordHasher implements password hashing using bcrypt.
type PasswordHasher struct{}

// NewPasswordHasher creates a new bcrypt password hasher.
func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{}
}

// Compare compares a hashed password with a plain text password.
func (h *PasswordHasher) Compare(hashedPassword, plainPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	if err != nil {
		return fmt.Errorf("password mismatch: %w", err)
	}
	return nil
}
