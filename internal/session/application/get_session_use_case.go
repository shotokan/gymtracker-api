package application

import (
	"context"
	"fmt"

	"github.com/shotokan/gymtracker-api/internal/session/port/input"
	"github.com/shotokan/gymtracker-api/internal/session/port/output"
)

type getSessionUseCase struct{ repo output.SessionRepository }

// NewGetSession constructs the use case for getting a session.
func NewGetSession(repo output.SessionRepository) input.GetSessionUseCase {
	return &getSessionUseCase{repo: repo}
}

// Execute retrieves a specific session for the user.
func (uc *getSessionUseCase) Execute(ctx context.Context, cmd input.GetSessionCommand) (input.SessionResult, error) {
	session, err := uc.repo.FindByID(ctx, cmd.UserID, cmd.SessionID)
	if err != nil {
		return input.SessionResult{}, fmt.Errorf("finding session: %w", err)
	}
	return toSessionResult(session), nil
}
