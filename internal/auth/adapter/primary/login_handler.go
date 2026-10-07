package primary

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/shotokan/gymtracker-api/internal/auth/domain"
	"github.com/shotokan/gymtracker-api/internal/auth/port/input"
)

// LoginHandler handles POST /auth/login requests.
type LoginHandler struct {
	uc input.LoginUseCase
}

// NewLoginHandler creates a handler for email/password authentication.
func NewLoginHandler(uc input.LoginUseCase) *LoginHandler {
	return &LoginHandler{uc: uc}
}

// Handle processes the login request.
func (h *LoginHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
		return jsonResp(http.StatusBadRequest, errResponse{"invalid request body"})
	}

	if body.Email == "" || body.Password == "" {
		return jsonResp(http.StatusBadRequest, errResponse{"email and password are required"})
	}

	result, err := h.uc.Execute(ctx, input.LoginCommand{
		Email:    body.Email,
		Password: body.Password,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCreds) {
			return jsonResp(http.StatusUnauthorized, errResponse{"invalid credentials"})
		}
		return jsonResp(http.StatusInternalServerError, errResponse{"internal error"})
	}

	return jsonResp(http.StatusOK, result)
}
