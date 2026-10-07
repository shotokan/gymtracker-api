package domain

import "errors"

// ErrSessionNotFound is returned when a session with the specified ID is not found for a user.
var ErrSessionNotFound = errors.New("session not found")
