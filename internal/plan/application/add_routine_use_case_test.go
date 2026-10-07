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

func TestAddRoutineUseCase_Execute_Success(t *testing.T) {
	t.Parallel()

	repo := new(mockPlanRepo)
	idGen := new(mockIDGen)

	plan := &domain.Plan{
		ID:        "plan-123",
		UserID:    "user-1",
		Name:      "Mi Plan",
		Active:    true,
		Routines:  []domain.Routine{},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	repo.On("FindByID", mock.Anything, "user-1", "plan-123").Return(plan, nil)
	idGen.On("New").Return("routine-456")
	repo.On("Update", mock.Anything, mock.MatchedBy(func(p *domain.Plan) bool {
		return len(p.Routines) == 1 &&
			p.Routines[0].ID == "routine-456" &&
			p.Routines[0].Name == "Push Day"
	})).Return(nil)

	uc := NewAddRoutine(repo, idGen)

	result, err := uc.Execute(context.Background(), input.AddRoutineCommand{
		UserID: "user-1",
		PlanID: "plan-123",
		Name:   "Push Day",
	})

	require.NoError(t, err)
	assert.Equal(t, "routine-456", result.ID)
	assert.Equal(t, "Push Day", result.Name)
	repo.AssertExpectations(t)
	idGen.AssertExpectations(t)
}

func TestAddRoutineUseCase_Execute_PlanNotFound(t *testing.T) {
	t.Parallel()

	repo := new(mockPlanRepo)
	idGen := new(mockIDGen)

	repo.On("FindByID", mock.Anything, "user-1", "plan-999").Return(nil, domain.ErrPlanNotFound)

	uc := NewAddRoutine(repo, idGen)

	_, err := uc.Execute(context.Background(), input.AddRoutineCommand{
		UserID: "user-1",
		PlanID: "plan-999",
		Name:   "Push Day",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "finding plan")
	assert.ErrorIs(t, err, domain.ErrPlanNotFound)
}
