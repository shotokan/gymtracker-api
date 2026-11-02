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

func TestDeactivateUseCase_Execute_Success(t *testing.T) {
	t.Parallel()

	repo := new(mockUserRepo)

	user := &domain.User{
		ID:        "user-123",
		Email:     "ivan@gym.com",
		Name:      "Ivan",
		Status:    domain.StatusActive,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	repo.On("FindByID", mock.Anything, "user-123").Return(user, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.ID == "user-123" && u.Status == domain.StatusDeactivated // debe estar desactivado
	})).Return(nil)

	uc := NewDeactivate(repo)

	err := uc.Execute(context.Background(), input.DeactivateCommand{
		ActorID: "user-123",
		UserID:  "user-123",
	})

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDeactivateUseCase_Execute_UserNotFound(t *testing.T) {
	t.Parallel()

	repo := new(mockUserRepo)

	repo.On("FindByID", mock.Anything, "user-999").Return(nil, domain.ErrNotFound)

	uc := NewDeactivate(repo)

	err := uc.Execute(context.Background(), input.DeactivateCommand{
		ActorID: "user-999",
		UserID:  "user-999",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "finding user")
	assert.ErrorIs(t, err, domain.ErrNotFound)
	repo.AssertExpectations(t)
}

func TestDeactivateUseCase_Execute_UpdateError(t *testing.T) {
	t.Parallel()

	repo := new(mockUserRepo)

	user := &domain.User{
		ID:        "user-123",
		Email:     "ivan@gym.com",
		Name:      "Ivan",
		Status:    domain.StatusActive,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	repo.On("FindByID", mock.Anything, "user-123").Return(user, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(errors.New("dynamo error"))

	uc := NewDeactivate(repo)

	err := uc.Execute(context.Background(), input.DeactivateCommand{
		ActorID: "user-123",
		UserID:  "user-123",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "deactivating user")
	repo.AssertExpectations(t)
}

func TestDeactivateUseCase_Execute_AdminDeactivatesOther(t *testing.T) {
	t.Parallel()

	repo := new(mockUserRepo)

	repo.On("FindByID", mock.Anything, "admin-1").
		Return(&domain.User{ID: "admin-1", Role: domain.RoleAdmin, Status: domain.StatusActive}, nil)
	repo.On("FindByID", mock.Anything, "user-123").
		Return(&domain.User{ID: "user-123", Status: domain.StatusActive}, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.ID == "user-123" && u.Status == domain.StatusDeactivated
	})).Return(nil)

	uc := NewDeactivate(repo)

	err := uc.Execute(context.Background(), input.DeactivateCommand{ActorID: "admin-1", UserID: "user-123"})

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDeactivateUseCase_Execute_ForbiddenOnOtherUser(t *testing.T) {
	t.Parallel()

	repo := new(mockUserRepo)

	repo.On("FindByID", mock.Anything, "user-456").
		Return(&domain.User{ID: "user-456", Role: domain.RoleUser, Status: domain.StatusActive}, nil)

	uc := NewDeactivate(repo)

	err := uc.Execute(context.Background(), input.DeactivateCommand{ActorID: "user-456", UserID: "user-123"})

	require.ErrorIs(t, err, domain.ErrForbidden)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestDeactivateUseCase_Execute_LastAdmin(t *testing.T) {
	t.Parallel()

	repo := new(mockUserRepo)
	admin := &domain.User{ID: "admin-1", Role: domain.RoleAdmin, Status: domain.StatusActive}

	repo.On("FindByID", mock.Anything, "admin-1").Return(admin, nil)
	repo.On("FindByStatus", mock.Anything, domain.StatusActive).Return([]domain.User{
		*admin,
		{ID: "user-123", Role: domain.RoleUser, Status: domain.StatusActive},
	}, nil)

	uc := NewDeactivate(repo)

	err := uc.Execute(context.Background(), input.DeactivateCommand{ActorID: "admin-1", UserID: "admin-1"})

	require.ErrorIs(t, err, domain.ErrLastAdmin)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestDeactivateUseCase_Execute_AdminWithAnotherAdmin(t *testing.T) {
	t.Parallel()

	repo := new(mockUserRepo)
	admin := &domain.User{ID: "admin-1", Role: domain.RoleAdmin, Status: domain.StatusActive}

	repo.On("FindByID", mock.Anything, "admin-1").Return(admin, nil)
	repo.On("FindByStatus", mock.Anything, domain.StatusActive).Return([]domain.User{
		*admin,
		{ID: "admin-2", Role: domain.RoleAdmin, Status: domain.StatusActive},
	}, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.ID == "admin-1" && u.Status == domain.StatusDeactivated
	})).Return(nil)

	uc := NewDeactivate(repo)

	err := uc.Execute(context.Background(), input.DeactivateCommand{ActorID: "admin-1", UserID: "admin-1"})

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDeactivateUseCase_Execute_AdminCountError(t *testing.T) {
	t.Parallel()

	repo := new(mockUserRepo)
	admin := &domain.User{ID: "admin-1", Role: domain.RoleAdmin, Status: domain.StatusActive}

	repo.On("FindByID", mock.Anything, "admin-1").Return(admin, nil)
	repo.On("FindByStatus", mock.Anything, domain.StatusActive).Return(nil, errors.New("dynamo error"))

	uc := NewDeactivate(repo)

	err := uc.Execute(context.Background(), input.DeactivateCommand{ActorID: "admin-1", UserID: "admin-1"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "finding active users")
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}
