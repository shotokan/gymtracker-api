// get_active_plan_handler.go
package primary

import (
	"context"
	"errors"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/shotokan/gymtracker-api/internal/plan/domain"
	"github.com/shotokan/gymtracker-api/internal/plan/port/input"
)

// GetActivePlanHandler adapta API Gateway al caso de uso de obtener el plan activo de un usuario.
type GetActivePlanHandler struct{ uc input.GetActivePlanUseCase }

// NewGetActivePlanHandler construye el handler para obtener el plan activo de un usuario.
func NewGetActivePlanHandler(uc input.GetActivePlanUseCase) *GetActivePlanHandler {
	return &GetActivePlanHandler{uc: uc}
}

// Handle procesa la solicitud de obtener el plan activo de un usuario, validando el userID y delegando al caso de uso.
func (h *GetActivePlanHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest, userID string) (events.APIGatewayProxyResponse, error) {
	result, err := h.uc.Execute(ctx, input.GetActivePlanCommand{UserID: userID})
	if err != nil {
		if errors.Is(err, domain.ErrNoActivePlan) {
			return jsonResp(http.StatusNotFound, errResponse{"no active plan"})
		}
		return jsonResp(http.StatusInternalServerError, errResponse{"internal error"})
	}
	return jsonResp(http.StatusOK, result)
}
