package application

import (
	"context"
	"testing"
	"time"

	"github.com/shotokan/gymtracker-api/internal/session/domain"
	"github.com/shotokan/gymtracker-api/internal/session/port/input"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetSessionUseCase_Execute_Success(t *testing.T) {
	t.Parallel()

	repo := new(mockSessionRepo)

	session := &domain.Session{
		ID:           "session-123",
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
		},
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	repo.On("FindByID", mock.Anything, "user-1", "session-123").Return(session, nil)

	uc := NewGetSession(repo)

	result, err := uc.Execute(context.Background(), input.GetSessionCommand{
		UserID:    "user-1",
		SessionID: "session-123",
	})

	require.NoError(t, err)
	assert.Equal(t, "session-123", result.ID)
	assert.Equal(t, "Push Day", result.RoutineName)
	assert.Equal(t, "2024-01-15", result.Date)
	assert.Equal(t, 2, result.TotalSets)
	assert.Equal(t, 1120.0, result.TotalVolume) // 60*10 + 65*8
	assert.Len(t, result.Exercises, 1)
	repo.AssertExpectations(t)
}

func TestGetSessionUseCase_Execute_NotFound(t *testing.T) {
	t.Parallel()

	repo := new(mockSessionRepo)
	repo.On("FindByID", mock.Anything, "user-1", "session-999").Return(nil, domain.ErrSessionNotFound)

	uc := NewGetSession(repo)

	_, err := uc.Execute(context.Background(), input.GetSessionCommand{
		UserID:    "user-1",
		SessionID: "session-999",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "finding session")
	assert.ErrorIs(t, err, domain.ErrSessionNotFound)
	repo.AssertExpectations(t)
}
