package primary

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

// Router dispatches API Gateway events to the correct handler.
type Router struct {
	createSession *CreateSessionHandler
	listSessions  *ListSessionsHandler
	getSession    *GetSessionHandler
}

// NewRouter constructs the router with all handlers.
func NewRouter(
	createSession *CreateSessionHandler,
	listSessions *ListSessionsHandler,
	getSession *GetSessionHandler,
) *Router {
	return &Router{
		createSession: createSession,
		listSessions:  listSessions,
		getSession:    getSession,
	}
}

// Handle dispatches the incoming request to the correct handler.
func (r *Router) Handle(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID := extractUserID(req)
	if userID == "" {
		return jsonResp(http.StatusUnauthorized, errResponse{"unauthorized"})
	}

	method := req.HTTPMethod
	resource := req.Resource

	switch {
	case method == "POST" && resource == "/sessions":
		return r.createSession.Handle(ctx, req, userID)
	case method == "GET" && resource == "/sessions":
		return r.listSessions.Handle(ctx, req, userID)
	case method == "GET" && resource == "/sessions/{sessionId}":
		return r.getSession.Handle(ctx, req, userID)
	default:
		return jsonResp(http.StatusNotFound, errResponse{"route not found"})
	}
}
