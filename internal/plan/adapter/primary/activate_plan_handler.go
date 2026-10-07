// activate_plan_handler.go
package primary

import (
	"context"
	"errors"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/shotokan/gymtracker-api/internal/plan/domain"
	"github.com/shotokan/gymtracker-api/internal/plan/port/input"
)

// ActivatePlanHandler adapta API Gateway al caso de uso de activar un plan.
type ActivatePlanHandler struct{ uc input.ActivatePlanUseCase }

// NewActivatePlanHandler construye el handler para activar un plan.
func NewActivatePlanHandler(uc input.ActivatePlanUseCase) *ActivatePlanHandler {
	return &ActivatePlanHandler{uc: uc}
}

// Handle procesa la solicitud de activación de un plan, validando el planId y delegando al caso de uso.
func (h *ActivatePlanHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest, userID string) (events.APIGatewayProxyResponse, error) {
	err := h.uc.Execute(ctx, input.ActivatePlanCommand{UserID: userID, PlanID: req.PathParameters["planId"]})
	if err != nil {
		if errors.Is(err, domain.ErrPlanNotFound) {
			return jsonResp(http.StatusNotFound, errResponse{"plan not found"})
		}
		return jsonResp(http.StatusInternalServerError, errResponse{"internal error"})
	}
	return jsonResp(http.StatusNoContent, nil)
}
