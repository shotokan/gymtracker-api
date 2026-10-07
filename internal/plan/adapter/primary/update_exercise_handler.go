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

// UpdateExerciseHandler handles PATCH /plans/{planId}/routines/{routineId}/exercises/{exerciseId}
type UpdateExerciseHandler struct {
	useCase input.UpdateExerciseUseCase
}

// NewUpdateExerciseHandler creates a new UpdateExerciseHandler.
func NewUpdateExerciseHandler(uc input.UpdateExerciseUseCase) *UpdateExerciseHandler {
	return &UpdateExerciseHandler{useCase: uc}
}

// Handle processes the request to update an exercise within a routine and plan, validating the existence of the plan, routine, and exercise, and returning the updated exercise or an appropriate error response.
func (h *UpdateExerciseHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest, userID string) (events.APIGatewayProxyResponse, error) {
	planID := req.PathParameters["planId"]
	routineID := req.PathParameters["routineId"]
	exerciseID := req.PathParameters["exerciseId"]

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

	result, err := h.useCase.Execute(ctx, input.UpdateExerciseCommand{
		UserID:     userID,
		PlanID:     planID,
		RoutineID:  routineID,
		ExerciseID: exerciseID,
		Name:       body.Name,
		Sets:       body.Sets,
		Reps:       body.Reps,
		Weight:     body.Weight,
		Notes:      body.Notes,
	})
	if err != nil {
		if errors.Is(err, domain.ErrPlanNotFound) || errors.Is(err, domain.ErrRoutineNotFound) || errors.Is(err, domain.ErrExerciseNotFound) {
			return jsonResp(http.StatusNotFound, errResponse{err.Error()})
		}
		return jsonResp(http.StatusInternalServerError, errResponse{"internal error"})
	}

	return jsonResp(http.StatusOK, result)
}
