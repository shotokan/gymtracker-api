package application

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/shotokan/gymtracker-api/internal/auth/domain"
	"github.com/shotokan/gymtracker-api/internal/auth/port/input"
	userDomain "github.com/shotokan/gymtracker-api/internal/user/domain"
)

// Mock implementations
type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*userDomain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDomain.User), args.Error(1)
}

func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*userDomain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDomain.User), args.Error(1)
}

func (m *mockUserRepo) FindByGoogleID(ctx context.Context, googleID string) (*userDomain.User, error) {
	args := m.Called(ctx, googleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDomain.User), args.Error(1)
}

func (m *mockUserRepo) Save(ctx context.Context, u *userDomain.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *mockUserRepo) Update(ctx context.Context, u *userDomain.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

type mockPasswordHasher struct {
	mock.Mock
}

func (m *mockPasswordHasher) Compare(hashedPassword, plainPassword string) error {
	args := m.Called(hashedPassword, plainPassword)
	return args.Error(0)
}

type mockJWTGen struct {
	mock.Mock
}

func (m *mockJWTGen) Generate(userID, email string) (string, error) {
	args := m.Called(userID, email)
	return args.String(0), args.Error(1)
}

func TestLoginUseCase_Execute_Success(t *testing.T) {
	t.Parallel()

	repo := new(mockUserRepo)
	hasher := new(mockPasswordHasher)
	jwtGen := new(mockJWTGen)

	user := &userDomain.User{
		ID:           "user-123",
		Email:        "test@example.com",
		PasswordHash: "$2a$10$hashedpassword",
		Name:         "Test User",
		Active:       true,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	repo.On("FindByEmail", mock.Anything, "test@example.com").Return(user, nil)
	hasher.On("Compare", "$2a$10$hashedpassword", "password123").Return(nil)
	jwtGen.On("Generate", "user-123", "test@example.com").Return("jwt-token", nil)

	uc := NewLogin(repo, hasher, jwtGen)

	result, err := uc.Execute(context.Background(), input.LoginCommand{
		Email:    "test@example.com",
		Password: "password123",
	})

	require.NoError(t, err)
	assert.Equal(t, "jwt-token", result.Token)
	assert.Equal(t, int64(86400), result.ExpiresIn)
	repo.AssertExpectations(t)
	hasher.AssertExpectations(t)
	jwtGen.AssertExpectations(t)
}

func TestLoginUseCase_Execute_UserNotFound(t *testing.T) {
	t.Parallel()

	repo := new(mockUserRepo)
	hasher := new(mockPasswordHasher)
	jwtGen := new(mockJWTGen)

	repo.On("FindByEmail", mock.Anything, "notfound@example.com").
		Return(nil, userDomain.ErrNotFound)

	uc := NewLogin(repo, hasher, jwtGen)

	_, err := uc.Execute(context.Background(), input.LoginCommand{
		Email:    "notfound@example.com",
		Password: "password123",
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidCreds)
	repo.AssertExpectations(t)
}

func TestLoginUseCase_Execute_InvalidPassword(t *testing.T) {
	t.Parallel()

	repo := new(mockUserRepo)
	hasher := new(mockPasswordHasher)
	jwtGen := new(mockJWTGen)

	user := &userDomain.User{
		ID:           "user-123",
		Email:        "test@example.com",
		PasswordHash: "$2a$10$hashedpassword",
		Name:         "Test User",
		Active:       true,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	repo.On("FindByEmail", mock.Anything, "test@example.com").Return(user, nil)
	hasher.On("Compare", "$2a$10$hashedpassword", "wrongpassword").
		Return(assert.AnError)

	uc := NewLogin(repo, hasher, jwtGen)

	_, err := uc.Execute(context.Background(), input.LoginCommand{
		Email:    "test@example.com",
		Password: "wrongpassword",
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidCreds)
	repo.AssertExpectations(t)
	hasher.AssertExpectations(t)
}
