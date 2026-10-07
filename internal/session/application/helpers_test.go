package application

import (
	"testing"
	"time"

	"github.com/shotokan/gymtracker-api/internal/session/domain"
	"github.com/stretchr/testify/assert"
)

func TestToSessionResult(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	finishedAt := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	session := &domain.Session{
		ID:           "session-1",
		UserID:       "user-1",
		PlanID:       "plan-1",
		RoutineID:    "routine-1",
		RoutineName:  "Push Day",
		Date:         "2024-01-15",
		DurationSecs: 3600,
		Exercises: []domain.SessionExercise{
			{
				ExerciseID: "ex-1",
				Name:       "Press Banca",
				Sets: []domain.CompletedSet{
					{Weight: 60.0, Reps: 10, RestSeconds: 90},
					{Weight: 65.0, Reps: 8, RestSeconds: 0},
				},
			},
			{
				ExerciseID: "ex-2",
				Name:       "Sentadilla",
				Sets: []domain.CompletedSet{
					{Weight: 100.0, Reps: 5, RestSeconds: 120},
				},
			},
		},
		StartedAt:  startedAt,
		FinishedAt: finishedAt,
		CreatedAt:  createdAt,
	}

	result := toSessionResult(session)

	assert.Equal(t, "session-1", result.ID)
	assert.Equal(t, "plan-1", result.PlanID)
	assert.Equal(t, "routine-1", result.RoutineID)
	assert.Equal(t, "Push Day", result.RoutineName)
	assert.Equal(t, "2024-01-15", result.Date)
	assert.Equal(t, 3600, result.DurationSecs)
	assert.Equal(t, "2024-01-15T09:00:00Z", result.StartedAt)
	assert.Equal(t, "2024-01-15T10:00:00Z", result.FinishedAt)
	assert.Equal(t, "2024-01-15T10:30:00Z", result.CreatedAt)
	assert.Equal(t, 3, result.TotalSets)
	assert.InDelta(t, 1620.0, result.TotalVolume, 0.01) // (60*10 + 65*8 + 100*5)
	assert.Equal(t, 210, result.TotalRestSeconds)       // 90 + 0 + 120

	assert.Len(t, result.Exercises, 2)
	assert.Equal(t, "Press Banca", result.Exercises[0].Name)
	assert.Len(t, result.Exercises[0].Sets, 2)
	assert.Equal(t, 60.0, result.Exercises[0].Sets[0].Weight)
	assert.Equal(t, 10, result.Exercises[0].Sets[0].Reps)
	assert.Equal(t, 90, result.Exercises[0].Sets[0].RestSeconds)
}

func TestToSessionResult_EmptyPlanID(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	session := &domain.Session{
		ID:           "session-1",
		UserID:       "user-1",
		PlanID:       "", // no active plan
		RoutineID:    "routine-1",
		RoutineName:  "Custom Workout",
		Date:         "2024-01-15",
		DurationSecs: 1800,
		Exercises:    []domain.SessionExercise{},
		StartedAt:    now,
		FinishedAt:   now,
		CreatedAt:    now,
	}

	result := toSessionResult(session)

	assert.Equal(t, "", result.PlanID)
	assert.Equal(t, 0, result.TotalSets)
	assert.Equal(t, 0.0, result.TotalVolume)
	assert.Equal(t, 0, result.TotalRestSeconds)
}
