package application

import (
	"context"
	"fmt"

	"github.com/shotokan/gymtracker-api/internal/plan/domain"
	"github.com/shotokan/gymtracker-api/internal/plan/port/input"
	"github.com/shotokan/gymtracker-api/internal/plan/port/output"
)

type activatePlanUseCase struct{ repo output.PlanRepository }

// NewActivatePlan construye el caso de uso para activar un plan, inyectando el repositorio necesario.
func NewActivatePlan(repo output.PlanRepository) input.ActivatePlanUseCase {
	return &activatePlanUseCase{repo: repo}
}

// Execute activa un plan para el usuario, desactivando cualquier otro plan activo previamente.
func (uc *activatePlanUseCase) Execute(ctx context.Context, cmd input.ActivatePlanCommand) error {
	// Deactivate the currently active plan (if any)
	current, err := uc.repo.FindActive(ctx, cmd.UserID)
	if err == nil && current.ID != cmd.PlanID {
		current.Deactivate()
		if err = uc.repo.Update(ctx, current); err != nil {
			return fmt.Errorf("deactivating current plan: %w", err)
		}
	}

	// Activate the target plan
	target, err := uc.repo.FindByID(ctx, cmd.UserID, cmd.PlanID)
	if err != nil {
		return fmt.Errorf("finding plan: %w", err)
	}

	if target.Active {
		return nil // already active, nothing to do
	}

	target.Activate()
	if err = uc.repo.Update(ctx, target); err != nil {
		return fmt.Errorf("activating plan: %w", err)
	}

	return nil
}

// Ensure domain is imported (used by the business logic above)
var _ = domain.Plan{}
