// delete_plan_handler.go
package primary

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/shotokan/gymtracker-api/internal/plan/port/input"
)

// DeletePlanHandler adapta API Gateway al caso de uso de eliminar un plan.
type DeletePlanHandler struct{ uc input.DeletePlanUseCase }

// NewDeletePlanHandler construye el handler para eliminar un plan.
func NewDeletePlanHandler(uc input.DeletePlanUseCase) *DeletePlanHandler {
	return &DeletePlanHandler{uc: uc}
}

// Handle procesa la solicitud de eliminar un plan, validando el userID y delegando al caso de uso.
func (h *DeletePlanHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest, userID string) (events.APIGatewayProxyResponse, error) {
	if err := h.uc.Execute(ctx, input.DeletePlanCommand{UserID: userID, PlanID: req.PathParameters["planId"]}); err != nil {
		return jsonResp(http.StatusInternalServerError, errResponse{"internal error"})
	}
	return jsonResp(http.StatusNoContent, nil)
}
