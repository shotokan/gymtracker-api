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

func TestGetPlanUseCase_Execute_Success(t *testing.T) {
	t.Parallel()

	repo := new(mockPlanRepo)

	plan := &domain.Plan{
		ID:        "plan-123",
		UserID:    "user-1",
		Name:      "Mi Plan de Fuerza",
		Active:    true,
		Routines:  []domain.Routine{},
		CreatedAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
	}

	repo.On("FindByID", mock.Anything, "user-1", "plan-123").Return(plan, nil)

	uc := NewGetPlan(repo)

	result, err := uc.Execute(context.Background(), input.GetPlanCommand{
		UserID: "user-1",
		PlanID: "plan-123",
	})

	require.NoError(t, err)
	assert.Equal(t, "plan-123", result.ID)
	assert.Equal(t, "Mi Plan de Fuerza", result.Name)
	assert.True(t, result.Active)
	assert.Equal(t, "2024-01-15T10:00:00Z", result.CreatedAt)
	repo.AssertExpectations(t)
}

func TestGetPlanUseCase_Execute_NotFound(t *testing.T) {
	t.Parallel()

	repo := new(mockPlanRepo)
	repo.On("FindByID", mock.Anything, "user-1", "plan-999").Return(nil, domain.ErrPlanNotFound)

	uc := NewGetPlan(repo)

	_, err := uc.Execute(context.Background(), input.GetPlanCommand{
		UserID: "user-1",
		PlanID: "plan-999",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "finding plan")
	assert.ErrorIs(t, err, domain.ErrPlanNotFound)
}
