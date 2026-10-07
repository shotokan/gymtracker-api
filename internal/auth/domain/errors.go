package domain

import "errors"

// ErrInvalidCreds is returned when email/password authentication fails.
var ErrInvalidCreds = errors.New("invalid credentials")

// ErrInvalidToken is returned when the Google ID token is invalid or expired.
var ErrInvalidToken = errors.New("invalid or expired token")
