// deactivate_handler.go
package primary

import (
	"context"
	"errors"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/shotokan/gymtracker-api/internal/user/domain"
	"github.com/shotokan/gymtracker-api/internal/user/port/input"
)

// DeactivateHandler adapta API Gateway al caso de uso de desactivación de usuario.
type DeactivateHandler struct{ useCase input.DeactivateUseCase }

// NewDeactivateHandler construye el handler del endpoint de desactivación de usuario.
func NewDeactivateHandler(uc input.DeactivateUseCase) *DeactivateHandler {
	return &DeactivateHandler{useCase: uc}
}

// Handle procesa la solicitud de desactivación de usuario, extrayendo el ID del usuario de los parámetros de ruta y delegando al caso de uso. Retorna un código 204 No Content si la operación es exitosa, o un error adecuado si ocurre algún problema.
func (h *DeactivateHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID := req.PathParameters["id"]
	if userID == "" {
		return jsonResp(http.StatusBadRequest, errResponse{"missing user id"})
	}

	if err := h.useCase.Execute(ctx, input.DeactivateCommand{UserID: userID}); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return jsonResp(http.StatusNotFound, errResponse{"user not found"})
		}
		return jsonResp(http.StatusInternalServerError, errResponse{"internal error"})
	}

	return jsonResp(http.StatusNoContent, nil)
}
