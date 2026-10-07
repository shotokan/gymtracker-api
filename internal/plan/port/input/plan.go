package input

import "context"

// ── Shared result types ───────────────────────────────────────────────────────

// RoutineResult representa la información de una rutina dentro de un plan, incluyendo su ID y nombre.
type RoutineResult struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	Exercises []ExerciseResult `json:"exercises"`
}

// PlanResult representa la información de un plan, incluyendo su ID, nombre, estado activo, fecha de creación y las rutinas asociadas.
type PlanResult struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Active    bool            `json:"active"`
	CreatedAt string          `json:"created_at"`
	Routines  []RoutineResult `json:"routines,omitempty"`
}

// ── Create ────────────────────────────────────────────────────────────────────

// CreatePlanCommand representa la solicitud de creación de un nuevo plan, conteniendo el ID del usuario y el nombre del plan a crear.
type CreatePlanCommand struct {
	UserID string
	Name   string
}

// CreatePlanUseCase es el puerto primario del dominio plan para la creación de un nuevo plan.
type CreatePlanUseCase interface {
	Execute(ctx context.Context, cmd CreatePlanCommand) (PlanResult, error)
}

// ── List ──────────────────────────────────────────────────────────────────────

// ListPlansCommand representa la solicitud de listado de planes, conteniendo el ID del usuario y un filtro opcional por nombre.
type ListPlansCommand struct {
	UserID     string
	NameFilter string // optional
}

// ListPlansUseCase es el puerto primario del dominio plan para listar los planes de un usuario, aplicando un filtro opcional por nombre.
type ListPlansUseCase interface {
	Execute(ctx context.Context, cmd ListPlansCommand) ([]PlanResult, error)
}

// ── Get ───────────────────────────────────────────────────────────────────────

// GetPlanCommand representa la solicitud de obtención de un plan específico, conteniendo el ID del usuario y el ID del plan a obtener.
type GetPlanCommand struct {
	UserID string
	PlanID string
}

// GetPlanUseCase es el puerto primario del dominio plan para obtener la información de un plan específico de un usuario.
type GetPlanUseCase interface {
	Execute(ctx context.Context, cmd GetPlanCommand) (PlanResult, error)
}

// ── Get active ────────────────────────────────────────────────────────────────

// GetActivePlanCommand representa la solicitud de obtención del plan activo de un usuario, conteniendo el ID del usuario.
type GetActivePlanCommand struct {
	UserID string
}

// GetActivePlanUseCase es el puerto primario del dominio plan para obtener el plan activo de un usuario.
type GetActivePlanUseCase interface {
	Execute(ctx context.Context, cmd GetActivePlanCommand) (PlanResult, error)
}

// ── Update ────────────────────────────────────────────────────────────────────

// UpdatePlanCommand representa la solicitud de actualización de un plan, conteniendo el ID del usuario, el ID del plan a actualizar y el nuevo nombre del plan.
type UpdatePlanCommand struct {
	UserID string
	PlanID string
	Name   string
}

// UpdatePlanUseCase es el puerto primario del dominio plan para actualizar la información de un plan específico de un usuario.
type UpdatePlanUseCase interface {
	Execute(ctx context.Context, cmd UpdatePlanCommand) (PlanResult, error)
}

// ── Activate ──────────────────────────────────────────────────────────────────

// ActivatePlanCommand representa la solicitud de activación de un plan, conteniendo el ID del usuario y el ID del plan a activar.
type ActivatePlanCommand struct {
	UserID string
	PlanID string
}

// ActivatePlanUseCase es el puerto primario del dominio plan para activar un plan específico de un usuario, desactivando cualquier otro plan activo previamente.
type ActivatePlanUseCase interface {
	Execute(ctx context.Context, cmd ActivatePlanCommand) error
}

// ── Delete ────────────────────────────────────────────────────────────────────

// DeletePlanCommand representa la solicitud de eliminación de un plan, conteniendo el ID del usuario y el ID del plan a eliminar.
type DeletePlanCommand struct {
	UserID string
	PlanID string
}

// DeletePlanUseCase es el puerto primario del dominio plan para eliminar un plan específico de un usuario, validando la existencia del plan y eliminándolo del repositorio.
type DeletePlanUseCase interface {
	Execute(ctx context.Context, cmd DeletePlanCommand) error
}
