package input

import "context"

// ExerciseResult representa la información de un ejercicio dentro de una rutina, incluyendo su ID, nombre, número de series, repeticiones, peso y notas adicionales.
type ExerciseResult struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Sets   int     `json:"sets"`
	Reps   int     `json:"reps"`
	Weight float64 `json:"weight"`
	Notes  string  `json:"notes"`
}

// AddExerciseCommand representa la solicitud de creación de un ejercicio, conteniendo el ID del usuario, el ID del plan al que pertenece la rutina, el ID de la rutina a la que se agregará el ejercicio y los detalles del nuevo ejercicio.
type AddExerciseCommand struct {
	UserID    string
	PlanID    string
	RoutineID string
	Name      string
	Sets      int
	Reps      int
	Weight    float64
	Notes     string
}

// AddExerciseUseCase es el puerto primario del dominio plan para agregar un nuevo ejercicio a una rutina específica de un plan de un usuario, validando la existencia del plan y la rutina, creando el ejercicio con un ID único y agregándolo al repositorio.
type AddExerciseUseCase interface {
	Execute(ctx context.Context, cmd AddExerciseCommand) (ExerciseResult, error)
}

// ListExercisesCommand representa la solicitud de listado de ejercicios, conteniendo el ID del usuario, el ID del plan al que pertenece la rutina y el ID de la rutina de la que se listarán los ejercicios.
type ListExercisesCommand struct {
	UserID    string
	PlanID    string
	RoutineID string
}

// ListExercisesUseCase es el puerto primario del dominio plan para listar los ejercicios de una rutina específica de un plan de un usuario, validando la existencia del plan y la rutina, y retornando la lista de ejercicios o un error adecuado.
type ListExercisesUseCase interface {
	Execute(ctx context.Context, cmd ListExercisesCommand) ([]ExerciseResult, error)
}

// UpdateExerciseCommand representa la solicitud de actualización de un ejercicio, conteniendo el ID del usuario, el ID del plan al que pertenece la rutina, el ID de la rutina a la que pertenece el ejercicio, el ID del ejercicio a actualizar y los nuevos detalles del ejercicio.
type UpdateExerciseCommand struct {
	UserID     string
	PlanID     string
	RoutineID  string
	ExerciseID string
	Name       string
	Sets       int
	Reps       int
	Weight     float64
	Notes      string
}

// UpdateExerciseUseCase es el puerto primario del dominio plan para actualizar la información de un ejercicio específico de una rutina de un plan de un usuario, validando la existencia del plan, la rutina y el ejercicio, y actualizando su información en el repositorio.
type UpdateExerciseUseCase interface {
	Execute(ctx context.Context, cmd UpdateExerciseCommand) (ExerciseResult, error)
}

// DeleteExerciseCommand representa la solicitud de eliminación de un ejercicio, conteniendo el ID del usuario, el ID del plan al que pertenece la rutina, el ID de la rutina a la que pertenece el ejercicio y el ID del ejercicio a eliminar.
type DeleteExerciseCommand struct {
	UserID     string
	PlanID     string
	RoutineID  string
	ExerciseID string
}

// DeleteExerciseUseCase es el puerto primario del dominio plan para eliminar un ejercicio específico de una rutina de un plan de un usuario, validando la existencia del plan, la rutina y el ejercicio, y eliminándolo del repositorio.
type DeleteExerciseUseCase interface {
	Execute(ctx context.Context, cmd DeleteExerciseCommand) error
}
