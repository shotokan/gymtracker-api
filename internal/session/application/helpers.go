package application

import (
	"github.com/shotokan/gymtracker-api/internal/session/domain"
	"github.com/shotokan/gymtracker-api/internal/session/port/input"
)

// toCompletedSetResult maps a domain CompletedSet to the API result type.
func toCompletedSetResult(s domain.CompletedSet) input.CompletedSetResult {
	return input.CompletedSetResult{
		Weight:      s.Weight,
		Reps:        s.Reps,
		RestSeconds: s.RestSeconds,
	}
}

// toSessionExerciseResult maps a domain SessionExercise to the API result type.
func toSessionExerciseResult(e domain.SessionExercise) input.SessionExerciseResult {
	sets := make([]input.CompletedSetResult, 0, len(e.Sets))
	for _, s := range e.Sets {
		sets = append(sets, toCompletedSetResult(s))
	}
	return input.SessionExerciseResult{
		ExerciseID: e.ExerciseID,
		Name:       e.Name,
		Sets:       sets,
	}
}

// toSessionResult maps a domain Session to the API result type.
func toSessionResult(s *domain.Session) input.SessionResult {
	exercises := make([]input.SessionExerciseResult, 0, len(s.Exercises))
	for _, e := range s.Exercises {
		exercises = append(exercises, toSessionExerciseResult(e))
	}
	return input.SessionResult{
		ID:               s.ID,
		PlanID:           s.PlanID,
		RoutineID:        s.RoutineID,
		RoutineName:      s.RoutineName,
		Date:             s.Date,
		DurationSecs:     s.DurationSecs,
		Exercises:        exercises,
		TotalSets:        s.TotalSets(),
		TotalVolume:      s.TotalVolume(),
		TotalRestSeconds: s.TotalRestSeconds(),
		StartedAt:        s.StartedAt.Format("2006-01-02T15:04:05Z07:00"),
		FinishedAt:       s.FinishedAt.Format("2006-01-02T15:04:05Z07:00"),
		CreatedAt:        s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
