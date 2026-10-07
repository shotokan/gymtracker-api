package application

import (
	"context"
	"fmt"
	"time"

	"github.com/shotokan/gymtracker-api/internal/plan/domain"
	"github.com/shotokan/gymtracker-api/internal/plan/port/input"
	"github.com/shotokan/gymtracker-api/internal/plan/port/output"
)

type createPlanUseCase struct {
	repo  output.PlanRepository
	idGen output.IDGenerator
}

// NewCreatePlan construye el caso de uso para crear un plan, inyectando el repositorio y generador de IDs necesarios.
func NewCreatePlan(repo output.PlanRepository, idGen output.IDGenerator) input.CreatePlanUseCase {
	return &createPlanUseCase{repo: repo, idGen: idGen}
}

// Execute crea un nuevo plan para el usuario, validando los datos de entrada y guardando el plan en el repositorio.
func (uc *createPlanUseCase) Execute(ctx context.Context, cmd input.CreatePlanCommand) (input.PlanResult, error) {
	now := time.Now().UTC()
	plan := &domain.Plan{
		ID:        uc.idGen.New(),
		UserID:    cmd.UserID,
		Name:      cmd.Name,
		Active:    false,
		Routines:  []domain.Routine{},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := uc.repo.Save(ctx, plan); err != nil {
		return input.PlanResult{}, fmt.Errorf("saving plan: %w", err)
	}

	return toPlanResult(plan), nil
}
