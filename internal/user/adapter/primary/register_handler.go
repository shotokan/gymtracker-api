// register_handler.go
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

// RegisterHandler adapta API Gateway al caso de uso de registro de usuario.
type RegisterHandler struct{ useCase input.RegisterUseCase }

// NewRegisterHandler construye el handler del registro de usuario.
func NewRegisterHandler(uc input.RegisterUseCase) *RegisterHandler {
	return &RegisterHandler{useCase: uc}
}

// Handle procesa la solicitud de registro de usuario, validando los datos de entrada y delegando al caso de uso. Retorna el resultado del registro o un error adecuado.
func (h *RegisterHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
		return jsonResp(http.StatusBadRequest, errResponse{"invalid request body"})
	}
	if strings.TrimSpace(body.Email) == "" || len(body.Password) < 8 || strings.TrimSpace(body.Name) == "" {
		return jsonResp(http.StatusBadRequest, errResponse{"email, name and password (min 8 chars) are required"})
	}

	result, err := h.useCase.Execute(ctx, input.RegisterCommand{
		Email:    strings.ToLower(strings.TrimSpace(body.Email)),
		Password: body.Password,
		Name:     strings.TrimSpace(body.Name),
	})
	if err != nil {
		if errors.Is(err, domain.ErrAlreadyExists) {
			return jsonResp(http.StatusConflict, errResponse{"email already registered"})
		}
		return jsonResp(http.StatusInternalServerError, errResponse{"internal error"})
	}

	return jsonResp(http.StatusCreated, result)
}
