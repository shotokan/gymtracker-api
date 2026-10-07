package input

import "context"

// ── Login (Email/Password) ────────────────────────────────────────────────

// LoginCommand represents a login request with email and password.
type LoginCommand struct {
	Email    string
	Password string
}

// LoginResult represents the result after successful login.
type LoginResult struct {
	Token     string `json:"token"`      // JWT token for API access
	ExpiresIn int64  `json:"expires_in"` // Token expiration in seconds
}

// LoginUseCase handles authentication via email and password.
type LoginUseCase interface {
	Execute(ctx context.Context, cmd LoginCommand) (LoginResult, error)
}

// ── Google OAuth ──────────────────────────────────────────────────────────

// GoogleAuthCommand is the input for authenticating with Google.
type GoogleAuthCommand struct {
	IDToken string // Google ID token from the client
}

// GoogleAuthResult is the output after successful authentication.
type GoogleAuthResult struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Token  string `json:"token"` // JWT token for API access
}

// GoogleAuthUseCase handles authentication via Google OAuth.
type GoogleAuthUseCase interface {
	Execute(ctx context.Context, cmd GoogleAuthCommand) (GoogleAuthResult, error)
}
