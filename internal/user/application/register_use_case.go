package application

import (
	"context"
	"fmt"
	"time"

	"github.com/shotokan/gymtracker-api/internal/user/domain"
	"github.com/shotokan/gymtracker-api/internal/user/port/input"
	"github.com/shotokan/gymtracker-api/internal/user/port/output"
)

type registerUseCase struct {
	repo   output.UserRepository
	hasher output.PasswordHasher
	idGen  output.IDGenerator
}

// NewRegister builds the register use case injecting its dependencies.
func NewRegister(
	repo output.UserRepository,
	hasher output.PasswordHasher,
	idGen output.IDGenerator,
) input.RegisterUseCase {
	return &registerUseCase{repo: repo, hasher: hasher, idGen: idGen}
}

func (uc *registerUseCase) Execute(ctx context.Context, cmd input.RegisterCommand) (input.RegisterResult, error) {
	// If FindByEmail succeeds, the email is already taken.
	if _, err := uc.repo.FindByEmail(ctx, cmd.Email); err == nil {
		return input.RegisterResult{}, domain.ErrAlreadyExists
	}

	hash, err := uc.hasher.Hash(cmd.Password)
	if err != nil {
		return input.RegisterResult{}, fmt.Errorf("hashing password: %w", err)
	}

	now := time.Now().UTC()
	u := &domain.User{
		ID:           uc.idGen.New(),
		Email:        cmd.Email,
		PasswordHash: hash,
		Name:         cmd.Name,
		Active:       true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err = uc.repo.Save(ctx, u); err != nil {
		return input.RegisterResult{}, fmt.Errorf("saving user: %w", err)
	}

	return input.RegisterResult{ID: u.ID, Email: u.Email, Name: u.Name}, nil
}
