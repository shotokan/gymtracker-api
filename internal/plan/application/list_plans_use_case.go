package application

import (
	"context"
	"fmt"

	"github.com/shotokan/gymtracker-api/internal/plan/port/input"
	"github.com/shotokan/gymtracker-api/internal/plan/port/output"
)

type listPlansUseCase struct{ repo output.PlanRepository }

// NewListPlans construye el caso de uso para listar los planes del usuario, inyectando el repositorio necesario.
func NewListPlans(repo output.PlanRepository) input.ListPlansUseCase {
	return &listPlansUseCase{repo: repo}
}

// Execute obtiene la lista de planes para el usuario, aplicando un filtro opcional por nombre, y devolviendo su representación.
func (uc *listPlansUseCase) Execute(ctx context.Context, cmd input.ListPlansCommand) ([]input.PlanResult, error) {
	plans, err := uc.repo.FindAll(ctx, cmd.UserID, cmd.NameFilter)
	if err != nil {
		return nil, fmt.Errorf("listing plans: %w", err)
	}

	results := make([]input.PlanResult, len(plans))
	for i, p := range plans {
		results[i] = toPlanResult(&p)
	}

	return results, nil
}
