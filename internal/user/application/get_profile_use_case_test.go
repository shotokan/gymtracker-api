package application

import (
	"context"
	"testing"
	"time"

	"github.com/shotokan/gymtracker-api/internal/user/domain"
	"github.com/shotokan/gymtracker-api/internal/user/port/input"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetProfileUseCase_Execute_Success(t *testing.T) {
	t.Parallel()

	repo := new(mockUserRepo)

	user := &domain.User{
		ID:        "user-123",
		Email:     "ivan@gym.com",
		Name:      "Ivan",
		Active:    true,
		CreatedAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
	}

	repo.On("FindByID", mock.Anything, "user-123").Return(user, nil)

	uc := NewGetProfile(repo)

	result, err := uc.Execute(context.Background(), input.GetProfileCommand{
		UserID: "user-123",
	})

	require.NoError(t, err)
	assert.Equal(t, "user-123", result.ID)
	assert.Equal(t, "ivan@gym.com", result.Email)
	assert.Equal(t, "Ivan", result.Name)
	assert.True(t, result.Active)
	assert.Equal(t, "2024-01-15T10:00:00Z", result.CreatedAt)
	repo.AssertExpectations(t)
}

func TestGetProfileUseCase_Execute_UserNotFound(t *testing.T) {
	t.Parallel()

	repo := new(mockUserRepo)

	repo.On("FindByID", mock.Anything, "user-999").Return(nil, domain.ErrNotFound)

	uc := NewGetProfile(repo)

	_, err := uc.Execute(context.Background(), input.GetProfileCommand{
		UserID: "user-999",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "finding user")
	assert.ErrorIs(t, err, domain.ErrNotFound)
	repo.AssertExpectations(t)
}
