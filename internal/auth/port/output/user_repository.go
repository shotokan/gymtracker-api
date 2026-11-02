package output

import (
	"context"

	userDomain "github.com/shotokan/gymtracker-api/internal/user/domain"
)

// UserRepository defines user lookup operations for authentication.
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*userDomain.User, error)
}
