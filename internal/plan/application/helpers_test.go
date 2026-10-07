package application

import (
	"testing"
	"time"

	"github.com/shotokan/gymtracker-api/internal/plan/domain"
	"github.com/stretchr/testify/assert"
)

func TestToPlanResult(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	plan := &domain.Plan{
		ID:     "plan-1",
		UserID: "user-1",
		Name:   "Mi Plan",
		Active: true,
		Routines: []domain.Routine{
			{
				ID:   "routine-1",
				Name: "Push Day",
				Exercises: []domain.Exercise{
					{ID: "ex-1", Name: "Press Banca", Sets: 3, Reps: 10, Weight: 60.0, Notes: "Calentar bien"},
					{ID: "ex-2", Name: "Press Inclinado", Sets: 3, Reps: 12, Weight: 50.0, Notes: ""},
				},
				CreatedAt: createdAt,
			},
		},
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}

	result := toPlanResult(plan)

	assert.Equal(t, "plan-1", result.ID)
	assert.Equal(t, "Mi Plan", result.Name)
	assert.True(t, result.Active)
	assert.Equal(t, "2024-01-15T10:30:00Z", result.CreatedAt)
	assert.Len(t, result.Routines, 1)
	assert.Equal(t, "Push Day", result.Routines[0].Name)
	assert.Len(t, result.Routines[0].Exercises, 2)
	assert.Equal(t, "Press Banca", result.Routines[0].Exercises[0].Name)
	assert.Equal(t, 3, result.Routines[0].Exercises[0].Sets)
	assert.Equal(t, 10, result.Routines[0].Exercises[0].Reps)
	assert.Equal(t, 60.0, result.Routines[0].Exercises[0].Weight)
	assert.Equal(t, "Calentar bien", result.Routines[0].Exercises[0].Notes)
}

func TestToPlanResult_EmptyRoutines(t *testing.T) {
	t.Parallel()

	plan := &domain.Plan{
		ID:        "plan-1",
		UserID:    "user-1",
		Name:      "Plan Vacío",
		Active:    false,
		Routines:  []domain.Routine{},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	result := toPlanResult(plan)

	assert.Equal(t, "plan-1", result.ID)
	assert.Equal(t, "Plan Vacío", result.Name)
	assert.False(t, result.Active)
	assert.Empty(t, result.Routines)
}
