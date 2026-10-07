package output

import (
	"context"

	"github.com/shotokan/gymtracker-api/internal/auth/domain"
)

// GoogleTokenValidator validates Google ID tokens and extracts user information.
type GoogleTokenValidator interface {
	// ValidateToken verifies the Google ID token and returns the user info.
	// Returns domain.ErrInvalidToken if the token is invalid or expired.
	ValidateToken(ctx context.Context, idToken string) (*domain.GoogleUser, error)
}
