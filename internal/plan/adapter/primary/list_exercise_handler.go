package primary

import (
	"context"
	"errors"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/shotokan/gymtracker-api/internal/plan/domain"
	"github.com/shotokan/gymtracker-api/internal/plan/port/input"
)

// ListExercisesHandler handles GET /plans/{planId}/routines/{routineId}/exercises
type ListExercisesHandler struct {
	useCase input.ListExercisesUseCase
}

// NewListExercisesHandler creates a new ListExercisesHandler.
func NewListExercisesHandler(uc input.ListExercisesUseCase) *ListExercisesHandler {
	return &ListExercisesHandler{useCase: uc}
}

// Handle processes the request to list exercises of a routine within a plan, validating the existence of the plan and routine, and returning the list of exercises or an appropriate error response.
func (h *ListExercisesHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest, userID string) (events.APIGatewayProxyResponse, error) {
	planID := req.PathParameters["planId"]
	routineID := req.PathParameters["routineId"]

	results, err := h.useCase.Execute(ctx, input.ListExercisesCommand{
		UserID:    userID,
		PlanID:    planID,
		RoutineID: routineID,
	})
	if err != nil {
		if errors.Is(err, domain.ErrPlanNotFound) || errors.Is(err, domain.ErrRoutineNotFound) {
			return jsonResp(http.StatusNotFound, errResponse{err.Error()})
		}
		return jsonResp(http.StatusInternalServerError, errResponse{"internal error"})
	}

	return jsonResp(http.StatusOK, results)
}
