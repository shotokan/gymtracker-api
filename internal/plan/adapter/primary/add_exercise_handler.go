package primary

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/shotokan/gymtracker-api/internal/plan/domain"
	"github.com/shotokan/gymtracker-api/internal/plan/port/input"
)

// AddExerciseHandler handles POST /plans/{planId}/routines/{routineId}/exercises
type AddExerciseHandler struct {
	useCase input.AddExerciseUseCase
}

// NewAddExerciseHandler creates a new AddExerciseHandler.
func NewAddExerciseHandler(uc input.AddExerciseUseCase) *AddExerciseHandler {
	return &AddExerciseHandler{useCase: uc}
}

// Handle processes the request to add an exercise to a routine within a plan, validating the existence of the plan and routine, and returning the created exercise or an appropriate error response.
func (h *AddExerciseHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest, userID string) (events.APIGatewayProxyResponse, error) {
	planID := req.PathParameters["planId"]
	routineID := req.PathParameters["routineId"]

	var body struct {
		Name   string  `json:"name"`
		Sets   int     `json:"sets"`
		Reps   int     `json:"reps"`
		Weight float64 `json:"weight"`
		Notes  string  `json:"notes"`
	}
	if err := json.Unmarshal([]byte(req.Body), &body); err != nil || body.Name == "" {
		return jsonResp(http.StatusBadRequest, errResponse{"invalid request body"})
	}

	result, err := h.useCase.Execute(ctx, input.AddExerciseCommand{
		UserID:    userID,
		PlanID:    planID,
		RoutineID: routineID,
		Name:      body.Name,
		Sets:      body.Sets,
		Reps:      body.Reps,
		Weight:    body.Weight,
		Notes:     body.Notes,
	})
	if err != nil {
		if errors.Is(err, domain.ErrPlanNotFound) || errors.Is(err, domain.ErrRoutineNotFound) {
			return jsonResp(http.StatusNotFound, errResponse{err.Error()})
		}
		return jsonResp(http.StatusInternalServerError, errResponse{"internal error"})
	}

	return jsonResp(http.StatusCreated, result)
}
