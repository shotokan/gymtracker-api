package google

import (
	"context"
	"fmt"

	"github.com/shotokan/gymtracker-api/internal/auth/domain"
	"google.golang.org/api/idtoken"
)

// TokenValidator validates Google ID tokens using Google's official library.
type TokenValidator struct {
	clientID string
}

// NewTokenValidator creates a new Google token validator.
// clientID is your Google OAuth 2.0 client ID from Google Cloud Console.
func NewTokenValidator(clientID string) *TokenValidator {
	return &TokenValidator{clientID: clientID}
}

// ValidateToken verifies the Google ID token and extracts user information.
func (v *TokenValidator) ValidateToken(ctx context.Context, idToken string) (*domain.GoogleUser, error) {
	// Validate the token and get the payload
	payload, err := idtoken.Validate(ctx, idToken, v.clientID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidToken, err)
	}

	// Extract user information from the token claims
	googleID, _ := payload.Claims["sub"].(string)
	email, _ := payload.Claims["email"].(string)
	name, _ := payload.Claims["name"].(string)
	picture, _ := payload.Claims["picture"].(string)

	if googleID == "" || email == "" {
		return nil, fmt.Errorf("%w: missing required claims", domain.ErrInvalidToken)
	}

	return &domain.GoogleUser{
		GoogleID: googleID,
		Email:    email,
		Name:     name,
		Picture:  picture,
	}, nil
}
