package application

import (
	"github.com/shotokan/gymtracker-api/internal/plan/domain"
	"github.com/shotokan/gymtracker-api/internal/plan/port/input"
)

// toExerciseResult maps a domain Exercise to the API result type.
func toExerciseResult(e domain.Exercise) input.ExerciseResult {
	return input.ExerciseResult{
		ID:     e.ID,
		Name:   e.Name,
		Sets:   e.Sets,
		Reps:   e.Reps,
		Weight: e.Weight,
		Notes:  e.Notes,
	}
}

// toRoutineResult maps a domain Routine to the API result type.
func toRoutineResult(r domain.Routine) input.RoutineResult {
	exercises := make([]input.ExerciseResult, 0, len(r.Exercises))
	for _, e := range r.Exercises {
		exercises = append(exercises, toExerciseResult(e))
	}
	return input.RoutineResult{
		ID:        r.ID,
		Name:      r.Name,
		Exercises: exercises,
	}
}

// toPlanResult maps a domain Plan to the API result type.
func toPlanResult(p *domain.Plan) input.PlanResult {
	routines := make([]input.RoutineResult, 0, len(p.Routines))
	for _, r := range p.Routines {
		routines = append(routines, toRoutineResult(r))
	}
	return input.PlanResult{
		ID:        p.ID,
		Name:      p.Name,
		Active:    p.Active,
		CreatedAt: p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Routines:  routines,
	}
}
