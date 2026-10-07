// create_plan_handler.go
package primary

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/shotokan/gymtracker-api/internal/plan/port/input"
)

// CreatePlanHandler adapta API Gateway al caso de uso de crear un plan.
type CreatePlanHandler struct{ uc input.CreatePlanUseCase }

// NewCreatePlanHandler construye el handler para crear un plan.
func NewCreatePlanHandler(uc input.CreatePlanUseCase) *CreatePlanHandler {
	return &CreatePlanHandler{uc: uc}
}

// Handle procesa la solicitud de creación de un plan, validando el cuerpo y delegando al caso de uso.
func (h *CreatePlanHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest, userID string) (events.APIGatewayProxyResponse, error) {
	var body struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(req.Body), &body); err != nil || strings.TrimSpace(body.Name) == "" {
		return jsonResp(http.StatusBadRequest, errResponse{"name is required"})
	}

	result, err := h.uc.Execute(ctx, input.CreatePlanCommand{UserID: userID, Name: strings.TrimSpace(body.Name)})
	if err != nil {
		return jsonResp(http.StatusInternalServerError, errResponse{"internal error"})
	}

	return jsonResp(http.StatusCreated, result)
}
