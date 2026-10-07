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

func TestAddExerciseUseCase_Execute_Success(t *testing.T) {
	t.Parallel()

	repo := new(mockPlanRepo)
	idGen := new(mockIDGen)

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
	idGen.On("New").Return("exercise-new")
	repo.On("Update", mock.Anything, mock.MatchedBy(func(p *domain.Plan) bool {
		// Verify exercise was added to routine
		if len(p.Routines) == 0 {
			return false
		}
		return len(p.Routines[0].Exercises) == 1 && p.Routines[0].Exercises[0].Name == "Press Banca"
	})).Return(nil)

	uc := NewAddExerciseUseCase(repo, idGen)

	result, err := uc.Execute(context.Background(), input.AddExerciseCommand{
		UserID:    "user-1",
		PlanID:    "plan-123",
		RoutineID: "routine-1",
		Name:      "Press Banca",
		Sets:      4,
		Reps:      10,
		Weight:    60.0,
		Notes:     "Controlado",
	})

	require.NoError(t, err)
	assert.Equal(t, "exercise-new", result.ID)
	assert.Equal(t, "Press Banca", result.Name)
	assert.Equal(t, 4, result.Sets)
	assert.Equal(t, 10, result.Reps)
	assert.Equal(t, 60.0, result.Weight)
	repo.AssertExpectations(t)
	idGen.AssertExpectations(t)
}

func TestAddExerciseUseCase_Execute_PlanNotFound(t *testing.T) {
	t.Parallel()

	repo := new(mockPlanRepo)
	idGen := new(mockIDGen)

	repo.On("FindByID", mock.Anything, "user-1", "plan-999").Return(nil, domain.ErrPlanNotFound)

	uc := NewAddExerciseUseCase(repo, idGen)

	_, err := uc.Execute(context.Background(), input.AddExerciseCommand{
		UserID:    "user-1",
		PlanID:    "plan-999",
		RoutineID: "routine-1",
		Name:      "Exercise",
		Sets:      3,
		Reps:      10,
		Weight:    50.0,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "find plan")
	assert.ErrorIs(t, err, domain.ErrPlanNotFound)
	repo.AssertExpectations(t)
}

func TestAddExerciseUseCase_Execute_RoutineNotFound(t *testing.T) {
	t.Parallel()

	repo := new(mockPlanRepo)
	idGen := new(mockIDGen)

	plan := &domain.Plan{
		ID:        "plan-123",
		UserID:    "user-1",
		Name:      "Mi Plan",
		Routines:  []domain.Routine{},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	repo.On("FindByID", mock.Anything, "user-1", "plan-123").Return(plan, nil)
	idGen.On("New").Return("exercise-new")

	uc := NewAddExerciseUseCase(repo, idGen)

	_, err := uc.Execute(context.Background(), input.AddExerciseCommand{
		UserID:    "user-1",
		PlanID:    "plan-123",
		RoutineID: "routine-999",
		Name:      "Exercise",
		Sets:      3,
		Reps:      10,
		Weight:    50.0,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "addExercise")
	assert.ErrorIs(t, err, domain.ErrRoutineNotFound)
	repo.AssertExpectations(t)
	idGen.AssertExpectations(t)
}

func TestAddExerciseUseCase_Execute_UpdateError(t *testing.T) {
	t.Parallel()

	repo := new(mockPlanRepo)
	idGen := new(mockIDGen)

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
	idGen.On("New").Return("exercise-new")
	repo.On("Update", mock.Anything, mock.Anything).Return(errors.New("dynamo error"))

	uc := NewAddExerciseUseCase(repo, idGen)

	_, err := uc.Execute(context.Background(), input.AddExerciseCommand{
		UserID:    "user-1",
		PlanID:    "plan-123",
		RoutineID: "routine-1",
		Name:      "Press Banca",
		Sets:      4,
		Reps:      10,
		Weight:    60.0,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "update plan")
	repo.AssertExpectations(t)
	idGen.AssertExpectations(t)
}
