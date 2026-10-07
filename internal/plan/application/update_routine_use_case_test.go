package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shotokan/gymtracker-api/internal/plan/domain"
	"github.com/shotokan/gymtracker-api/internal/plan/port/input"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdateRoutineUseCase_Execute_Success(t *testing.T) {
	t.Parallel()

	repo := new(mockPlanRepo)

	plan := &domain.Plan{
		ID:     "plan-123",
		UserID: "user-1",
		Name:   "Mi Plan",
		Routines: []domain.Routine{
			{
				ID:        "routine-1",
				Name:      "Rutina Vieja",
				Exercises: []domain.Exercise{},
			},
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	repo.On("FindByID", mock.Anything, "user-1", "plan-123").Return(plan, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(p *domain.Plan) bool {
		// Verify routine was updated
		if len(p.Routines) == 0 {
			return false
		}
		return p.Routines[0].Name == "Rutina Actualizada"
	})).Return(nil)

	uc := NewUpdateRoutine(repo)

	result, err := uc.Execute(context.Background(), input.UpdateRoutineCommand{
		UserID:    "user-1",
		PlanID:    "plan-123",
		RoutineID: "routine-1",
		Name:      "Rutina Actualizada",
	})

	require.NoError(t, err)
	assert.Equal(t, "routine-1", result.ID)
	assert.Equal(t, "Rutina Actualizada", result.Name)
	repo.AssertExpectations(t)
}

func TestUpdateRoutineUseCase_Execute_PlanNotFound(t *testing.T) {
	t.Parallel()

	repo := new(mockPlanRepo)

	repo.On("FindByID", mock.Anything, "user-1", "plan-999").Return(nil, domain.ErrPlanNotFound)

	uc := NewUpdateRoutine(repo)

	_, err := uc.Execute(context.Background(), input.UpdateRoutineCommand{
		UserID:    "user-1",
		PlanID:    "plan-999",
		RoutineID: "routine-1",
		Name:      "Nueva Rutina",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "finding plan")
	assert.ErrorIs(t, err, domain.ErrPlanNotFound)
	repo.AssertExpectations(t)
}

func TestUpdateRoutineUseCase_Execute_RoutineNotFound(t *testing.T) {
	t.Parallel()

	repo := new(mockPlanRepo)

	plan := &domain.Plan{
		ID:        "plan-123",
		UserID:    "user-1",
		Name:      "Mi Plan",
		Routines:  []domain.Routine{},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	repo.On("FindByID", mock.Anything, "user-1", "plan-123").Return(plan, nil)

	uc := NewUpdateRoutine(repo)

	_, err := uc.Execute(context.Background(), input.UpdateRoutineCommand{
		UserID:    "user-1",
		PlanID:    "plan-123",
		RoutineID: "routine-999",
		Name:      "Nueva Rutina",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "updating routine")
	assert.ErrorIs(t, err, domain.ErrRoutineNotFound)
	repo.AssertExpectations(t)
}

func TestUpdateRoutineUseCase_Execute_UpdateError(t *testing.T) {
	t.Parallel()

	repo := new(mockPlanRepo)

	plan := &domain.Plan{
		ID:     "plan-123",
		UserID: "user-1",
		Name:   "Mi Plan",
		Routines: []domain.Routine{
			{
				ID:        "routine-1",
				Name:      "Rutina Original",
				Exercises: []domain.Exercise{},
			},
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	repo.On("FindByID", mock.Anything, "user-1", "plan-123").Return(plan, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(errors.New("dynamo error"))

	uc := NewUpdateRoutine(repo)

	_, err := uc.Execute(context.Background(), input.UpdateRoutineCommand{
		UserID:    "user-1",
		PlanID:    "plan-123",
		RoutineID: "routine-1",
		Name:      "Rutina Actualizada",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "updating plan")
	repo.AssertExpectations(t)
}
