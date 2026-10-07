// update_profile_use_case.go
package application

import (
	"context"
	"fmt"

	"github.com/shotokan/gymtracker-api/internal/user/port/input"
	"github.com/shotokan/gymtracker-api/internal/user/port/output"
)

type updateProfileUseCase struct{ repo output.UserRepository }

// NewUpdateProfile construye el caso de uso de actualización del perfil de usuario inyectando su dependencia.
func NewUpdateProfile(repo output.UserRepository) input.UpdateProfileUseCase {
	return &updateProfileUseCase{repo: repo}
}

// Execute realiza la actualización del perfil de usuario, encontrándolo por ID, llamando al método UpdateName de la entidad y actualizando el repositorio. Retorna el perfil actualizado o un error si ocurre algún problema durante el proceso.
func (uc *updateProfileUseCase) Execute(ctx context.Context, cmd input.UpdateProfileCommand) (input.UpdateProfileResult, error) {
	u, err := uc.repo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return input.UpdateProfileResult{}, fmt.Errorf("finding user: %w", err)
	}

	u.UpdateName(cmd.Name) // lógica de negocio en la entidad

	if err = uc.repo.Update(ctx, u); err != nil {
		return input.UpdateProfileResult{}, fmt.Errorf("updating user: %w", err)
	}

	return input.UpdateProfileResult{ID: u.ID, Email: u.Email, Name: u.Name}, nil
}
