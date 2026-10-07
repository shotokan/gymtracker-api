package application

import (
	"context"
	"fmt"

	"github.com/shotokan/gymtracker-api/internal/plan/port/input"
	"github.com/shotokan/gymtracker-api/internal/plan/port/output"
)

type deleteExerciseUseCase struct {
	repo output.PlanRepository
}

// NewDeleteExerciseUseCase creates a new DeleteExerciseUseCase.
func NewDeleteExerciseUseCase(repo output.PlanRepository) input.DeleteExerciseUseCase {
	return &deleteExerciseUseCase{repo: repo}
}

// Execute processes the command to delete an exercise from a routine within a plan, validating the existence of the plan, the routine and the exercise, and returning an appropriate error if any of them does not exist.
func (uc *deleteExerciseUseCase) Execute(ctx context.Context, cmd input.DeleteExerciseCommand) error {
	plan, err := uc.repo.FindByID(ctx, cmd.UserID, cmd.PlanID)
	if err != nil {
		return fmt.Errorf("deleteExercise: find plan: %w", err)
	}

	if err := plan.RemoveExercise(cmd.RoutineID, cmd.ExerciseID); err != nil {
		return fmt.Errorf("deleteExercise: %w", err)
	}

	if err := uc.repo.Update(ctx, plan); err != nil {
		return fmt.Errorf("deleteExercise: update plan: %w", err)
	}

	return nil
}
