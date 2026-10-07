package domain

import "errors"

var (
	// ErrPlanNotFound se retorna cuando no se encuentra un plan con el ID especificado para un usuario.
	ErrPlanNotFound = errors.New("plan not found")
	// ErrRoutineNotFound se retorna cuando no se encuentra una rutina con el ID especificado dentro de un plan.
	ErrRoutineNotFound = errors.New("routine not found")
	// ErrNoActivePlan se retorna cuando un usuario no tiene un plan activo.
	ErrNoActivePlan = errors.New("no active plan")
	// ErrExerciseNotFound se retorna cuando no se encuentra un ejercicio con el ID especificado dentro de una rutina.
	ErrExerciseNotFound = errors.New("exercise not found")
)
