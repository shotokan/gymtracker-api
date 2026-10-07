package application

import (
	"context"
	"fmt"

	"github.com/shotokan/gymtracker-api/internal/plan/port/input"
	"github.com/shotokan/gymtracker-api/internal/plan/port/output"
)

type updateExerciseUseCase struct {
	repo output.PlanRepository
}

// NewUpdateExerciseUseCase creates a new UpdateExerciseUseCase.
func NewUpdateExerciseUseCase(repo output.PlanRepository) input.UpdateExerciseUseCase {
	return &updateExerciseUseCase{repo: repo}
}

// Execute procesa el comando para actualizar un ejercicio de una rutina dentro de un plan, validando la existencia del plan, la rutina y el ejercicio, y retornando el ejercicio actualizado o un error adecuado.
func (uc *updateExerciseUseCase) Execute(ctx context.Context, cmd input.UpdateExerciseCommand) (input.ExerciseResult, error) {
	plan, err := uc.repo.FindByID(ctx, cmd.UserID, cmd.PlanID)
	if err != nil {
		return input.ExerciseResult{}, fmt.Errorf("updateExercise: find plan: %w", err)
	}

	if err := plan.UpdateExercise(cmd.RoutineID, cmd.ExerciseID, cmd.Name, cmd.Sets, cmd.Reps, cmd.Weight, cmd.Notes); err != nil {
		return input.ExerciseResult{}, fmt.Errorf("updateExercise: %w", err)
	}

	if err := uc.repo.Update(ctx, plan); err != nil {
		return input.ExerciseResult{}, fmt.Errorf("updateExercise: update plan: %w", err)
	}

	// Return the updated exercise from the plan
	exercises, _ := plan.GetRoutineExercises(cmd.RoutineID)
	for _, e := range exercises {
		if e.ID == cmd.ExerciseID {
			return toExerciseResult(e), nil
		}
	}

	return input.ExerciseResult{}, fmt.Errorf("updateExercise: exercise not found after update")
}
