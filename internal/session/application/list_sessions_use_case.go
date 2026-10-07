package application

import (
	"context"
	"fmt"

	"github.com/shotokan/gymtracker-api/internal/session/port/input"
	"github.com/shotokan/gymtracker-api/internal/session/port/output"
)

type listSessionsUseCase struct{ repo output.SessionRepository }

// NewListSessions constructs the use case for listing sessions.
func NewListSessions(repo output.SessionRepository) input.ListSessionsUseCase {
	return &listSessionsUseCase{repo: repo}
}

// Execute lists sessions for the user, applying the limit.
func (uc *listSessionsUseCase) Execute(ctx context.Context, cmd input.ListSessionsCommand) ([]input.SessionResult, error) {
	// Clamp limit: if <= 0 or > 100, use 20
	limit := cmd.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	sessions, err := uc.repo.FindAll(ctx, cmd.UserID, limit)
	if err != nil {
		return nil, fmt.Errorf("listing sessions: %w", err)
	}

	results := make([]input.SessionResult, len(sessions))
	for i, s := range sessions {
		results[i] = toSessionResult(&s)
	}

	return results, nil
}
