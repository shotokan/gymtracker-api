package output

// PasswordHasher defines the interface for password hashing and comparison.
type PasswordHasher interface {
	// Compare compares a hashed password with a plain text password.
	// Returns nil if they match, error otherwise.
	Compare(hashedPassword, plainPassword string) error
}
