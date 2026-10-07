package application

import (
	"context"
	"fmt"

	"github.com/shotokan/gymtracker-api/internal/plan/domain"
	"github.com/shotokan/gymtracker-api/internal/plan/port/input"
	"github.com/shotokan/gymtracker-api/internal/plan/port/output"
)

type addRoutineUseCase struct {
	repo  output.PlanRepository
	idGen output.IDGenerator
}

// NewAddRoutine construye el caso de uso para agregar una rutina a un plan, inyectando el repositorio y generador de IDs necesarios.
func NewAddRoutine(repo output.PlanRepository, idGen output.IDGenerator) input.AddRoutineUseCase {
	return &addRoutineUseCase{repo: repo, idGen: idGen}
}

// Execute agrega una nueva rutina al plan especificado, validando la existencia del plan y actualizando el repositorio.
func (uc *addRoutineUseCase) Execute(ctx context.Context, cmd input.AddRoutineCommand) (input.RoutineResult, error) {
	plan, err := uc.repo.FindByID(ctx, cmd.UserID, cmd.PlanID)
	if err != nil {
		return input.RoutineResult{}, fmt.Errorf("finding plan: %w", err)
	}

	routine := domain.Routine{
		ID:        uc.idGen.New(),
		Name:      cmd.Name,
		Exercises: []domain.Exercise{},
	}

	plan.AddRoutine(routine)

	if err = uc.repo.Update(ctx, plan); err != nil {
		return input.RoutineResult{}, fmt.Errorf("updating plan: %w", err)
	}

	return input.RoutineResult{ID: routine.ID, Name: routine.Name}, nil
}
