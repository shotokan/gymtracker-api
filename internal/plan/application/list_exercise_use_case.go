package application

import (
	"context"
	"fmt"

	"github.com/shotokan/gymtracker-api/internal/plan/port/input"
	"github.com/shotokan/gymtracker-api/internal/plan/port/output"
)

type listExercisesUseCase struct {
	repo output.PlanRepository
}

// NewListExercisesUseCase creates a new ListExercisesUseCase.
func NewListExercisesUseCase(repo output.PlanRepository) input.ListExercisesUseCase {
	return &listExercisesUseCase{repo: repo}
}

// Execute procesa el comando para listar los ejercicios de una rutina dentro de un plan, validando la existencia del plan y la rutina, y retornando la lista de ejercicios o un error adecuado.
func (uc *listExercisesUseCase) Execute(ctx context.Context, cmd input.ListExercisesCommand) ([]input.ExerciseResult, error) {
	plan, err := uc.repo.FindByID(ctx, cmd.UserID, cmd.PlanID)
	if err != nil {
		return nil, fmt.Errorf("listExercises: find plan: %w", err)
	}

	exercises, err := plan.GetRoutineExercises(cmd.RoutineID)
	if err != nil {
		return nil, fmt.Errorf("listExercises: %w", err)
	}

	results := make([]input.ExerciseResult, 0, len(exercises))
	for _, e := range exercises {
		results = append(results, toExerciseResult(e))
	}

	return results, nil
}
