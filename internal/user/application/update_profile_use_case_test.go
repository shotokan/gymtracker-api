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

func TestUpdateProfileUseCase_Execute_Success(t *testing.T) {
	t.Parallel()

	repo := new(mockUserRepo)

	user := &domain.User{
		ID:        "user-123",
		Email:     "ivan@gym.com",
		Name:      "Ivan Old",
		Status:    domain.StatusActive,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	repo.On("FindByID", mock.Anything, "user-123").Return(user, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.ID == "user-123" && u.Name == "Ivan New"
	})).Return(nil)

	uc := NewUpdateProfile(repo)

	result, err := uc.Execute(context.Background(), input.UpdateProfileCommand{
		ActorID: "user-123",
		UserID:  "user-123",
		Name:    "Ivan New",
	})

	require.NoError(t, err)
	assert.Equal(t, "user-123", result.ID)
	assert.Equal(t, "ivan@gym.com", result.Email)
	assert.Equal(t, "Ivan New", result.Name)
	repo.AssertExpectations(t)
}

func TestUpdateProfileUseCase_Execute_UserNotFound(t *testing.T) {
	t.Parallel()

	repo := new(mockUserRepo)

	repo.On("FindByID", mock.Anything, "user-999").Return(nil, domain.ErrNotFound)

	uc := NewUpdateProfile(repo)

	_, err := uc.Execute(context.Background(), input.UpdateProfileCommand{
		ActorID: "user-999",
		UserID:  "user-999",
		Name:    "New Name",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "finding user")
	assert.ErrorIs(t, err, domain.ErrNotFound)
	repo.AssertExpectations(t)
}

func TestUpdateProfileUseCase_Execute_UpdateError(t *testing.T) {
	t.Parallel()

	repo := new(mockUserRepo)

	user := &domain.User{
		ID:        "user-123",
		Email:     "ivan@gym.com",
		Name:      "Ivan Old",
		Status:    domain.StatusActive,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	repo.On("FindByID", mock.Anything, "user-123").Return(user, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(errors.New("dynamo error"))

	uc := NewUpdateProfile(repo)

	_, err := uc.Execute(context.Background(), input.UpdateProfileCommand{
		ActorID: "user-123",
		UserID:  "user-123",
		Name:    "Ivan New",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "updating user")
	repo.AssertExpectations(t)
}

func TestUpdateProfileUseCase_Execute_ForbiddenOnOtherUser(t *testing.T) {
	t.Parallel()

	repo := new(mockUserRepo)

	uc := NewUpdateProfile(repo)

	_, err := uc.Execute(context.Background(), input.UpdateProfileCommand{ActorID: "user-456", UserID: "user-123", Name: "Hacked"})

	require.ErrorIs(t, err, domain.ErrForbidden)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}
