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

type mockGoogleValidator struct {
	mock.Mock
}

func (m *mockGoogleValidator) ValidateToken(ctx context.Context, idToken string) (*domain.GoogleUser, error) {
	args := m.Called(ctx, idToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.GoogleUser), args.Error(1)
}

type mockIDGen struct {
	mock.Mock
}

func (m *mockIDGen) New() string {
	args := m.Called()
	return args.String(0)
}

func TestGoogleAuthUseCase_Execute_NewUser(t *testing.T) {
	t.Parallel()

	validator := new(mockGoogleValidator)
	repo := new(mockUserRepo)
	idGen := new(mockIDGen)
	jwtGen := new(mockJWTGen)

	googleUser := &domain.GoogleUser{
		GoogleID: "google-123",
		Email:    "user@gmail.com",
		Name:     "Test User",
		Picture:  "https://example.com/photo.jpg",
	}

	validator.On("ValidateToken", mock.Anything, "google-id-token").Return(googleUser, nil)
	repo.On("FindByGoogleID", mock.Anything, "google-123").Return(nil, userDomain.ErrNotFound)
	repo.On("FindByEmail", mock.Anything, "user@gmail.com").Return(nil, userDomain.ErrNotFound)
	idGen.On("New").Return("new-user-id")
	repo.On("Save", mock.Anything, mock.MatchedBy(func(u *userDomain.User) bool {
		return u.ID == "new-user-id" &&
			u.Email == "user@gmail.com" &&
			u.GoogleID == "google-123" &&
			u.Name == "Test User" &&
			u.Active == true
	})).Return(nil)
	jwtGen.On("Generate", "new-user-id", "user@gmail.com").Return("jwt-token", nil)

	uc := NewGoogleAuth(validator, repo, idGen, jwtGen)

	result, err := uc.Execute(context.Background(), input.GoogleAuthCommand{
		IDToken: "google-id-token",
	})

	require.NoError(t, err)
	assert.Equal(t, "new-user-id", result.UserID)
	assert.Equal(t, "user@gmail.com", result.Email)
	assert.Equal(t, "Test User", result.Name)
	assert.Equal(t, "jwt-token", result.Token)
	validator.AssertExpectations(t)
	repo.AssertExpectations(t)
	idGen.AssertExpectations(t)
	jwtGen.AssertExpectations(t)
}

func TestGoogleAuthUseCase_Execute_ExistingUserByGoogleID(t *testing.T) {
	t.Parallel()

	validator := new(mockGoogleValidator)
	repo := new(mockUserRepo)
	idGen := new(mockIDGen)
	jwtGen := new(mockJWTGen)

	googleUser := &domain.GoogleUser{
		GoogleID: "google-123",
		Email:    "user@gmail.com",
		Name:     "Test User",
		Picture:  "https://example.com/photo.jpg",
	}

	existingUser := &userDomain.User{
		ID:        "existing-user-id",
		Email:     "user@gmail.com",
		GoogleID:  "google-123",
		Name:      "Test User",
		Active:    true,
		CreatedAt: time.Now().UTC().Add(-24 * time.Hour),
		UpdatedAt: time.Now().UTC(),
	}

	validator.On("ValidateToken", mock.Anything, "google-id-token").Return(googleUser, nil)
	repo.On("FindByGoogleID", mock.Anything, "google-123").Return(existingUser, nil)
	jwtGen.On("Generate", "existing-user-id", "user@gmail.com").Return("jwt-token", nil)

	uc := NewGoogleAuth(validator, repo, idGen, jwtGen)

	result, err := uc.Execute(context.Background(), input.GoogleAuthCommand{
		IDToken: "google-id-token",
	})

	require.NoError(t, err)
	assert.Equal(t, "existing-user-id", result.UserID)
	assert.Equal(t, "user@gmail.com", result.Email)
	assert.Equal(t, "Test User", result.Name)
	assert.Equal(t, "jwt-token", result.Token)
	validator.AssertExpectations(t)
	repo.AssertExpectations(t)
	jwtGen.AssertExpectations(t)
}

func TestGoogleAuthUseCase_Execute_ExistingUserByEmail_LinkGoogle(t *testing.T) {
	t.Parallel()

	validator := new(mockGoogleValidator)
	repo := new(mockUserRepo)
	idGen := new(mockIDGen)
	jwtGen := new(mockJWTGen)

	googleUser := &domain.GoogleUser{
		GoogleID: "google-123",
		Email:    "user@gmail.com",
		Name:     "Test User",
		Picture:  "https://example.com/photo.jpg",
	}

	existingUser := &userDomain.User{
		ID:           "existing-user-id",
		Email:        "user@gmail.com",
		GoogleID:     "", // No Google ID yet
		PasswordHash: "$2a$10$hashedpassword",
		Name:         "Test User",
		Active:       true,
		CreatedAt:    time.Now().UTC().Add(-24 * time.Hour),
		UpdatedAt:    time.Now().UTC(),
	}

	validator.On("ValidateToken", mock.Anything, "google-id-token").Return(googleUser, nil)
	repo.On("FindByGoogleID", mock.Anything, "google-123").Return(nil, userDomain.ErrNotFound)
	repo.On("FindByEmail", mock.Anything, "user@gmail.com").Return(existingUser, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(u *userDomain.User) bool {
		return u.ID == "existing-user-id" && u.GoogleID == "google-123"
	})).Return(nil)
	jwtGen.On("Generate", "existing-user-id", "user@gmail.com").Return("jwt-token", nil)

	uc := NewGoogleAuth(validator, repo, idGen, jwtGen)

	result, err := uc.Execute(context.Background(), input.GoogleAuthCommand{
		IDToken: "google-id-token",
	})

	require.NoError(t, err)
	assert.Equal(t, "existing-user-id", result.UserID)
	assert.Equal(t, "user@gmail.com", result.Email)
	assert.Equal(t, "jwt-token", result.Token)
	validator.AssertExpectations(t)
	repo.AssertExpectations(t)
	jwtGen.AssertExpectations(t)
}

func TestGoogleAuthUseCase_Execute_InvalidToken(t *testing.T) {
	t.Parallel()

	validator := new(mockGoogleValidator)
	repo := new(mockUserRepo)
	idGen := new(mockIDGen)
	jwtGen := new(mockJWTGen)

	validator.On("ValidateToken", mock.Anything, "invalid-token").
		Return(nil, domain.ErrInvalidToken)

	uc := NewGoogleAuth(validator, repo, idGen, jwtGen)

	_, err := uc.Execute(context.Background(), input.GoogleAuthCommand{
		IDToken: "invalid-token",
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidToken)
	validator.AssertExpectations(t)
}
