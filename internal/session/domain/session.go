package domain

import "time"

// CompletedSet represents a single completed set within an exercise.
type CompletedSet struct {
	Weight      float64 // kg — 0 is valid (bodyweight)
	Reps        int
	RestSeconds int // actual seconds rested after this set
	// 0 on the last set of each exercise
}

// SessionExercise represents a single exercise completed during a session.
type SessionExercise struct {
	ExerciseID string
	Name       string
	Sets       []CompletedSet
}

// Session is the aggregate root representing a completed workout session.
type Session struct {
	ID           string
	UserID       string
	PlanID       string // empty string if user had no active plan
	RoutineID    string
	RoutineName  string
	Date         string // YYYY-MM-DD, user local date
	DurationSecs int    // total wall-clock seconds from first set to last set
	Exercises    []SessionExercise
	StartedAt    time.Time // when the session started
	FinishedAt   time.Time // when the session finished
	CreatedAt    time.Time
}

// TotalSets returns the sum of all completed sets across all exercises.
func (s *Session) TotalSets() int {
	total := 0
	for _, ex := range s.Exercises {
		total += len(ex.Sets)
	}
	return total
}

// TotalVolume returns total kg lifted: sum of (weight × reps) across all sets.
func (s *Session) TotalVolume() float64 {
	var total float64
	for _, ex := range s.Exercises {
		for _, set := range ex.Sets {
			total += set.Weight * float64(set.Reps)
		}
	}
	return total
}

// TotalRestSeconds returns the sum of all rest seconds across all sets.
func (s *Session) TotalRestSeconds() int {
	total := 0
	for _, ex := range s.Exercises {
		for _, set := range ex.Sets {
			total += set.RestSeconds
		}
	}
	return total
}
