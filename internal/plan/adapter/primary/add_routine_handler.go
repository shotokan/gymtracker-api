// add_routine_handler.go
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

// AddRoutineHandler adapta API Gateway al caso de uso de agregar una rutina a un plan.
type AddRoutineHandler struct{ uc input.AddRoutineUseCase }

// NewAddRoutineHandler construye el handler para agregar una rutina a un plan.
func NewAddRoutineHandler(uc input.AddRoutineUseCase) *AddRoutineHandler {
	return &AddRoutineHandler{uc: uc}
}

// Handle procesa la solicitud de agregar una rutina a un plan, validando el cuerpo y delegando al caso de uso.
func (h *AddRoutineHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest, userID string) (events.APIGatewayProxyResponse, error) {
	var body struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(req.Body), &body); err != nil || strings.TrimSpace(body.Name) == "" {
		return jsonResp(http.StatusBadRequest, errResponse{"name is required"})
	}

	result, err := h.uc.Execute(ctx, input.AddRoutineCommand{
		UserID: userID, PlanID: req.PathParameters["planId"], Name: strings.TrimSpace(body.Name),
	})
	if err != nil {
		if errors.Is(err, domain.ErrPlanNotFound) {
			return jsonResp(http.StatusNotFound, errResponse{"plan not found"})
		}
		return jsonResp(http.StatusInternalServerError, errResponse{"internal error"})
	}
	return jsonResp(http.StatusCreated, result)
}
