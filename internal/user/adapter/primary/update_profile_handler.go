// update_profile_handler.go
package primary

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/shotokan/gymtracker-api/internal/user/domain"
	"github.com/shotokan/gymtracker-api/internal/user/port/input"
)

// UpdateProfileHandler adapta API Gateway al caso de uso de actualización del perfil de usuario.
type UpdateProfileHandler struct{ useCase input.UpdateProfileUseCase }

// NewUpdateProfileHandler construye el handler del perfil de usuario.
func NewUpdateProfileHandler(uc input.UpdateProfileUseCase) *UpdateProfileHandler {
	return &UpdateProfileHandler{useCase: uc}
}

// Handle procesa la solicitud de actualización del perfil de usuario, extrayendo el ID del usuario de los parámetros de la ruta y delegando al caso de uso. Retorna el perfil del usuario o un error adecuado.
func (h *UpdateProfileHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	actorID := extractUserID(req)
	if actorID == "" {
		return jsonResp(http.StatusUnauthorized, errResponse{"unauthorized"})
	}

	userID := req.PathParameters["id"]
	if userID == "" {
		return jsonResp(http.StatusBadRequest, errResponse{"missing user id"})
	}

	var body struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
		return jsonResp(http.StatusBadRequest, errResponse{"invalid request body"})
	}
	if strings.TrimSpace(body.Name) == "" {
		return jsonResp(http.StatusBadRequest, errResponse{"name is required"})
	}

	result, err := h.useCase.Execute(ctx, input.UpdateProfileCommand{
		ActorID: actorID,
		UserID:  userID,
		Name:    strings.TrimSpace(body.Name),
	})
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return jsonResp(http.StatusForbidden, errResponse{"forbidden"})
		}
		if errors.Is(err, domain.ErrNotFound) {
			return jsonResp(http.StatusNotFound, errResponse{"user not found"})
		}
		return jsonResp(http.StatusInternalServerError, errResponse{"internal error"})
	}

	return jsonResp(http.StatusOK, result)
}
