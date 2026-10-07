// get_profile_handler.go
package primary

import (
	"context"
	"errors"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/shotokan/gymtracker-api/internal/user/domain"
	"github.com/shotokan/gymtracker-api/internal/user/port/input"
)

// GetProfileHandler adapta API Gateway al caso de uso de obtener el perfil de usuario.
type GetProfileHandler struct{ useCase input.GetProfileUseCase }

// NewGetProfileHandler construye el handler del perfil de usuario.
func NewGetProfileHandler(uc input.GetProfileUseCase) *GetProfileHandler {
	return &GetProfileHandler{useCase: uc}
}

// Handle procesa la solicitud de obtener el perfil de usuario, extrayendo el ID del usuario de los parámetros de la ruta y delegando al caso de uso. Retorna el perfil del usuario o un error adecuado.
func (h *GetProfileHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID := req.PathParameters["id"]
	if userID == "" {
		return jsonResp(http.StatusBadRequest, errResponse{"missing user id"})
	}

	result, err := h.useCase.Execute(ctx, input.GetProfileCommand{UserID: userID})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return jsonResp(http.StatusNotFound, errResponse{"user not found"})
		}
		return jsonResp(http.StatusInternalServerError, errResponse{"internal error"})
	}

	return jsonResp(http.StatusOK, result)
}
