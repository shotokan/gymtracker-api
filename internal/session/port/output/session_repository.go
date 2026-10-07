package output

import (
	"context"

	"github.com/shotokan/gymtracker-api/internal/session/domain"
)

// SessionRepository defines all persistence operations for the session aggregate.
type SessionRepository interface {
	Save(ctx context.Context, s *domain.Session) error
	FindByID(ctx context.Context, userID, sessionID string) (*domain.Session, error)
	FindAll(ctx context.Context, userID string, limit int) ([]domain.Session, error)
}
