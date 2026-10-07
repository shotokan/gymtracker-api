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

func TestListExercisesUseCase_Execute_Success(t *testing.T) {
	t.Parallel()

	repo := new(mockPlanRepo)

	plan := &domain.Plan{
		ID:     "plan-123",
		UserID: "user-1",
		Name:   "Mi Plan",
		Routines: []domain.Routine{
			{
				ID:   "routine-1",
				Name: "Push Day",
				Exercises: []domain.Exercise{
					{ID: "ex-1", Name: "Press Banca", Sets: 4, Reps: 10, Weight: 60.0},
					{ID: "ex-2", Name: "Press Inclinado", Sets: 3, Reps: 12, Weight: 50.0},
				},
			},
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	repo.On("FindByID", mock.Anything, "user-1", "plan-123").Return(plan, nil)

	uc := NewListExercisesUseCase(repo)

	results, err := uc.Execute(context.Background(), input.ListExercisesCommand{
		UserID:    "user-1",
		PlanID:    "plan-123",
		RoutineID: "routine-1",
	})

	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "ex-1", results[0].ID)
	assert.Equal(t, "Press Banca", results[0].Name)
	assert.Equal(t, "ex-2", results[1].ID)
	assert.Equal(t, "Press Inclinado", results[1].Name)
	repo.AssertExpectations(t)
}

func TestListExercisesUseCase_Execute_EmptyRoutine(t *testing.T) {
	t.Parallel()

	repo := new(mockPlanRepo)

	plan := &domain.Plan{
		ID:     "plan-123",
		UserID: "user-1",
		Name:   "Mi Plan",
		Routines: []domain.Routine{
			{
				ID:        "routine-1",
				Name:      "Push Day",
				Exercises: []domain.Exercise{},
			},
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	repo.On("FindByID", mock.Anything, "user-1", "plan-123").Return(plan, nil)

	uc := NewListExercisesUseCase(repo)

	results, err := uc.Execute(context.Background(), input.ListExercisesCommand{
		UserID:    "user-1",
		PlanID:    "plan-123",
		RoutineID: "routine-1",
	})

	require.NoError(t, err)
	assert.Empty(t, results)
	repo.AssertExpectations(t)
}

func TestListExercisesUseCase_Execute_PlanNotFound(t *testing.T) {
	t.Parallel()

	repo := new(mockPlanRepo)

	repo.On("FindByID", mock.Anything, "user-1", "plan-999").Return(nil, domain.ErrPlanNotFound)

	uc := NewListExercisesUseCase(repo)

	_, err := uc.Execute(context.Background(), input.ListExercisesCommand{
		UserID:    "user-1",
		PlanID:    "plan-999",
		RoutineID: "routine-1",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "find plan")
	assert.ErrorIs(t, err, domain.ErrPlanNotFound)
	repo.AssertExpectations(t)
}

func TestListExercisesUseCase_Execute_RoutineNotFound(t *testing.T) {
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

	uc := NewListExercisesUseCase(repo)

	_, err := uc.Execute(context.Background(), input.ListExercisesCommand{
		UserID:    "user-1",
		PlanID:    "plan-123",
		RoutineID: "routine-999",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "listExercises")
	assert.ErrorIs(t, err, domain.ErrRoutineNotFound)
	repo.AssertExpectations(t)
}
