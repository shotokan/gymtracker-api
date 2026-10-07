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

// GoogleAuthHandler handles POST /auth/google requests.
type GoogleAuthHandler struct {
	uc input.GoogleAuthUseCase
}

// NewGoogleAuthHandler creates a handler for Google OAuth authentication.
func NewGoogleAuthHandler(uc input.GoogleAuthUseCase) *GoogleAuthHandler {
	return &GoogleAuthHandler{uc: uc}
}

// Handle processes the Google OAuth authentication request.
func (h *GoogleAuthHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var body struct {
		IDToken string `json:"id_token"`
	}

	if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
		return jsonResp(http.StatusBadRequest, errResponse{"invalid request body"})
	}

	if body.IDToken == "" {
		return jsonResp(http.StatusBadRequest, errResponse{"id_token is required"})
	}

	result, err := h.uc.Execute(ctx, input.GoogleAuthCommand{
		IDToken: body.IDToken,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidToken) {
			return jsonResp(http.StatusUnauthorized, errResponse{"invalid or expired Google token"})
		}
		return jsonResp(http.StatusInternalServerError, errResponse{"internal error"})
	}

	return jsonResp(http.StatusOK, result)
}

// ── Helpers ──────────────────────────────────────────────────────────────────

type errResponse struct {
	Error string `json:"error"`
}

func jsonResp(statusCode int, body interface{}) (events.APIGatewayProxyResponse, error) {
	data, _ := json.Marshal(body)
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(data),
	}, nil
}
