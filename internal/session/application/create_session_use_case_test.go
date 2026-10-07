package application

import (
	"context"
	"errors"
	"testing"

	"github.com/shotokan/gymtracker-api/internal/session/domain"
	"github.com/shotokan/gymtracker-api/internal/session/port/input"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ── Mocks ─────────────────────────────────────────────────────────────────

type mockSessionRepo struct{ mock.Mock }

func (m *mockSessionRepo) Save(ctx context.Context, s *domain.Session) error {
	return m.Called(ctx, s).Error(0)
}

func (m *mockSessionRepo) FindByID(ctx context.Context, userID, sessionID string) (*domain.Session, error) {
	args := m.Called(ctx, userID, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Session), args.Error(1)
}

func (m *mockSessionRepo) FindAll(ctx context.Context, userID string, limit int) ([]domain.Session, error) {
	args := m.Called(ctx, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Session), args.Error(1)
}

type mockIDGen struct{ mock.Mock }

func (m *mockIDGen) New() string {
	return m.Called().String(0)
}

// ── Tests ─────────────────────────────────────────────────────────────────

func TestCreateSessionUseCase_Execute_Success(t *testing.T) {
	t.Parallel()

	repo := new(mockSessionRepo)
	idGen := new(mockIDGen)

	idGen.On("New").Return("session-123")
	repo.On("Save", mock.Anything, mock.MatchedBy(func(s *domain.Session) bool {
		return s.ID == "session-123" &&
			s.UserID == "user-1" &&
			s.RoutineID == "routine-1" &&
			s.RoutineName == "Push Day" &&
			s.Date == "2024-01-15" &&
			s.DurationSecs == 3600 &&
			len(s.Exercises) == 1
	})).Return(nil)

	uc := NewCreateSession(repo, idGen)

	cmd := input.CreateSessionCommand{
		UserID:       "user-1",
		PlanID:       "plan-1",
		RoutineID:    "routine-1",
		RoutineName:  "Push Day",
		Date:         "2024-01-15",
		DurationSecs: 3600, // This will be ignored and calculated from started_at and finished_at
		StartedAt:    "2024-01-15T09:00:00Z",
		FinishedAt:   "2024-01-15T10:00:00Z", // 1 hour later = 3600 seconds
		Exercises: []input.SessionExerciseInput{
			{
				ExerciseID: "ex-1",
				Name:       "Press Banca",
				Sets: []input.CompletedSetInput{
					{Weight: 60.0, Reps: 10, RestSeconds: 90},
				},
			},
		},
	}

	result, err := uc.Execute(context.Background(), cmd)

	require.NoError(t, err)
	assert.Equal(t, "session-123", result.ID)
	assert.Equal(t, "Push Day", result.RoutineName)
	assert.Equal(t, 1, result.TotalSets)
	assert.Equal(t, 600.0, result.TotalVolume)
	repo.AssertExpectations(t)
	idGen.AssertExpectations(t)
}

func TestCreateSessionUseCase_Execute_DefaultDate(t *testing.T) {
	t.Parallel()

	repo := new(mockSessionRepo)
	idGen := new(mockIDGen)

	idGen.On("New").Return("session-123")
	repo.On("Save", mock.Anything, mock.MatchedBy(func(s *domain.Session) bool {
		// Date should be set to today in YYYY-MM-DD format
		return s.Date != "" && len(s.Date) == 10
	})).Return(nil)

	uc := NewCreateSession(repo, idGen)

	cmd := input.CreateSessionCommand{
		UserID:       "user-1",
		RoutineID:    "routine-1",
		RoutineName:  "Workout",
		Date:         "", // empty, should default to today
		DurationSecs: 1800,
		Exercises: []input.SessionExerciseInput{
			{
				ExerciseID: "ex-1",
				Name:       "Squat",
				Sets:       []input.CompletedSetInput{{Weight: 100, Reps: 5, RestSeconds: 120}},
			},
		},
	}

	result, err := uc.Execute(context.Background(), cmd)

	require.NoError(t, err)
	assert.NotEmpty(t, result.Date)
	repo.AssertExpectations(t)
}

func TestCreateSessionUseCase_Execute_ValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cmd     input.CreateSessionCommand
		wantErr string
	}{
		{
			name: "routine_id vacío",
			cmd: input.CreateSessionCommand{
				UserID:       "user-1",
				RoutineID:    "",
				RoutineName:  "Workout",
				DurationSecs: 1800,
				Exercises:    []input.SessionExerciseInput{{ExerciseID: "ex-1", Name: "Ex", Sets: []input.CompletedSetInput{{Weight: 50, Reps: 10, RestSeconds: 0}}}},
			},
			wantErr: "routine_id is required",
		},
		{
			name: "routine_name vacío",
			cmd: input.CreateSessionCommand{
				UserID:       "user-1",
				RoutineID:    "routine-1",
				RoutineName:  "",
				DurationSecs: 1800,
				Exercises:    []input.SessionExerciseInput{{ExerciseID: "ex-1", Name: "Ex", Sets: []input.CompletedSetInput{{Weight: 50, Reps: 10, RestSeconds: 0}}}},
			},
			wantErr: "routine_name is required",
		},
		{
			name: "duration negativa",
			cmd: input.CreateSessionCommand{
				UserID:       "user-1",
				RoutineID:    "routine-1",
				RoutineName:  "Workout",
				DurationSecs: -100,
				Exercises:    []input.SessionExerciseInput{{ExerciseID: "ex-1", Name: "Ex", Sets: []input.CompletedSetInput{{Weight: 50, Reps: 10, RestSeconds: 0}}}},
			},
			wantErr: "duration_secs must be >= 0",
		},
		{
			name: "sin ejercicios",
			cmd: input.CreateSessionCommand{
				UserID:       "user-1",
				RoutineID:    "routine-1",
				RoutineName:  "Workout",
				DurationSecs: 1800,
				Exercises:    []input.SessionExerciseInput{},
			},
			wantErr: "at least one exercise is required",
		},
		{
			name: "ejercicio sin sets",
			cmd: input.CreateSessionCommand{
				UserID:       "user-1",
				RoutineID:    "routine-1",
				RoutineName:  "Workout",
				DurationSecs: 1800,
				Exercises:    []input.SessionExerciseInput{{ExerciseID: "ex-1", Name: "Ex", Sets: []input.CompletedSetInput{}}},
			},
			wantErr: "exercise 0 must have at least one set",
		},
		{
			name: "reps <= 0",
			cmd: input.CreateSessionCommand{
				UserID:       "user-1",
				RoutineID:    "routine-1",
				RoutineName:  "Workout",
				DurationSecs: 1800,
				Exercises:    []input.SessionExerciseInput{{ExerciseID: "ex-1", Name: "Ex", Sets: []input.CompletedSetInput{{Weight: 50, Reps: 0, RestSeconds: 0}}}},
			},
			wantErr: "exercise 0, set 0: reps must be > 0",
		},
		{
			name: "weight < 0",
			cmd: input.CreateSessionCommand{
				UserID:       "user-1",
				RoutineID:    "routine-1",
				RoutineName:  "Workout",
				DurationSecs: 1800,
				Exercises:    []input.SessionExerciseInput{{ExerciseID: "ex-1", Name: "Ex", Sets: []input.CompletedSetInput{{Weight: -10, Reps: 10, RestSeconds: 0}}}},
			},
			wantErr: "exercise 0, set 0: weight must be >= 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := new(mockSessionRepo)
			idGen := new(mockIDGen)
			uc := NewCreateSession(repo, idGen)

			_, err := uc.Execute(context.Background(), tt.cmd)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestCreateSessionUseCase_Execute_RepositoryError(t *testing.T) {
	t.Parallel()

	repo := new(mockSessionRepo)
	idGen := new(mockIDGen)

	idGen.On("New").Return("session-123")
	repo.On("Save", mock.Anything, mock.Anything).Return(errors.New("dynamo error"))

	uc := NewCreateSession(repo, idGen)

	cmd := input.CreateSessionCommand{
		UserID:       "user-1",
		RoutineID:    "routine-1",
		RoutineName:  "Workout",
		DurationSecs: 1800,
		Exercises: []input.SessionExerciseInput{
			{
				ExerciseID: "ex-1",
				Name:       "Press",
				Sets:       []input.CompletedSetInput{{Weight: 60, Reps: 10, RestSeconds: 90}},
			},
		},
	}

	_, err := uc.Execute(context.Background(), cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "saving session")
}
