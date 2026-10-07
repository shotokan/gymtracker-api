package input

import "context"

// ── Result types ──────────────────────────────────────────────────────────

// CompletedSetResult represents a completed set in the API response.
type CompletedSetResult struct {
	Weight      float64 `json:"weight"`
	Reps        int     `json:"reps"`
	RestSeconds int     `json:"rest_seconds"`
}

// SessionExerciseResult represents an exercise within a session in the API response.
type SessionExerciseResult struct {
	ExerciseID string               `json:"exercise_id"`
	Name       string               `json:"name"`
	Sets       []CompletedSetResult `json:"sets"`
}

// SessionResult represents a session in the API response.
type SessionResult struct {
	ID               string                  `json:"id"`
	PlanID           string                  `json:"plan_id,omitempty"`
	RoutineID        string                  `json:"routine_id"`
	RoutineName      string                  `json:"routine_name"`
	Date             string                  `json:"date"`
	DurationSecs     int                     `json:"duration_secs"`
	Exercises        []SessionExerciseResult `json:"exercises"`
	TotalSets        int                     `json:"total_sets"`
	TotalVolume      float64                 `json:"total_volume"`
	TotalRestSeconds int                     `json:"total_rest_seconds"`
	StartedAt        string                  `json:"started_at"`
	FinishedAt       string                  `json:"finished_at"`
	CreatedAt        string                  `json:"created_at"`
}

// ── Input types for CreateSession ─────────────────────────────────────────

// CompletedSetInput represents a completed set in the request.
type CompletedSetInput struct {
	Weight      float64 `json:"weight"`
	Reps        int     `json:"reps"`
	RestSeconds int     `json:"rest_seconds"`
}

// SessionExerciseInput represents an exercise within a session in the request.
type SessionExerciseInput struct {
	ExerciseID string              `json:"exercise_id"`
	Name       string              `json:"name"`
	Sets       []CompletedSetInput `json:"sets"`
}

// ── Create ────────────────────────────────────────────────────────────────

// CreateSessionCommand represents the request to create a new session.
type CreateSessionCommand struct {
	UserID       string
	PlanID       string
	RoutineID    string
	RoutineName  string
	Date         string
	DurationSecs int
	Exercises    []SessionExerciseInput
	StartedAt    string // ISO 8601 format
	FinishedAt   string // ISO 8601 format
}

// CreateSessionUseCase is the primary port for creating a new workout session.
type CreateSessionUseCase interface {
	Execute(ctx context.Context, cmd CreateSessionCommand) (SessionResult, error)
}

// ── List ──────────────────────────────────────────────────────────────────

// ListSessionsCommand represents the request to list sessions for a user.
type ListSessionsCommand struct {
	UserID string
	Limit  int
}

// ListSessionsUseCase is the primary port for listing sessions for a user.
type ListSessionsUseCase interface {
	Execute(ctx context.Context, cmd ListSessionsCommand) ([]SessionResult, error)
}

// ── Get ───────────────────────────────────────────────────────────────────

// GetSessionCommand represents the request to get a specific session.
type GetSessionCommand struct {
	UserID    string
	SessionID string
}

// GetSessionUseCase is the primary port for getting a specific session.
type GetSessionUseCase interface {
	Execute(ctx context.Context, cmd GetSessionCommand) (SessionResult, error)
}
