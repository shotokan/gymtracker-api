package primary

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

// Router despacha las peticiones de API Gateway al handler correcto.
// Evita librerías externas de routing — puro stdlib.
type Router struct {
	register      *RegisterHandler
	getProfile    *GetProfileHandler
	updateProfile *UpdateProfileHandler
	deactivate    *DeactivateHandler
	activate      *ActivateHandler
	listByStatus  *ListByStatusHandler
}

// NewRouter construye el router con los handlers necesarios para cada ruta.
func NewRouter(
	reg *RegisterHandler,
	get *GetProfileHandler,
	upd *UpdateProfileHandler,
	deac *DeactivateHandler,
	act *ActivateHandler,
	list *ListByStatusHandler,
) *Router {
	return &Router{
		register:      reg,
		getProfile:    get,
		updateProfile: upd,
		deactivate:    deac,
		activate:      act,
		listByStatus:  list,
	}
}

// Handle es el único punto de entrada de la Lambda.
// API Gateway envía el método HTTP y el recurso en el evento.
func (r *Router) Handle(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch {
	case req.HTTPMethod == http.MethodPost && req.Resource == "/users":
		return r.register.Handle(ctx, req)

	case req.HTTPMethod == http.MethodGet && req.Resource == "/users":
		return r.listByStatus.Handle(ctx, req)

	case req.HTTPMethod == http.MethodGet && req.Resource == "/users/{id}":
		return r.getProfile.Handle(ctx, req)

	case req.HTTPMethod == http.MethodPatch && req.Resource == "/users/{id}":
		return r.updateProfile.Handle(ctx, req)

	case req.HTTPMethod == http.MethodDelete && req.Resource == "/users/{id}":
		return r.deactivate.Handle(ctx, req)

	case req.HTTPMethod == http.MethodPatch && req.Resource == "/users/{id}/activate":
		return r.activate.Handle(ctx, req)

	default:
		return jsonResp(http.StatusNotFound, errResponse{Message: "route not found"})
	}
}
