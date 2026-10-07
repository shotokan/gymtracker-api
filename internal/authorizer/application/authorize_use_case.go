package application

import (
	"context"
	"fmt"

	"github.com/shotokan/gymtracker-api/internal/authorizer/domain"
	"github.com/shotokan/gymtracker-api/internal/authorizer/port/input"
	"github.com/shotokan/gymtracker-api/internal/authorizer/port/output"
)

type authorizeUseCase struct {
	validator output.TokenValidator
}

// New construye el caso de uso del authorizer.
func New(validator output.TokenValidator) input.AuthorizeUseCase {
	return &authorizeUseCase{validator: validator}
}

func (uc *authorizeUseCase) Execute(ctx context.Context, cmd input.AuthorizeCommand) (input.AuthorizeResult, error) {
	claims, err := uc.validator.Validate(ctx, cmd.Token)
	if err != nil {
		return input.AuthorizeResult{}, fmt.Errorf("%w: %s", domain.ErrUnauthorized, err)
	}

	return input.AuthorizeResult{
		UserID:  claims.UserID,
		Allowed: true,
	}, nil
}
