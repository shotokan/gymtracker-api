package primary

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/shotokan/gymtracker-api/internal/session/domain"
	"github.com/shotokan/gymtracker-api/internal/session/port/input"
)

// ── CreateSessionHandler ──────────────────────────────────────────────────

// CreateSessionHandler handles POST /sessions
type CreateSessionHandler struct{ uc input.CreateSessionUseCase }

// NewCreateSessionHandler constructs the handler for creating a session.
func NewCreateSessionHandler(uc input.CreateSessionUseCase) *CreateSessionHandler {
	return &CreateSessionHandler{uc: uc}
}

// Handle processes the session creation request.
func (h *CreateSessionHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest, userID string) (events.APIGatewayProxyResponse, error) {
	var body struct {
		PlanID       string                       `json:"plan_id"`
		RoutineID    string                       `json:"routine_id"`
		RoutineName  string                       `json:"routine_name"`
		Date         string                       `json:"date"`
		DurationSecs int                          `json:"duration_secs"`
		Exercises    []input.SessionExerciseInput `json:"exercises"`
		StartedAt    string                       `json:"started_at"`
		FinishedAt   string                       `json:"finished_at"`
	}

	if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
		return jsonResp(http.StatusBadRequest, errResponse{"invalid request body"})
	}

	if strings.TrimSpace(body.RoutineID) == "" {
		return jsonResp(http.StatusBadRequest, errResponse{"routine_id is required"})
	}
	if strings.TrimSpace(body.RoutineName) == "" {
		return jsonResp(http.StatusBadRequest, errResponse{"routine_name is required"})
	}

	result, err := h.uc.Execute(ctx, input.CreateSessionCommand{
		UserID:       userID,
		PlanID:       body.PlanID,
		RoutineID:    body.RoutineID,
		RoutineName:  body.RoutineName,
		Date:         body.Date,
		DurationSecs: body.DurationSecs,
		Exercises:    body.Exercises,
		StartedAt:    body.StartedAt,
		FinishedAt:   body.FinishedAt,
	})
	if err != nil {
		return jsonResp(http.StatusInternalServerError, errResponse{"internal error"})
	}

	return jsonResp(http.StatusCreated, result)
}

// ── ListSessionsHandler ───────────────────────────────────────────────────

// ListSessionsHandler handles GET /sessions
type ListSessionsHandler struct{ uc input.ListSessionsUseCase }

// NewListSessionsHandler constructs the handler for listing sessions.
func NewListSessionsHandler(uc input.ListSessionsUseCase) *ListSessionsHandler {
	return &ListSessionsHandler{uc: uc}
}

// Handle processes the list sessions request.
func (h *ListSessionsHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest, userID string) (events.APIGatewayProxyResponse, error) {
	limit := 20
	if limitStr := req.QueryStringParameters["limit"]; limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	results, err := h.uc.Execute(ctx, input.ListSessionsCommand{UserID: userID, Limit: limit})
	if err != nil {
		return jsonResp(http.StatusInternalServerError, errResponse{"internal error"})
	}

	// Return empty array if no sessions, never null
	if results == nil {
		results = []input.SessionResult{}
	}

	return jsonResp(http.StatusOK, results)
}

// ── GetSessionHandler ─────────────────────────────────────────────────────

// GetSessionHandler handles GET /sessions/{sessionId}
type GetSessionHandler struct{ uc input.GetSessionUseCase }

// NewGetSessionHandler constructs the handler for getting a session.
func NewGetSessionHandler(uc input.GetSessionUseCase) *GetSessionHandler {
	return &GetSessionHandler{uc: uc}
}

// Handle processes the get session request.
func (h *GetSessionHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest, userID string) (events.APIGatewayProxyResponse, error) {
	result, err := h.uc.Execute(ctx, input.GetSessionCommand{
		UserID:    userID,
		SessionID: req.PathParameters["sessionId"],
	})
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			return jsonResp(http.StatusNotFound, errResponse{"session not found"})
		}
		return jsonResp(http.StatusInternalServerError, errResponse{"internal error"})
	}
	return jsonResp(http.StatusOK, result)
}
