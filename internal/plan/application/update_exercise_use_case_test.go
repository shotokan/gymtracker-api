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

func TestUpdateExerciseUseCase_Execute_Success(t *testing.T) {
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
				},
			},
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	repo.On("FindByID", mock.Anything, "user-1", "plan-123").Return(plan, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(p *domain.Plan) bool {
		// Verify exercise was updated
		if len(p.Routines) == 0 || len(p.Routines[0].Exercises) == 0 {
			return false
		}
		ex := p.Routines[0].Exercises[0]
		return ex.Name == "Press Banca Pesado" && ex.Weight == 80.0
	})).Return(nil)

	uc := NewUpdateExerciseUseCase(repo)

	result, err := uc.Execute(context.Background(), input.UpdateExerciseCommand{
		UserID:     "user-1",
		PlanID:     "plan-123",
		RoutineID:  "routine-1",
		ExerciseID: "ex-1",
		Name:       "Press Banca Pesado",
		Sets:       5,
		Reps:       8,
		Weight:     80.0,
		Notes:      "Progresión",
	})

	require.NoError(t, err)
	assert.Equal(t, "ex-1", result.ID)
	assert.Equal(t, "Press Banca Pesado", result.Name)
	assert.Equal(t, 5, result.Sets)
	assert.Equal(t, 8, result.Reps)
	assert.Equal(t, 80.0, result.Weight)
	repo.AssertExpectations(t)
}

func TestUpdateExerciseUseCase_Execute_PlanNotFound(t *testing.T) {
	t.Parallel()

	repo := new(mockPlanRepo)

	repo.On("FindByID", mock.Anything, "user-1", "plan-999").Return(nil, domain.ErrPlanNotFound)

	uc := NewUpdateExerciseUseCase(repo)

	_, err := uc.Execute(context.Background(), input.UpdateExerciseCommand{
		UserID:     "user-1",
		PlanID:     "plan-999",
		RoutineID:  "routine-1",
		ExerciseID: "ex-1",
		Name:       "Exercise",
		Sets:       3,
		Reps:       10,
		Weight:     50.0,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "find plan")
	assert.ErrorIs(t, err, domain.ErrPlanNotFound)
	repo.AssertExpectations(t)
}

func TestUpdateExerciseUseCase_Execute_RoutineNotFound(t *testing.T) {
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

	uc := NewUpdateExerciseUseCase(repo)

	_, err := uc.Execute(context.Background(), input.UpdateExerciseCommand{
		UserID:     "user-1",
		PlanID:     "plan-123",
		RoutineID:  "routine-999",
		ExerciseID: "ex-1",
		Name:       "Exercise",
		Sets:       3,
		Reps:       10,
		Weight:     50.0,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "updateExercise")
	assert.ErrorIs(t, err, domain.ErrRoutineNotFound)
	repo.AssertExpectations(t)
}

func TestUpdateExerciseUseCase_Execute_ExerciseNotFound(t *testing.T) {
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

	uc := NewUpdateExerciseUseCase(repo)

	_, err := uc.Execute(context.Background(), input.UpdateExerciseCommand{
		UserID:     "user-1",
		PlanID:     "plan-123",
		RoutineID:  "routine-1",
		ExerciseID: "ex-999",
		Name:       "Exercise",
		Sets:       3,
		Reps:       10,
		Weight:     50.0,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "updateExercise")
	assert.ErrorIs(t, err, domain.ErrExerciseNotFound)
	repo.AssertExpectations(t)
}

func TestUpdateExerciseUseCase_Execute_UpdateError(t *testing.T) {
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
				},
			},
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	repo.On("FindByID", mock.Anything, "user-1", "plan-123").Return(plan, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(errors.New("dynamo error"))

	uc := NewUpdateExerciseUseCase(repo)

	_, err := uc.Execute(context.Background(), input.UpdateExerciseCommand{
		UserID:     "user-1",
		PlanID:     "plan-123",
		RoutineID:  "routine-1",
		ExerciseID: "ex-1",
		Name:       "Press Banca Pesado",
		Sets:       5,
		Reps:       8,
		Weight:     80.0,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "update plan")
	repo.AssertExpectations(t)
}
