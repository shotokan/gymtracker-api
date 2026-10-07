package application

import (
	"context"
	"fmt"

	"github.com/shotokan/gymtracker-api/internal/plan/port/input"
	"github.com/shotokan/gymtracker-api/internal/plan/port/output"
)

type deletePlanUseCase struct{ repo output.PlanRepository }

// NewDeletePlan construye el caso de uso para eliminar un plan, inyectando el repositorio necesario.
func NewDeletePlan(repo output.PlanRepository) input.DeletePlanUseCase {
	return &deletePlanUseCase{repo: repo}
}

// Execute elimina el plan especificado para el usuario, validando la existencia del plan y actualizando el repositorio.
func (uc *deletePlanUseCase) Execute(ctx context.Context, cmd input.DeletePlanCommand) error {
	if err := uc.repo.Delete(ctx, cmd.UserID, cmd.PlanID); err != nil {
		return fmt.Errorf("deleting plan: %w", err)
	}
	return nil
}
