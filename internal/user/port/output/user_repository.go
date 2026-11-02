package output

import (
	"context"

	"github.com/shotokan/gymtracker-api/internal/user/domain"
)

// UserRepository define el CRUD completo del usuario.
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindByStatus(ctx context.Context, status string) ([]domain.User, error)
	Save(ctx context.Context, u *domain.User) error
	Update(ctx context.Context, u *domain.User) error
}
