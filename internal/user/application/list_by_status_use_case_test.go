package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shotokan/gymtracker-api/internal/user/domain"
	"github.com/shotokan/gymtracker-api/internal/user/port/input"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func activeAdmin() *domain.User {
	return &domain.User{ID: "admin-1", Role: domain.RoleAdmin, Status: domain.StatusActive}
}

func TestListByStatusUseCase_Execute_SortedOldestFirst(t *testing.T) {
	t.Parallel()

	repo := new(mockUserRepo)
	older := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	newer := older.AddDate(0, 1, 0)

	repo.On("FindByID", mock.Anything, "admin-1").Return(activeAdmin(), nil)
	repo.On("FindByStatus", mock.Anything, domain.StatusPending).Return([]domain.User{
		{ID: "new", Status: domain.StatusPending, CreatedAt: newer, UpdatedAt: newer},
		{ID: "old", Status: domain.StatusPending, CreatedAt: older, UpdatedAt: older},
	}, nil)

	uc := NewListByStatus(repo)

	result, err := uc.Execute(context.Background(), input.ListByStatusCommand{ActorID: "admin-1", Status: domain.StatusPending})

	require.NoError(t, err)
	require.Len(t, result.Users, 2)
	assert.Equal(t, "old", result.Users[0].ID)
	assert.Equal(t, domain.StatusPending, result.Users[0].Status)
	assert.Equal(t, "2024-01-01T00:00:00Z", result.Users[0].CreatedAt)
	assert.Equal(t, "new", result.Users[1].ID)
}

func TestListByStatusUseCase_Execute_Empty(t *testing.T) {
	t.Parallel()

	repo := new(mockUserRepo)
	repo.On("FindByID", mock.Anything, "admin-1").Return(activeAdmin(), nil)
	repo.On("FindByStatus", mock.Anything, domain.StatusDeactivated).Return([]domain.User{}, nil)

	uc := NewListByStatus(repo)

	result, err := uc.Execute(context.Background(), input.ListByStatusCommand{ActorID: "admin-1", Status: domain.StatusDeactivated})

	require.NoError(t, err)
	assert.NotNil(t, result.Users) // se serializa como [] y no como null
	assert.Empty(t, result.Users)
}

func TestListByStatusUseCase_Execute_Forbidden(t *testing.T) {
	t.Parallel()

	repo := new(mockUserRepo)
	repo.On("FindByID", mock.Anything, "user-123").
		Return(&domain.User{ID: "user-123", Role: domain.RoleUser, Status: domain.StatusActive}, nil)

	uc := NewListByStatus(repo)

	_, err := uc.Execute(context.Background(), input.ListByStatusCommand{ActorID: "user-123", Status: domain.StatusPending})

	require.ErrorIs(t, err, domain.ErrForbidden)
	repo.AssertNotCalled(t, "FindByStatus", mock.Anything, mock.Anything)
}

func TestListByStatusUseCase_Execute_RepoError(t *testing.T) {
	t.Parallel()

	repo := new(mockUserRepo)
	repo.On("FindByID", mock.Anything, "admin-1").Return(activeAdmin(), nil)
	repo.On("FindByStatus", mock.Anything, domain.StatusPending).Return(nil, errors.New("dynamo error"))

	uc := NewListByStatus(repo)

	_, err := uc.Execute(context.Background(), input.ListByStatusCommand{ActorID: "admin-1", Status: domain.StatusPending})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "finding users by status")
}
