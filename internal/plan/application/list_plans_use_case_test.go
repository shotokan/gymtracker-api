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

func TestListPlansUseCase_Execute_Success(t *testing.T) {
	t.Parallel()

	repo := new(mockPlanRepo)

	plans := []domain.Plan{
		{
			ID:        "plan-1",
			UserID:    "user-1",
			Name:      "Plan A",
			Active:    true,
			Routines:  []domain.Routine{},
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
		{
			ID:        "plan-2",
			UserID:    "user-1",
			Name:      "Plan B",
			Active:    false,
			Routines:  []domain.Routine{},
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
	}

	repo.On("FindAll", mock.Anything, "user-1", "").Return(plans, nil)

	uc := NewListPlans(repo)

	result, err := uc.Execute(context.Background(), input.ListPlansCommand{
		UserID:     "user-1",
		NameFilter: "",
	})

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "plan-1", result[0].ID)
	assert.Equal(t, "Plan A", result[0].Name)
	assert.True(t, result[0].Active)
	assert.Equal(t, "plan-2", result[1].ID)
	assert.Equal(t, "Plan B", result[1].Name)
	assert.False(t, result[1].Active)
	repo.AssertExpectations(t)
}

func TestListPlansUseCase_Execute_WithFilter(t *testing.T) {
	t.Parallel()

	repo := new(mockPlanRepo)

	plans := []domain.Plan{
		{ID: "plan-1", Name: "Push Pull Legs", Active: true, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
	}

	repo.On("FindAll", mock.Anything, "user-1", "push").Return(plans, nil)

	uc := NewListPlans(repo)

	result, err := uc.Execute(context.Background(), input.ListPlansCommand{
		UserID:     "user-1",
		NameFilter: "push",
	})

	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Push Pull Legs", result[0].Name)
	repo.AssertExpectations(t)
}

func TestListPlansUseCase_Execute_EmptyList(t *testing.T) {
	t.Parallel()

	repo := new(mockPlanRepo)
	repo.On("FindAll", mock.Anything, "user-1", "").Return([]domain.Plan{}, nil)

	uc := NewListPlans(repo)

	result, err := uc.Execute(context.Background(), input.ListPlansCommand{
		UserID:     "user-1",
		NameFilter: "",
	})

	require.NoError(t, err)
	assert.Empty(t, result)
	repo.AssertExpectations(t)
}

func TestListPlansUseCase_Execute_RepositoryError(t *testing.T) {
	t.Parallel()

	repo := new(mockPlanRepo)
	repo.On("FindAll", mock.Anything, "user-1", "").Return(nil, errors.New("dynamo error"))

	uc := NewListPlans(repo)

	_, err := uc.Execute(context.Background(), input.ListPlansCommand{
		UserID:     "user-1",
		NameFilter: "",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "listing plans")
}
