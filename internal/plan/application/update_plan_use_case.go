package application

import (
	"context"
	"fmt"

	"github.com/shotokan/gymtracker-api/internal/plan/port/input"
	"github.com/shotokan/gymtracker-api/internal/plan/port/output"
)

type updatePlanUseCase struct{ repo output.PlanRepository }

// NewUpdatePlan construye el caso de uso para actualizar un plan, inyectando el repositorio necesario.
func NewUpdatePlan(repo output.PlanRepository) input.UpdatePlanUseCase {
	return &updatePlanUseCase{repo: repo}
}

// Execute actualiza el nombre del plan especificado para el usuario, validando la existencia del plan y guardando los cambios en el repositorio.
func (uc *updatePlanUseCase) Execute(ctx context.Context, cmd input.UpdatePlanCommand) (input.PlanResult, error) {
	plan, err := uc.repo.FindByID(ctx, cmd.UserID, cmd.PlanID)
	if err != nil {
		return input.PlanResult{}, fmt.Errorf("finding plan: %w", err)
	}

	plan.Name = cmd.Name

	if err = uc.repo.Update(ctx, plan); err != nil {
		return input.PlanResult{}, fmt.Errorf("updating plan: %w", err)
	}

	return toPlanResult(plan), nil
}
