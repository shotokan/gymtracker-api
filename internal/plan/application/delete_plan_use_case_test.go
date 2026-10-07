package application

import (
	"context"
	"errors"
	"testing"

	"github.com/shotokan/gymtracker-api/internal/plan/port/input"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDeletePlanUseCase_Execute_Success(t *testing.T) {
	t.Parallel()

	repo := new(mockPlanRepo)

	repo.On("Delete", mock.Anything, "user-1", "plan-123").Return(nil)

	uc := NewDeletePlan(repo)

	err := uc.Execute(context.Background(), input.DeletePlanCommand{
		UserID: "user-1",
		PlanID: "plan-123",
	})

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDeletePlanUseCase_Execute_Error(t *testing.T) {
	t.Parallel()

	repo := new(mockPlanRepo)

	repo.On("Delete", mock.Anything, "user-1", "plan-999").Return(errors.New("dynamo error"))

	uc := NewDeletePlan(repo)

	err := uc.Execute(context.Background(), input.DeletePlanCommand{
		UserID: "user-1",
		PlanID: "plan-999",
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "deleting plan")
	repo.AssertExpectations(t)
}
