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
		Active:    true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	repo.On("FindByID", mock.Anything, "user-123").Return(user, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.ID == "user-123" && !u.Active // debe estar desactivado
	})).Return(nil)

	uc := NewDeactivate(repo)

	err := uc.Execute(context.Background(), input.DeactivateCommand{
		UserID: "user-123",
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
		UserID: "user-999",
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
		Active:    true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	repo.On("FindByID", mock.Anything, "user-123").Return(user, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(errors.New("dynamo error"))

	uc := NewDeactivate(repo)

	err := uc.Execute(context.Background(), input.DeactivateCommand{
		UserID: "user-123",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "deactivating user")
	repo.AssertExpectations(t)
}
