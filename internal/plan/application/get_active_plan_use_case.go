package application

import (
	"context"
	"fmt"

	"github.com/shotokan/gymtracker-api/internal/plan/port/input"
	"github.com/shotokan/gymtracker-api/internal/plan/port/output"
)

type getActivePlanUseCase struct{ repo output.PlanRepository }

// NewGetActivePlan construye el caso de uso para obtener el plan activo del usuario, inyectando el repositorio necesario.
func NewGetActivePlan(repo output.PlanRepository) input.GetActivePlanUseCase {
	return &getActivePlanUseCase{repo: repo}
}

// Execute obtiene el plan activo para el usuario, validando la existencia de un plan activo y devolviendo su representación.
func (uc *getActivePlanUseCase) Execute(ctx context.Context, cmd input.GetActivePlanCommand) (input.PlanResult, error) {
	plan, err := uc.repo.FindActive(ctx, cmd.UserID)
	if err != nil {
		return input.PlanResult{}, fmt.Errorf("finding active plan: %w", err)
	}
	return toPlanResult(plan), nil
}
