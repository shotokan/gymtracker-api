package application

import (
	"context"
	"errors"
	"testing"

	"github.com/shotokan/gymtracker-api/internal/user/domain"
	"github.com/shotokan/gymtracker-api/internal/user/port/input"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestActivateUseCase_Execute_Success(t *testing.T) {
	t.Parallel()

	repo := new(mockUserRepo)

	admin := &domain.User{ID: "admin-1", Role: domain.RoleAdmin, Status: domain.StatusActive}
	pending := &domain.User{ID: "user-123", Email: "ivan@gym.com", Role: domain.RoleUser, Status: domain.StatusPending}

	repo.On("FindByID", mock.Anything, "admin-1").Return(admin, nil)
	repo.On("FindByID", mock.Anything, "user-123").Return(pending, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.ID == "user-123" && u.Status == domain.StatusActive && !u.UpdatedAt.IsZero()
	})).Return(nil)

	uc := NewActivate(repo)

	err := uc.Execute(context.Background(), input.ActivateCommand{ActorID: "admin-1", UserID: "user-123"})

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestActivateUseCase_Execute_Forbidden(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		actor *domain.User
	}{
		{"usuario sin rol de admin", &domain.User{ID: "actor", Role: domain.RoleUser, Status: domain.StatusActive}},
		{"usuario sin rol asignado", &domain.User{ID: "actor", Status: domain.StatusActive}},
		{"admin desactivado", &domain.User{ID: "actor", Role: domain.RoleAdmin, Status: domain.StatusDeactivated}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := new(mockUserRepo)
			repo.On("FindByID", mock.Anything, "actor").Return(tt.actor, nil)

			uc := NewActivate(repo)

			err := uc.Execute(context.Background(), input.ActivateCommand{ActorID: "actor", UserID: "user-123"})

			require.ErrorIs(t, err, domain.ErrForbidden)
			repo.AssertNotCalled(t, "FindByID", mock.Anything, "user-123")
			repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
		})
	}
}

func TestActivateUseCase_Execute_ActorNotFound(t *testing.T) {
	t.Parallel()

	repo := new(mockUserRepo)
	repo.On("FindByID", mock.Anything, "ghost").Return(nil, domain.ErrNotFound)

	uc := NewActivate(repo)

	err := uc.Execute(context.Background(), input.ActivateCommand{ActorID: "ghost", UserID: "user-123"})

	require.ErrorIs(t, err, domain.ErrForbidden)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestActivateUseCase_Execute_ActorLookupError(t *testing.T) {
	t.Parallel()

	repo := new(mockUserRepo)
	repo.On("FindByID", mock.Anything, "admin-1").Return(nil, errors.New("dynamo error"))

	uc := NewActivate(repo)

	err := uc.Execute(context.Background(), input.ActivateCommand{ActorID: "admin-1", UserID: "user-123"})

	require.Error(t, err)
	assert.NotErrorIs(t, err, domain.ErrForbidden)
	assert.Contains(t, err.Error(), "finding actor")
}

func TestActivateUseCase_Execute_UserNotFound(t *testing.T) {
	t.Parallel()

	repo := new(mockUserRepo)
	admin := &domain.User{ID: "admin-1", Role: domain.RoleAdmin, Status: domain.StatusActive}

	repo.On("FindByID", mock.Anything, "admin-1").Return(admin, nil)
	repo.On("FindByID", mock.Anything, "user-999").Return(nil, domain.ErrNotFound)

	uc := NewActivate(repo)

	err := uc.Execute(context.Background(), input.ActivateCommand{ActorID: "admin-1", UserID: "user-999"})

	require.ErrorIs(t, err, domain.ErrNotFound)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestActivateUseCase_Execute_UpdateError(t *testing.T) {
	t.Parallel()

	repo := new(mockUserRepo)
	admin := &domain.User{ID: "admin-1", Role: domain.RoleAdmin, Status: domain.StatusActive}
	pending := &domain.User{ID: "user-123", Status: domain.StatusPending}

	repo.On("FindByID", mock.Anything, "admin-1").Return(admin, nil)
	repo.On("FindByID", mock.Anything, "user-123").Return(pending, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(errors.New("dynamo error"))

	uc := NewActivate(repo)

	err := uc.Execute(context.Background(), input.ActivateCommand{ActorID: "admin-1", UserID: "user-123"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "activating user")
}
