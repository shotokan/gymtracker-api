// activate_use_case.go
package application

import (
	"context"
	"fmt"

	"github.com/shotokan/gymtracker-api/internal/user/port/input"
	"github.com/shotokan/gymtracker-api/internal/user/port/output"
)

type activateUseCase struct{ repo output.UserRepository }

// NewActivate construye el caso de uso de activación de usuario inyectando su dependencia.
func NewActivate(repo output.UserRepository) input.ActivateUseCase {
	return &activateUseCase{repo: repo}
}

// Execute activa al usuario indicado si quien lo solicita es un administrador activo.
func (uc *activateUseCase) Execute(ctx context.Context, cmd input.ActivateCommand) error {
	if err := requireActiveAdmin(ctx, uc.repo, cmd.ActorID); err != nil {
		return err
	}

	u, err := uc.repo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return fmt.Errorf("finding user: %w", err)
	}

	u.Activate() // lógica de negocio en la entidad

	if err = uc.repo.Update(ctx, u); err != nil {
		return fmt.Errorf("activating user: %w", err)
	}

	return nil
}
