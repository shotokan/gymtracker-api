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

func TestGetActivePlanUseCase_Execute_Success(t *testing.T) {
	t.Parallel()

	repo := new(mockPlanRepo)

	activePlan := &domain.Plan{
		ID:        "plan-active",
		UserID:    "user-1",
		Name:      "Plan Activo",
		Active:    true,
		Routines:  []domain.Routine{},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	repo.On("FindActive", mock.Anything, "user-1").Return(activePlan, nil)

	uc := NewGetActivePlan(repo)

	result, err := uc.Execute(context.Background(), input.GetActivePlanCommand{
		UserID: "user-1",
	})

	require.NoError(t, err)
	assert.Equal(t, "plan-active", result.ID)
	assert.Equal(t, "Plan Activo", result.Name)
	assert.True(t, result.Active)
	repo.AssertExpectations(t)
}

func TestGetActivePlanUseCase_Execute_NoActivePlan(t *testing.T) {
	t.Parallel()

	repo := new(mockPlanRepo)

	repo.On("FindActive", mock.Anything, "user-1").Return(nil, domain.ErrNoActivePlan)

	uc := NewGetActivePlan(repo)

	_, err := uc.Execute(context.Background(), input.GetActivePlanCommand{
		UserID: "user-1",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "finding active plan")
	assert.ErrorIs(t, err, domain.ErrNoActivePlan)
	repo.AssertExpectations(t)
}
