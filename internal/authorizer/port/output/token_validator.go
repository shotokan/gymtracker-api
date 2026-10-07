package output

import (
	"context"

	"github.com/shotokan/gymtracker-api/internal/authorizer/domain"
)

// TokenValidator valida tokens JWT.
// Interface Segregation: solo expone lo que authorizer necesita.
type TokenValidator interface {
	Validate(ctx context.Context, token string) (*domain.TokenClaims, error)
}
