// delete_routine_handler.go
package primary

import (
	"context"
	"errors"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/shotokan/gymtracker-api/internal/plan/domain"
	"github.com/shotokan/gymtracker-api/internal/plan/port/input"
)

// DeleteRoutineHandler adapta API Gateway al caso de uso de eliminar una rutina de un plan.
type DeleteRoutineHandler struct{ uc input.DeleteRoutineUseCase }

// NewDeleteRoutineHandler construye el handler para eliminar una rutina de un plan.
func NewDeleteRoutineHandler(uc input.DeleteRoutineUseCase) *DeleteRoutineHandler {
	return &DeleteRoutineHandler{uc: uc}
}

// Handle procesa la solicitud de eliminación de una rutina de un plan, validando los IDs y delegando al caso de uso.
func (h *DeleteRoutineHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest, userID string) (events.APIGatewayProxyResponse, error) {
	err := h.uc.Execute(ctx, input.DeleteRoutineCommand{
		UserID: userID, PlanID: req.PathParameters["planId"], RoutineID: req.PathParameters["routineId"],
	})
	if err != nil {
		if errors.Is(err, domain.ErrRoutineNotFound) {
			return jsonResp(http.StatusNotFound, errResponse{"routine not found"})
		}
		return jsonResp(http.StatusInternalServerError, errResponse{"internal error"})
	}
	return jsonResp(http.StatusNoContent, nil)
}
