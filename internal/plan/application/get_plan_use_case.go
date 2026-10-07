package application

import (
	"context"
	"fmt"

	"github.com/shotokan/gymtracker-api/internal/plan/port/input"
	"github.com/shotokan/gymtracker-api/internal/plan/port/output"
)

type getPlanUseCase struct{ repo output.PlanRepository }

// NewGetPlan construye el caso de uso para obtener un plan, inyectando el repositorio necesario.
func NewGetPlan(repo output.PlanRepository) input.GetPlanUseCase {
	return &getPlanUseCase{repo: repo}
}

// Execute obtiene el plan especificado para el usuario, validando la existencia del plan y devolviendo su representación.
func (uc *getPlanUseCase) Execute(ctx context.Context, cmd input.GetPlanCommand) (input.PlanResult, error) {
	plan, err := uc.repo.FindByID(ctx, cmd.UserID, cmd.PlanID)
	if err != nil {
		return input.PlanResult{}, fmt.Errorf("finding plan: %w", err)
	}
	return toPlanResult(plan), nil
}
