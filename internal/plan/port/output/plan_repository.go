package output

import (
	"context"

	"github.com/shotokan/gymtracker-api/internal/plan/domain"
)

// PlanRepository defines all persistence operations for the plan aggregate.
// Routines are managed through the plan since they are embedded in the same item.
type PlanRepository interface {
	Save(ctx context.Context, plan *domain.Plan) error
	FindByID(ctx context.Context, userID, planID string) (*domain.Plan, error)
	FindAll(ctx context.Context, userID, nameFilter string) ([]domain.Plan, error)
	FindActive(ctx context.Context, userID string) (*domain.Plan, error)
	Update(ctx context.Context, plan *domain.Plan) error
	Delete(ctx context.Context, userID, planID string) error
}
