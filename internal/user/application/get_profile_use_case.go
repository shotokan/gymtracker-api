// get_profile_use_case.go
package application

import (
	"context"
	"fmt"

	"github.com/shotokan/gymtracker-api/internal/user/port/input"
	"github.com/shotokan/gymtracker-api/internal/user/port/output"
)

type getProfileUseCase struct{ repo output.UserRepository }

// NewGetProfile construye el caso de uso de obtención del perfil de usuario inyectando su dependencia.
func NewGetProfile(repo output.UserRepository) input.GetProfileUseCase {
	return &getProfileUseCase{repo: repo}
}

// Execute realiza la obtención del perfil de usuario, encontrándolo por ID. Retorna el perfil o un error si ocurre algún problema durante el proceso.
func (uc *getProfileUseCase) Execute(ctx context.Context, cmd input.GetProfileCommand) (input.GetProfileResult, error) {
	u, err := uc.repo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return input.GetProfileResult{}, fmt.Errorf("finding user: %w", err)
	}

	return input.GetProfileResult{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Active:    u.Active,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
