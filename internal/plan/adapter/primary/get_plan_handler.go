// get_plan_handler.go
package primary

import (
	"context"
	"errors"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/shotokan/gymtracker-api/internal/plan/domain"
	"github.com/shotokan/gymtracker-api/internal/plan/port/input"
)

// GetPlanHandler adapta API Gateway al caso de uso de obtener un plan.
type GetPlanHandler struct{ uc input.GetPlanUseCase }

// NewGetPlanHandler construye el handler para obtener un plan.
func NewGetPlanHandler(uc input.GetPlanUseCase) *GetPlanHandler { return &GetPlanHandler{uc: uc} }

// Handle procesa la solicitud de obtener un plan, validando el ID y delegando al caso de uso.
func (h *GetPlanHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest, userID string) (events.APIGatewayProxyResponse, error) {
	result, err := h.uc.Execute(ctx, input.GetPlanCommand{UserID: userID, PlanID: req.PathParameters["planId"]})
	if err != nil {
		if errors.Is(err, domain.ErrPlanNotFound) {
			return jsonResp(http.StatusNotFound, errResponse{"plan not found"})
		}
		return jsonResp(http.StatusInternalServerError, errResponse{"internal error"})
	}
	return jsonResp(http.StatusOK, result)
}
