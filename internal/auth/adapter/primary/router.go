package primary

import (
	"context"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

// Router dispatches API Gateway events to the correct auth handler.
type Router struct {
	loginHandler *LoginHandler
}

// NewRouter constructs the router with all auth handlers.
func NewRouter(loginHandler *LoginHandler) *Router {
	return &Router{loginHandler: loginHandler}
}

// Handle dispatches the incoming request to the correct handler.
func (r *Router) Handle(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Normalize path by removing /v1 prefix if present
	path := strings.TrimPrefix(req.Path, "/v1")

	switch path {
	case "/auth/login":
		return r.loginHandler.Handle(ctx, req)
	default:
		return jsonResp(http.StatusNotFound, errResponse{"route not found"})
	}
}
