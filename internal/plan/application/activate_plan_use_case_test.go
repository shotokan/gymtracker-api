package application

import (
	"context"
	"testing"
	"time"

	"github.com/shotokan/gymtracker-api/internal/plan/domain"
	"github.com/shotokan/gymtracker-api/internal/plan/port/input"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestActivatePlanUseCase_Execute_Success(t *testing.T) {
	t.Parallel()

	repo := new(mockPlanRepo)

	currentPlan := &domain.Plan{
		ID:        "plan-old",
		UserID:    "user-1",
		Name:      "Plan Viejo",
		Active:    true,
		Routines:  []domain.Routine{},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	targetPlan := &domain.Plan{
		ID:        "plan-new",
		UserID:    "user-1",
		Name:      "Plan Nuevo",
		Active:    false,
		Routines:  []domain.Routine{},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	repo.On("FindActive", mock.Anything, "user-1").Return(currentPlan, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(p *domain.Plan) bool {
		return p.ID == "plan-old" && !p.Active // debe desactivarse
	})).Return(nil)
	repo.On("FindByID", mock.Anything, "user-1", "plan-new").Return(targetPlan, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(p *domain.Plan) bool {
		return p.ID == "plan-new" && p.Active // debe activarse
	})).Return(nil)

	uc := NewActivatePlan(repo)

	err := uc.Execute(context.Background(), input.ActivatePlanCommand{
		UserID: "user-1",
		PlanID: "plan-new",
	})

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestActivatePlanUseCase_Execute_AlreadyActive(t *testing.T) {
	t.Parallel()

	repo := new(mockPlanRepo)

	activePlan := &domain.Plan{
		ID:        "plan-123",
		UserID:    "user-1",
		Name:      "Plan Activo",
		Active:    true,
		Routines:  []domain.Routine{},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	repo.On("FindActive", mock.Anything, "user-1").Return(activePlan, nil)
	repo.On("FindByID", mock.Anything, "user-1", "plan-123").Return(activePlan, nil)
	// No debe llamarse Update porque ya está activo

	uc := NewActivatePlan(repo)

	err := uc.Execute(context.Background(), input.ActivatePlanCommand{
		UserID: "user-1",
		PlanID: "plan-123",
	})

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestActivatePlanUseCase_Execute_NoCurrentActive(t *testing.T) {
	t.Parallel()

	repo := new(mockPlanRepo)

	targetPlan := &domain.Plan{
		ID:        "plan-123",
		UserID:    "user-1",
		Name:      "Primer Plan",
		Active:    false,
		Routines:  []domain.Routine{},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	repo.On("FindActive", mock.Anything, "user-1").Return(nil, domain.ErrNoActivePlan)
	repo.On("FindByID", mock.Anything, "user-1", "plan-123").Return(targetPlan, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(p *domain.Plan) bool {
		return p.ID == "plan-123" && p.Active
	})).Return(nil)

	uc := NewActivatePlan(repo)

	err := uc.Execute(context.Background(), input.ActivatePlanCommand{
		UserID: "user-1",
		PlanID: "plan-123",
	})

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestActivatePlanUseCase_Execute_PlanNotFound(t *testing.T) {
	t.Parallel()

	repo := new(mockPlanRepo)

	repo.On("FindActive", mock.Anything, "user-1").Return(nil, domain.ErrNoActivePlan)
	repo.On("FindByID", mock.Anything, "user-1", "plan-999").Return(nil, domain.ErrPlanNotFound)

	uc := NewActivatePlan(repo)

	err := uc.Execute(context.Background(), input.ActivatePlanCommand{
		UserID: "user-1",
		PlanID: "plan-999",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "finding plan")
	assert.ErrorIs(t, err, domain.ErrPlanNotFound)
}
