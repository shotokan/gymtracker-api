package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shotokan/gymtracker-api/internal/session/domain"
	"github.com/shotokan/gymtracker-api/internal/session/port/input"
	"github.com/shotokan/gymtracker-api/internal/session/port/output"
)

type idGenerator interface {
	New() string
}

type createSessionUseCase struct {
	repo  output.SessionRepository
	idGen idGenerator
}

// NewCreateSession constructs the use case for creating a session.
func NewCreateSession(repo output.SessionRepository, idGen idGenerator) input.CreateSessionUseCase {
	return &createSessionUseCase{repo: repo, idGen: idGen}
}

// Execute creates a new workout session for the user.
func (uc *createSessionUseCase) Execute(ctx context.Context, cmd input.CreateSessionCommand) (input.SessionResult, error) {
	// Validate required fields
	if strings.TrimSpace(cmd.RoutineID) == "" {
		return input.SessionResult{}, errors.New("routine_id is required")
	}
	if strings.TrimSpace(cmd.RoutineName) == "" {
		return input.SessionResult{}, errors.New("routine_name is required")
	}
	if cmd.DurationSecs < 0 {
		return input.SessionResult{}, errors.New("duration_secs must be >= 0")
	}

	// Validate exercises
	if len(cmd.Exercises) == 0 {
		return input.SessionResult{}, errors.New("at least one exercise is required")
	}
	for i, ex := range cmd.Exercises {
		if len(ex.Sets) == 0 {
			return input.SessionResult{}, fmt.Errorf("exercise %d must have at least one set", i)
		}
		for j, set := range ex.Sets {
			if set.Reps <= 0 {
				return input.SessionResult{}, fmt.Errorf("exercise %d, set %d: reps must be > 0", i, j)
			}
			if set.Weight < 0 {
				return input.SessionResult{}, fmt.Errorf("exercise %d, set %d: weight must be >= 0", i, j)
			}
		}
	}

	// Default date to today if empty
	date := cmd.Date
	if strings.TrimSpace(date) == "" {
		date = time.Now().UTC().Format("2006-01-02")
	}

	// Parse started_at and finished_at
	var startedAt, finishedAt time.Time
	var err error

	if cmd.StartedAt != "" {
		startedAt, err = time.Parse(time.RFC3339, cmd.StartedAt)
		if err != nil {
			return input.SessionResult{}, fmt.Errorf("invalid started_at format: %w", err)
		}
	} else {
		startedAt = time.Now().UTC()
	}

	if cmd.FinishedAt != "" {
		finishedAt, err = time.Parse(time.RFC3339, cmd.FinishedAt)
		if err != nil {
			return input.SessionResult{}, fmt.Errorf("invalid finished_at format: %w", err)
		}
	} else {
		finishedAt = time.Now().UTC()
	}

	// Calculate duration_secs from started_at and finished_at
	durationSecs := int(finishedAt.Sub(startedAt).Seconds())
	if durationSecs < 0 {
		durationSecs = 0
	}

	// Build domain exercises
	exercises := make([]domain.SessionExercise, 0, len(cmd.Exercises))
	for _, ex := range cmd.Exercises {
		sets := make([]domain.CompletedSet, 0, len(ex.Sets))
		for _, s := range ex.Sets {
			sets = append(sets, domain.CompletedSet{
				Weight:      s.Weight,
				Reps:        s.Reps,
				RestSeconds: s.RestSeconds,
			})
		}
		exercises = append(exercises, domain.SessionExercise{
			ExerciseID: ex.ExerciseID,
			Name:       ex.Name,
			Sets:       sets,
		})
	}

	now := time.Now().UTC()
	session := &domain.Session{
		ID:           uc.idGen.New(),
		UserID:       cmd.UserID,
		PlanID:       cmd.PlanID,
		RoutineID:    cmd.RoutineID,
		RoutineName:  cmd.RoutineName,
		Date:         date,
		DurationSecs: durationSecs,
		Exercises:    exercises,
		StartedAt:    startedAt,
		FinishedAt:   finishedAt,
		CreatedAt:    now,
	}

	if err := uc.repo.Save(ctx, session); err != nil {
		return input.SessionResult{}, fmt.Errorf("saving session: %w", err)
	}

	return toSessionResult(session), nil
}
