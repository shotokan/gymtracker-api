package application

import (
	"context"
	"fmt"

	"github.com/shotokan/gymtracker-api/internal/plan/domain"
	"github.com/shotokan/gymtracker-api/internal/plan/port/input"
	"github.com/shotokan/gymtracker-api/internal/plan/port/output"
)

type addExerciseUseCase struct {
	repo  output.PlanRepository
	idGen output.IDGenerator
}

// NewAddExerciseUseCase construye el caso de uso para agregar un ejercicio a una rutina dentro de un plan, inyectando las dependencias necesarias.
func NewAddExerciseUseCase(repo output.PlanRepository, idGen output.IDGenerator) input.AddExerciseUseCase {
	return &addExerciseUseCase{repo: repo, idGen: idGen}
}

// Execute procesa el comando para agregar un ejercicio a una rutina dentro de un plan, validando la existencia del plan y la rutina, y retornando el ejercicio creado o un error adecuado.
func (uc *addExerciseUseCase) Execute(ctx context.Context, cmd input.AddExerciseCommand) (input.ExerciseResult, error) {
	plan, err := uc.repo.FindByID(ctx, cmd.UserID, cmd.PlanID)
	if err != nil {
		return input.ExerciseResult{}, fmt.Errorf("addExercise: find plan: %w", err)
	}

	exercise := domain.Exercise{
		ID:     uc.idGen.New(),
		Name:   cmd.Name,
		Sets:   cmd.Sets,
		Reps:   cmd.Reps,
		Weight: cmd.Weight,
		Notes:  cmd.Notes,
	}

	if err := plan.AddExercise(cmd.RoutineID, exercise); err != nil {
		return input.ExerciseResult{}, fmt.Errorf("addExercise: %w", err)
	}

	if err := uc.repo.Update(ctx, plan); err != nil {
		return input.ExerciseResult{}, fmt.Errorf("addExercise: update plan: %w", err)
	}

	return toExerciseResult(exercise), nil
}
