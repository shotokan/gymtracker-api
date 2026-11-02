package domain

import "errors"

// ErrInvalidCreds is returned when email/password authentication fails.
var ErrInvalidCreds = errors.New("invalid credentials")
