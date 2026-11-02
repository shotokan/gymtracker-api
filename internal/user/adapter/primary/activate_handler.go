// activate_handler.go
package primary

import (
	"context"
	"errors"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/shotokan/gymtracker-api/internal/user/domain"
	"github.com/shotokan/gymtracker-api/internal/user/port/input"
)

// ActivateHandler adapta API Gateway al caso de uso de activación de usuario.
type ActivateHandler struct{ useCase input.ActivateUseCase }

// NewActivateHandler construye el handler del endpoint de activación de usuario.
func NewActivateHandler(uc input.ActivateUseCase) *ActivateHandler {
	return &ActivateHandler{useCase: uc}
}

// Handle procesa la solicitud de activación de usuario, tomando al administrador del principalId del Authorizer y al usuario de los parámetros de ruta. Retorna un código 204 No Content si la operación es exitosa, o un error adecuado si ocurre algún problema.
func (h *ActivateHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	actorID := extractUserID(req)
	if actorID == "" {
		return jsonResp(http.StatusUnauthorized, errResponse{"unauthorized"})
	}

	userID := req.PathParameters["id"]
	if userID == "" {
		return jsonResp(http.StatusBadRequest, errResponse{"missing user id"})
	}

	err := h.useCase.Execute(ctx, input.ActivateCommand{ActorID: actorID, UserID: userID})
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return jsonResp(http.StatusForbidden, errResponse{"admin role required"})
		}
		if errors.Is(err, domain.ErrNotFound) {
			return jsonResp(http.StatusNotFound, errResponse{"user not found"})
		}
		return jsonResp(http.StatusInternalServerError, errResponse{"internal error"})
	}

	return jsonResp(http.StatusNoContent, nil)
}
