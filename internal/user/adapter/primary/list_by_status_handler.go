// list_by_status_handler.go
package primary

import (
	"context"
	"errors"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/shotokan/gymtracker-api/internal/user/domain"
	"github.com/shotokan/gymtracker-api/internal/user/port/input"
)

// ListByStatusHandler adapta API Gateway al caso de uso de listado de usuarios por estado.
type ListByStatusHandler struct{ useCase input.ListByStatusUseCase }

// NewListByStatusHandler construye el handler del listado de usuarios por estado.
func NewListByStatusHandler(uc input.ListByStatusUseCase) *ListByStatusHandler {
	return &ListByStatusHandler{useCase: uc}
}

// Handle procesa GET /users?status=pending|deactivated, tomando al administrador del principalId del Authorizer. Retorna los usuarios con ese estado o un error adecuado.
func (h *ListByStatusHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	actorID := extractUserID(req)
	if actorID == "" {
		return jsonResp(http.StatusUnauthorized, errResponse{"unauthorized"})
	}

	status := req.QueryStringParameters["status"]
	if status != domain.StatusPending && status != domain.StatusDeactivated {
		return jsonResp(http.StatusBadRequest, errResponse{"status query parameter must be pending or deactivated"})
	}

	result, err := h.useCase.Execute(ctx, input.ListByStatusCommand{ActorID: actorID, Status: status})
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return jsonResp(http.StatusForbidden, errResponse{"admin role required"})
		}
		return jsonResp(http.StatusInternalServerError, errResponse{"internal error"})
	}

	return jsonResp(http.StatusOK, result)
}
