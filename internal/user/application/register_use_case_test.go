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

func TestRegisterUseCase_Execute_Success(t *testing.T) {
	t.Parallel()

	repo := new(mockUserRepo)
	hasher := new(mockPasswordHasher)
	idGen := new(mockIDGen)

	// Email doesn't exist
	repo.On("FindByEmail", mock.Anything, "ivan@gym.com").Return(nil, domain.ErrNotFound)
	hasher.On("Hash", "secret123").Return("hashedPassword", nil)
	idGen.On("New").Return("user-new")
	repo.On("Save", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.ID == "user-new" &&
			u.Email == "ivan@gym.com" &&
			u.Name == "Ivan" &&
			u.PasswordHash == "hashedPassword" &&
			u.Role == domain.RoleUser &&
			u.Status == domain.StatusPending
	})).Return(nil)

	uc := NewRegister(repo, hasher, idGen)

	result, err := uc.Execute(context.Background(), input.RegisterCommand{
		Email:    "ivan@gym.com",
		Password: "secret123",
		Name:     "Ivan",
	})

	require.NoError(t, err)
	assert.Equal(t, "user-new", result.ID)
	assert.Equal(t, "ivan@gym.com", result.Email)
	assert.Equal(t, "Ivan", result.Name)
	repo.AssertExpectations(t)
	hasher.AssertExpectations(t)
	idGen.AssertExpectations(t)
}

func TestRegisterUseCase_Execute_EmailAlreadyExists(t *testing.T) {
	t.Parallel()

	repo := new(mockUserRepo)
	hasher := new(mockPasswordHasher)
	idGen := new(mockIDGen)

	existingUser := &domain.User{
		ID:    "user-existing",
		Email: "ivan@gym.com",
		Name:  "Ivan Old",
	}

	repo.On("FindByEmail", mock.Anything, "ivan@gym.com").Return(existingUser, nil)

	uc := NewRegister(repo, hasher, idGen)

	_, err := uc.Execute(context.Background(), input.RegisterCommand{
		Email:    "ivan@gym.com",
		Password: "secret123",
		Name:     "Ivan New",
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrAlreadyExists)
	repo.AssertExpectations(t)
}

func TestRegisterUseCase_Execute_HashError(t *testing.T) {
	t.Parallel()

	repo := new(mockUserRepo)
	hasher := new(mockPasswordHasher)
	idGen := new(mockIDGen)

	repo.On("FindByEmail", mock.Anything, "ivan@gym.com").Return(nil, domain.ErrNotFound)
	hasher.On("Hash", "secret123").Return("", errors.New("hash error"))

	uc := NewRegister(repo, hasher, idGen)

	_, err := uc.Execute(context.Background(), input.RegisterCommand{
		Email:    "ivan@gym.com",
		Password: "secret123",
		Name:     "Ivan",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "hashing password")
	repo.AssertExpectations(t)
	hasher.AssertExpectations(t)
}

func TestRegisterUseCase_Execute_SaveError(t *testing.T) {
	t.Parallel()

	repo := new(mockUserRepo)
	hasher := new(mockPasswordHasher)
	idGen := new(mockIDGen)

	repo.On("FindByEmail", mock.Anything, "ivan@gym.com").Return(nil, domain.ErrNotFound)
	hasher.On("Hash", "secret123").Return("hashedPassword", nil)
	idGen.On("New").Return("user-new")
	repo.On("Save", mock.Anything, mock.Anything).Return(errors.New("dynamo error"))

	uc := NewRegister(repo, hasher, idGen)

	_, err := uc.Execute(context.Background(), input.RegisterCommand{
		Email:    "ivan@gym.com",
		Password: "secret123",
		Name:     "Ivan",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "saving user")
	repo.AssertExpectations(t)
	hasher.AssertExpectations(t)
	idGen.AssertExpectations(t)
}
