package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shotokan/gymtracker-api/internal/session/domain"
	"github.com/shotokan/gymtracker-api/internal/session/port/input"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListSessionsUseCase_Execute_Success(t *testing.T) {
	t.Parallel()

	repo := new(mockSessionRepo)

	sessions := []domain.Session{
		{
			ID:           "session-1",
			UserID:       "user-1",
			RoutineID:    "routine-1",
			RoutineName:  "Push Day",
			Date:         "2024-01-15",
			DurationSecs: 3600,
			Exercises: []domain.SessionExercise{
				{ExerciseID: "ex-1", Name: "Press", Sets: []domain.CompletedSet{{Weight: 60, Reps: 10, RestSeconds: 0}}},
			},
			CreatedAt: time.Now().UTC(),
		},
		{
			ID:           "session-2",
			UserID:       "user-1",
			RoutineID:    "routine-2",
			RoutineName:  "Pull Day",
			Date:         "2024-01-14",
			DurationSecs: 3000,
			Exercises: []domain.SessionExercise{
				{ExerciseID: "ex-2", Name: "Deadlift", Sets: []domain.CompletedSet{{Weight: 100, Reps: 5, RestSeconds: 0}}},
			},
			CreatedAt: time.Now().UTC(),
		},
	}

	repo.On("FindAll", mock.Anything, "user-1", 20).Return(sessions, nil)

	uc := NewListSessions(repo)

	result, err := uc.Execute(context.Background(), input.ListSessionsCommand{
		UserID: "user-1",
		Limit:  20,
	})

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "session-1", result[0].ID)
	assert.Equal(t, "Push Day", result[0].RoutineName)
	assert.Equal(t, "session-2", result[1].ID)
	assert.Equal(t, "Pull Day", result[1].RoutineName)
	repo.AssertExpectations(t)
}

func TestListSessionsUseCase_Execute_EmptyList(t *testing.T) {
	t.Parallel()

	repo := new(mockSessionRepo)
	repo.On("FindAll", mock.Anything, "user-1", 20).Return([]domain.Session{}, nil)

	uc := NewListSessions(repo)

	result, err := uc.Execute(context.Background(), input.ListSessionsCommand{
		UserID: "user-1",
		Limit:  20,
	})

	require.NoError(t, err)
	assert.Empty(t, result)
	repo.AssertExpectations(t)
}

func TestListSessionsUseCase_Execute_LimitClamping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		inputLimit    int
		expectedLimit int
	}{
		{"limit 0 → clamped to 20", 0, 20},
		{"limit -5 → clamped to 20", -5, 20},
		{"limit 150 → clamped to 20", 150, 20},
		{"limit 50 → kept as 50", 50, 50},
		{"limit 100 → kept as 100", 100, 100},
		{"limit 10 → kept as 10", 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := new(mockSessionRepo)
			repo.On("FindAll", mock.Anything, "user-1", tt.expectedLimit).Return([]domain.Session{}, nil)

			uc := NewListSessions(repo)

			_, err := uc.Execute(context.Background(), input.ListSessionsCommand{
				UserID: "user-1",
				Limit:  tt.inputLimit,
			})

			require.NoError(t, err)
			repo.AssertExpectations(t)
		})
	}
}

func TestListSessionsUseCase_Execute_RepositoryError(t *testing.T) {
	t.Parallel()

	repo := new(mockSessionRepo)
	repo.On("FindAll", mock.Anything, "user-1", 20).Return(nil, errors.New("dynamo error"))

	uc := NewListSessions(repo)

	_, err := uc.Execute(context.Background(), input.ListSessionsCommand{
		UserID: "user-1",
		Limit:  20,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "listing sessions")
	repo.AssertExpectations(t)
}
