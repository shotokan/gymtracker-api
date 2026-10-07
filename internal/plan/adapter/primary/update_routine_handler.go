// update_routine_handler.go
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

// UpdateRoutineHandler adapta API Gateway al caso de uso de actualizar una rutina.
type UpdateRoutineHandler struct{ uc input.UpdateRoutineUseCase }

// NewUpdateRoutineHandler construye el handler para actualizar una rutina.
func NewUpdateRoutineHandler(uc input.UpdateRoutineUseCase) *UpdateRoutineHandler {
	return &UpdateRoutineHandler{uc: uc}
}

// Handle procesa la solicitud de actualizar una rutina, validando el cuerpo y delegando al caso de uso.
func (h *UpdateRoutineHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest, userID string) (events.APIGatewayProxyResponse, error) {
	var body struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(req.Body), &body); err != nil || strings.TrimSpace(body.Name) == "" {
		return jsonResp(http.StatusBadRequest, errResponse{"name is required"})
	}

	result, err := h.uc.Execute(ctx, input.UpdateRoutineCommand{
		UserID: userID, PlanID: req.PathParameters["planId"],
		RoutineID: req.PathParameters["routineId"], Name: strings.TrimSpace(body.Name),
	})
	if err != nil {
		if errors.Is(err, domain.ErrRoutineNotFound) {
			return jsonResp(http.StatusNotFound, errResponse{"routine not found"})
		}
		return jsonResp(http.StatusInternalServerError, errResponse{"internal error"})
	}
	return jsonResp(http.StatusOK, result)
}
