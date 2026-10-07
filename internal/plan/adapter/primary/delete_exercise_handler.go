package primary

import (
	"context"
	"errors"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/shotokan/gymtracker-api/internal/plan/domain"
	"github.com/shotokan/gymtracker-api/internal/plan/port/input"
)

// DeleteExerciseHandler handles DELETE /plans/{planId}/routines/{routineId}/exercises/{exerciseId}
type DeleteExerciseHandler struct {
	useCase input.DeleteExerciseUseCase
}

// NewDeleteExerciseHandler creates a new DeleteExerciseHandler.
func NewDeleteExerciseHandler(uc input.DeleteExerciseUseCase) *DeleteExerciseHandler {
	return &DeleteExerciseHandler{useCase: uc}
}

// Handle processes the request to delete an exercise from a routine and plan, validating the existence of the plan, routine, and exercise, and returning an appropriate success or error response.
func (h *DeleteExerciseHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest, userID string) (events.APIGatewayProxyResponse, error) {
	planID := req.PathParameters["planId"]
	routineID := req.PathParameters["routineId"]
	exerciseID := req.PathParameters["exerciseId"]

	err := h.useCase.Execute(ctx, input.DeleteExerciseCommand{
		UserID:     userID,
		PlanID:     planID,
		RoutineID:  routineID,
		ExerciseID: exerciseID,
	})
	if err != nil {
		if errors.Is(err, domain.ErrPlanNotFound) || errors.Is(err, domain.ErrRoutineNotFound) || errors.Is(err, domain.ErrExerciseNotFound) {
			return jsonResp(http.StatusNotFound, errResponse{err.Error()})
		}
		return jsonResp(http.StatusInternalServerError, errResponse{"internal error"})
	}

	return jsonResp(http.StatusNoContent, nil)
}
