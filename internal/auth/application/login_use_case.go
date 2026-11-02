package application

import (
	"context"
	"fmt"

	"github.com/shotokan/gymtracker-api/internal/auth/domain"
	"github.com/shotokan/gymtracker-api/internal/auth/port/input"
	"github.com/shotokan/gymtracker-api/internal/auth/port/output"
)

type jwtGenerator interface {
	Generate(userID, email string) (string, error)
}

type loginUseCase struct {
	userRepo output.UserRepository
	hasher   output.PasswordHasher
	jwtGen   jwtGenerator
}

// NewLogin constructs the login use case.
func NewLogin(
	userRepo output.UserRepository,
	hasher output.PasswordHasher,
	jwtGen jwtGenerator,
) input.LoginUseCase {
	return &loginUseCase{
		userRepo: userRepo,
		hasher:   hasher,
		jwtGen:   jwtGen,
	}
}

// Execute authenticates a user with email and password.
func (uc *loginUseCase) Execute(ctx context.Context, cmd input.LoginCommand) (input.LoginResult, error) {
	u, err := uc.userRepo.FindByEmail(ctx, cmd.Email)
	if err != nil {
		// Don't distinguish between "not found" and "wrong password"
		// to avoid user enumeration attacks
		return input.LoginResult{}, domain.ErrInvalidCreds
	}

	if err = uc.hasher.Compare(u.PasswordHash, cmd.Password); err != nil {
		return input.LoginResult{}, domain.ErrInvalidCreds
	}

	// Same error as bad credentials: don't reveal that the account exists
	// but was deactivated
	if !u.IsActive() {
		return input.LoginResult{}, domain.ErrInvalidCreds
	}

	token, err := uc.jwtGen.Generate(u.ID, u.Email)
	if err != nil {
		return input.LoginResult{}, fmt.Errorf("generating token: %w", err)
	}

	return input.LoginResult{
		Token:     token,
		ExpiresIn: 86400, // 24 hours
	}, nil
}
