package domain

import (
	"time"
)

// ── Errors ────────────────────────────────────────────────────────────────────

// ── Exercise ──────────────────────────────────────────────────────────────────

// Exercise represents a single exercise within a routine.
type Exercise struct {
	ID     string
	Name   string
	Sets   int
	Reps   int
	Weight float64 // kg — 0 means not specified
	Notes  string
}

// ── Routine ───────────────────────────────────────────────────────────────────

// Routine represents a workout day within a plan.
type Routine struct {
	ID        string
	Name      string
	Exercises []Exercise
	CreatedAt time.Time
}

// ── Plan ──────────────────────────────────────────────────────────────────────

// Plan is the aggregate root that owns Routines and their Exercises.
type Plan struct {
	ID        string
	UserID    string
	Name      string
	Active    bool
	Routines  []Routine
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Activate sets this plan as active.
func (p *Plan) Activate() { p.Active = true }

// Deactivate sets this plan as inactive.
func (p *Plan) Deactivate() { p.Active = false }

// AddRoutine appends a new routine to the plan.
func (p *Plan) AddRoutine(r Routine) { p.Routines = append(p.Routines, r) }

// UpdateRoutine replaces the name of a routine by ID.
func (p *Plan) UpdateRoutine(routineID, name string) error {
	for i, r := range p.Routines {
		if r.ID == routineID {
			p.Routines[i].Name = name
			return nil
		}
	}
	return ErrRoutineNotFound
}

// RemoveRoutine deletes a routine by ID.
func (p *Plan) RemoveRoutine(routineID string) error {
	for i, r := range p.Routines {
		if r.ID == routineID {
			p.Routines = append(p.Routines[:i], p.Routines[i+1:]...)
			return nil
		}
	}
	return ErrRoutineNotFound
}

// ── Exercise domain methods ───────────────────────────────────────────────────

// AddExercise appends an exercise to a routine.
func (p *Plan) AddExercise(routineID string, e Exercise) error {
	for i, r := range p.Routines {
		if r.ID == routineID {
			p.Routines[i].Exercises = append(p.Routines[i].Exercises, e)
			return nil
		}
	}
	return ErrRoutineNotFound
}

// UpdateExercise updates an exercise within a routine.
func (p *Plan) UpdateExercise(routineID, exerciseID, name string, sets, reps int, weight float64, notes string) error {
	for i, r := range p.Routines {
		if r.ID != routineID {
			continue
		}
		for j, e := range r.Exercises {
			if e.ID == exerciseID {
				p.Routines[i].Exercises[j] = Exercise{
					ID:     exerciseID,
					Name:   name,
					Sets:   sets,
					Reps:   reps,
					Weight: weight,
					Notes:  notes,
				}
				return nil
			}
		}
		return ErrExerciseNotFound
	}
	return ErrRoutineNotFound
}

// RemoveExercise deletes an exercise from a routine.
func (p *Plan) RemoveExercise(routineID, exerciseID string) error {
	for i, r := range p.Routines {
		if r.ID != routineID {
			continue
		}
		for j, e := range r.Exercises {
			if e.ID == exerciseID {
				p.Routines[i].Exercises = append(
					p.Routines[i].Exercises[:j],
					p.Routines[i].Exercises[j+1:]...,
				)
				return nil
			}
		}
		return ErrExerciseNotFound
	}
	return ErrRoutineNotFound
}

// GetRoutineExercises returns exercises for a given routine.
func (p *Plan) GetRoutineExercises(routineID string) ([]Exercise, error) {
	for _, r := range p.Routines {
		if r.ID == routineID {
			return r.Exercises, nil
		}
	}
	return nil, ErrRoutineNotFound
}
