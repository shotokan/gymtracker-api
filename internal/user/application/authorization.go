package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/shotokan/gymtracker-api/internal/user/domain"
	"github.com/shotokan/gymtracker-api/internal/user/port/output"
)

// requireActiveAdmin retorna domain.ErrForbidden si actorID no es un administrador activo. El rol se lee del repositorio en cada petición para que retirar el permiso tenga efecto inmediato.
func requireActiveAdmin(ctx context.Context, repo output.UserRepository, actorID string) error {
	actor, err := repo.FindByID(ctx, actorID)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.ErrForbidden
	}
	if err != nil {
		return fmt.Errorf("finding actor: %w", err)
	}
	if !actor.IsActiveAdmin() {
		return domain.ErrForbidden
	}

	return nil
}
