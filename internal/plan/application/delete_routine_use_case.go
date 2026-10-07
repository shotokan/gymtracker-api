package application

import (
	"context"
	"fmt"

	"github.com/shotokan/gymtracker-api/internal/plan/port/input"
	"github.com/shotokan/gymtracker-api/internal/plan/port/output"
)

type deleteRoutineUseCase struct{ repo output.PlanRepository }

// NewDeleteRoutine construye el caso de uso para eliminar una rutina de un plan, inyectando el repositorio necesario.
func NewDeleteRoutine(repo output.PlanRepository) input.DeleteRoutineUseCase {
	return &deleteRoutineUseCase{repo: repo}
}

// Execute elimina una rutina del plan especificado, validando la existencia del plan y la rutina, y actualizando el repositorio.
func (uc *deleteRoutineUseCase) Execute(ctx context.Context, cmd input.DeleteRoutineCommand) error {
	plan, err := uc.repo.FindByID(ctx, cmd.UserID, cmd.PlanID)
	if err != nil {
		return fmt.Errorf("finding plan: %w", err)
	}

	if err := plan.RemoveRoutine(cmd.RoutineID); err != nil {
		return fmt.Errorf("removing routine: %w", err)
	}

	if err = uc.repo.Update(ctx, plan); err != nil {
		return fmt.Errorf("updating plan: %w", err)
	}

	return nil
}
