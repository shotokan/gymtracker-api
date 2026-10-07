package input

import "context"

// ── Add routine ───────────────────────────────────────────────────────────────

// AddRoutineCommand representa la solicitud de creación de una rutina, conteniendo el ID del usuario, el ID del plan al que se agregará la rutina y el nombre de la nueva rutina.
type AddRoutineCommand struct {
	UserID string
	PlanID string
	Name   string
}

// AddRoutineUseCase es el puerto primario del dominio plan para agregar una nueva rutina a un plan específico de un usuario, validando la existencia del plan y agregando la rutina al repositorio.
type AddRoutineUseCase interface {
	Execute(ctx context.Context, cmd AddRoutineCommand) (RoutineResult, error)
}

// ── Update routine ────────────────────────────────────────────────────────────

// UpdateRoutineCommand representa la solicitud de actualización de una rutina, conteniendo el ID del usuario, el ID del plan al que pertenece la rutina, el ID de la rutina a actualizar y el nuevo nombre de la rutina.
type UpdateRoutineCommand struct {
	UserID    string
	PlanID    string
	RoutineID string
	Name      string
}

// UpdateRoutineUseCase es el puerto primario del dominio plan para actualizar la información de una rutina específica de un plan de un usuario, validando la existencia de la rutina dentro del plan y actualizando su información en el repositorio.
type UpdateRoutineUseCase interface {
	Execute(ctx context.Context, cmd UpdateRoutineCommand) (RoutineResult, error)
}

// ── Delete routine ────────────────────────────────────────────────────────────

// DeleteRoutineCommand representa la solicitud de eliminación de una rutina, conteniendo el ID del usuario, el ID del plan al que pertenece la rutina y el ID de la rutina a eliminar.
type DeleteRoutineCommand struct {
	UserID    string
	PlanID    string
	RoutineID string
}

// DeleteRoutineUseCase es el puerto primario del dominio plan para eliminar una rutina específica de un plan de un usuario, validando la existencia de la rutina dentro del plan y eliminándola del repositorio.
type DeleteRoutineUseCase interface {
	Execute(ctx context.Context, cmd DeleteRoutineCommand) error
}
