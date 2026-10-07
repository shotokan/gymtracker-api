// update_plan_handler.go
package primary

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/shotokan/gymtracker-api/internal/plan/domain"
	"github.com/shotokan/gymtracker-api/internal/plan/port/input"
)

// UpdatePlanHandler adapta API Gateway al caso de uso de actualizar un plan.
type UpdatePlanHandler struct{ uc input.UpdatePlanUseCase }

// NewUpdatePlanHandler construye el handler para actualizar un plan.
func NewUpdatePlanHandler(uc input.UpdatePlanUseCase) *UpdatePlanHandler {
	return &UpdatePlanHandler{uc: uc}
}

// Handle procesa la solicitud de actualizar un plan, validando el cuerpo y delegando al caso de uso.
func (h *UpdatePlanHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest, userID string) (events.APIGatewayProxyResponse, error) {
	var body struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(req.Body), &body); err != nil || strings.TrimSpace(body.Name) == "" {
		return jsonResp(http.StatusBadRequest, errResponse{"name is required"})
	}

	result, err := h.uc.Execute(ctx, input.UpdatePlanCommand{UserID: userID, PlanID: req.PathParameters["planId"], Name: strings.TrimSpace(body.Name)})
	if err != nil {
		if errors.Is(err, domain.ErrPlanNotFound) {
			return jsonResp(http.StatusNotFound, errResponse{"plan not found"})
		}
		return jsonResp(http.StatusInternalServerError, errResponse{"internal error"})
	}
	return jsonResp(http.StatusOK, result)
}
