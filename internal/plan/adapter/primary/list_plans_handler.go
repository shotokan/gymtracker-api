// list_plans_handler.go
package primary

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/shotokan/gymtracker-api/internal/plan/port/input"
)

// ListPlansHandler adapta API Gateway al caso de uso de listar planes.
type ListPlansHandler struct{ uc input.ListPlansUseCase }

// NewListPlansHandler construye el handler para listar planes.
func NewListPlansHandler(uc input.ListPlansUseCase) *ListPlansHandler {
	return &ListPlansHandler{uc: uc}
}

// Handle procesa la solicitud de listar planes, aplicando un filtro opcional por nombre y delegando al caso de uso.
func (h *ListPlansHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest, userID string) (events.APIGatewayProxyResponse, error) {
	nameFilter := req.QueryStringParameters["name"]

	results, err := h.uc.Execute(ctx, input.ListPlansCommand{UserID: userID, NameFilter: nameFilter})
	if err != nil {
		return jsonResp(http.StatusInternalServerError, errResponse{"internal error"})
	}

	return jsonResp(http.StatusOK, results)
}
