// deactivate_use_case.go
package application

import (
	"context"
	"fmt"

	"github.com/shotokan/gymtracker-api/internal/user/domain"
	"github.com/shotokan/gymtracker-api/internal/user/port/input"
	"github.com/shotokan/gymtracker-api/internal/user/port/output"
)

type deactivateUseCase struct{ repo output.UserRepository }

// NewDeactivate construye el caso de uso de desactivación de usuario inyectando su dependencia.
func NewDeactivate(repo output.UserRepository) input.DeactivateUseCase {
	return &deactivateUseCase{repo: repo}
}

// Execute realiza la desactivación del usuario, encontrándolo por ID, llamando al método Deactivate de la entidad y actualizando el repositorio. Retorna un error si ocurre algún problema durante el proceso.
func (uc *deactivateUseCase) Execute(ctx context.Context, cmd input.DeactivateCommand) error {
	// Un usuario puede darse de baja a sí mismo; dar de baja a otro requiere ser administrador
	if cmd.ActorID != cmd.UserID {
		if err := requireActiveAdmin(ctx, uc.repo, cmd.ActorID); err != nil {
			return err
		}
	}

	u, err := uc.repo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return fmt.Errorf("finding user: %w", err)
	}

	// Sin administradores nadie podría aprobar cuentas
	if u.IsActiveAdmin() {
		if err = uc.requireAnotherAdmin(ctx); err != nil {
			return err
		}
	}

	u.Deactivate() // lógica de negocio en la entidad

	if err = uc.repo.Update(ctx, u); err != nil {
		return fmt.Errorf("deactivating user: %w", err)
	}

	return nil
}

// requireAnotherAdmin retorna domain.ErrLastAdmin si solo queda un administrador activo.
func (uc *deactivateUseCase) requireAnotherAdmin(ctx context.Context) error {
	active, err := uc.repo.FindByStatus(ctx, domain.StatusActive)
	if err != nil {
		return fmt.Errorf("finding active users: %w", err)
	}

	admins := 0
	for i := range active {
		if active[i].IsActiveAdmin() {
			admins++
		}
	}
	if admins <= 1 {
		return domain.ErrLastAdmin
	}

	return nil
}
