package application

import (
	"context"
	"fmt"

	"github.com/shotokan/gymtracker-api/internal/plan/port/input"
	"github.com/shotokan/gymtracker-api/internal/plan/port/output"
)

type updateRoutineUseCase struct{ repo output.PlanRepository }

// NewUpdateRoutine construye el caso de uso para actualizar una rutina de un plan, inyectando el repositorio necesario.
func NewUpdateRoutine(repo output.PlanRepository) input.UpdateRoutineUseCase {
	return &updateRoutineUseCase{repo: repo}
}

// Execute actualiza el nombre de una rutina del plan especificado, validando la existencia del plan y la rutina, y guardando los cambios en el repositorio.
func (uc *updateRoutineUseCase) Execute(ctx context.Context, cmd input.UpdateRoutineCommand) (input.RoutineResult, error) {
	plan, err := uc.repo.FindByID(ctx, cmd.UserID, cmd.PlanID)
	if err != nil {
		return input.RoutineResult{}, fmt.Errorf("finding plan: %w", err)
	}

	if err := plan.UpdateRoutine(cmd.RoutineID, cmd.Name); err != nil {
		return input.RoutineResult{}, fmt.Errorf("updating routine: %w", err)
	}

	if err = uc.repo.Update(ctx, plan); err != nil {
		return input.RoutineResult{}, fmt.Errorf("updating plan: %w", err)
	}

	return input.RoutineResult{ID: cmd.RoutineID, Name: cmd.Name}, nil
}
