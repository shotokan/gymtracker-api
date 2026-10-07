package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shotokan/gymtracker-api/internal/auth/port/input"
	"github.com/shotokan/gymtracker-api/internal/auth/port/output"
	userDomain "github.com/shotokan/gymtracker-api/internal/user/domain"
	userOutput "github.com/shotokan/gymtracker-api/internal/user/port/output"
)

type idGenerator interface {
	New() string
}

type jwtGenerator interface {
	Generate(userID, email string) (string, error)
}

type googleAuthUseCase struct {
	validator output.GoogleTokenValidator
	userRepo  userOutput.UserRepository
	idGen     idGenerator
	jwtGen    jwtGenerator
}

// NewGoogleAuth constructs the Google authentication use case.
func NewGoogleAuth(
	validator output.GoogleTokenValidator,
	userRepo userOutput.UserRepository,
	idGen idGenerator,
	jwtGen jwtGenerator,
) input.GoogleAuthUseCase {
	return &googleAuthUseCase{
		validator: validator,
		userRepo:  userRepo,
		idGen:     idGen,
		jwtGen:    jwtGen,
	}
}

// Execute authenticates a user via Google OAuth.
// If the user doesn't exist, it creates a new account.
// If the user exists, it updates the login timestamp.
func (uc *googleAuthUseCase) Execute(ctx context.Context, cmd input.GoogleAuthCommand) (input.GoogleAuthResult, error) {
	// Validate the Google ID token
	googleUser, err := uc.validator.ValidateToken(ctx, cmd.IDToken)
	if err != nil {
		return input.GoogleAuthResult{}, fmt.Errorf("validating google token: %w", err)
	}

	// Try to find existing user by Google ID
	user, err := uc.userRepo.FindByGoogleID(ctx, googleUser.GoogleID)
	if err != nil && !errors.Is(err, userDomain.ErrNotFound) {
		return input.GoogleAuthResult{}, fmt.Errorf("finding user by google id: %w", err)
	}

	// If user doesn't exist by Google ID, try by email
	if user == nil {
		user, err = uc.userRepo.FindByEmail(ctx, googleUser.Email)
		if err != nil && !errors.Is(err, userDomain.ErrNotFound) {
			return input.GoogleAuthResult{}, fmt.Errorf("finding user by email: %w", err)
		}
	}

	now := time.Now().UTC()

	// Create new user if doesn't exist
	if user == nil {
		user = &userDomain.User{
			ID:           uc.idGen.New(),
			Email:        googleUser.Email,
			GoogleID:     googleUser.GoogleID,
			Name:         googleUser.Name,
			PasswordHash: "", // No password for Google auth users
			Active:       true,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := uc.userRepo.Save(ctx, user); err != nil {
			return input.GoogleAuthResult{}, fmt.Errorf("saving new user: %w", err)
		}
	} else {
		// Update existing user with Google ID if not set
		if user.GoogleID == "" {
			user.GoogleID = googleUser.GoogleID
			user.UpdatedAt = now
			if err := uc.userRepo.Update(ctx, user); err != nil {
				return input.GoogleAuthResult{}, fmt.Errorf("updating user with google id: %w", err)
			}
		}
	}

	// Generate JWT token for API access
	token, err := uc.jwtGen.Generate(user.ID, user.Email)
	if err != nil {
		return input.GoogleAuthResult{}, fmt.Errorf("generating jwt token: %w", err)
	}

	return input.GoogleAuthResult{
		UserID: user.ID,
		Email:  user.Email,
		Name:   user.Name,
		Token:  token,
	}, nil
}
