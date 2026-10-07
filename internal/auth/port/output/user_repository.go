package output

import (
	"context"

	userDomain "github.com/shotokan/gymtracker-api/internal/user/domain"
)

// UserRepository defines user lookup operations for authentication.
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*userDomain.User, error)
	FindByGoogleID(ctx context.Context, googleID string) (*userDomain.User, error)
	Save(ctx context.Context, u *userDomain.User) error
	Update(ctx context.Context, u *userDomain.User) error
}
