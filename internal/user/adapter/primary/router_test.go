package primary

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/shotokan/gymtracker-api/internal/user/application"
	"github.com/shotokan/gymtracker-api/internal/user/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeRepo es un repositorio en memoria para probar la autorización de punta a punta.
type fakeRepo struct{ users map[string]*domain.User }

func (r *fakeRepo) FindByEmail(_ context.Context, _ string) (*domain.User, error) {
	return nil, domain.ErrNotFound
}

func (r *fakeRepo) FindByID(_ context.Context, id string) (*domain.User, error) {
	u, ok := r.users[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	copied := *u
	return &copied, nil
}

func (r *fakeRepo) Save(_ context.Context, _ *domain.User) error { return nil }

func (r *fakeRepo) Update(_ context.Context, u *domain.User) error {
	r.users[u.ID] = u
	return nil
}

func (r *fakeRepo) FindByStatus(_ context.Context, status string) ([]domain.User, error) {
	var out []domain.User
	for _, u := range r.users {
		if u.Status == status {
			out = append(out, *u)
		}
	}
	return out, nil
}

func newTestRouter(repo *fakeRepo) *Router {
	return NewRouter(
		NewRegisterHandler(application.NewRegister(repo, nil, nil)),
		NewGetProfileHandler(application.NewGetProfile(repo)),
		NewUpdateProfileHandler(application.NewUpdateProfile(repo)),
		NewDeactivateHandler(application.NewDeactivate(repo)),
		NewActivateHandler(application.NewActivate(repo)),
		NewListByStatusHandler(application.NewListByStatus(repo)),
	)
}

func TestRouter_Authorization(t *testing.T) {
	t.Parallel()

	const (
		userResource     = "/users/{id}"
		activateResource = "/users/{id}/activate"
	)

	tests := []struct {
		name     string
		method   string
		resource string
		actor    string
		target   string
		query    map[string]string
		want     int
	}{
		{"ver perfil propio", http.MethodGet, userResource, "alice", "alice", nil, http.StatusOK},
		{"ver perfil ajeno", http.MethodGet, userResource, "alice", "bob", nil, http.StatusForbidden},
		{"admin no ve perfil ajeno", http.MethodGet, userResource, "admin", "bob", nil, http.StatusForbidden},
		{"ver perfil sin principal", http.MethodGet, userResource, "", "bob", nil, http.StatusUnauthorized},

		{"editar perfil propio", http.MethodPatch, userResource, "alice", "alice", nil, http.StatusOK},
		{"editar perfil ajeno", http.MethodPatch, userResource, "alice", "bob", nil, http.StatusForbidden},
		{"editar perfil sin principal", http.MethodPatch, userResource, "", "bob", nil, http.StatusUnauthorized},

		{"desactivar cuenta propia", http.MethodDelete, userResource, "alice", "alice", nil, http.StatusNoContent},
		{"desactivar cuenta ajena", http.MethodDelete, userResource, "alice", "bob", nil, http.StatusForbidden},
		{"usuario no desactiva al admin", http.MethodDelete, userResource, "alice", "admin", nil, http.StatusForbidden},
		{"admin desactiva a otro", http.MethodDelete, userResource, "admin", "bob", nil, http.StatusNoContent},
		{"único admin no se da de baja", http.MethodDelete, userResource, "admin", "admin", nil, http.StatusConflict},
		{"admin desactivado no desactiva", http.MethodDelete, userResource, "exadmin", "bob", nil, http.StatusForbidden},
		{"desactivar sin principal", http.MethodDelete, userResource, "", "bob", nil, http.StatusUnauthorized},

		{"usuario no activa", http.MethodPatch, activateResource, "alice", "pending", nil, http.StatusForbidden},
		{"admin activa", http.MethodPatch, activateResource, "admin", "pending", nil, http.StatusNoContent},
		{"activar sin principal", http.MethodPatch, activateResource, "", "pending", nil, http.StatusUnauthorized},

		{"admin lista pendientes", http.MethodGet, "/users", "admin", "", map[string]string{"status": "pending"}, http.StatusOK},
		{"admin lista dados de baja", http.MethodGet, "/users", "admin", "", map[string]string{"status": "deactivated"}, http.StatusOK},
		{"usuario no lista", http.MethodGet, "/users", "alice", "", map[string]string{"status": "pending"}, http.StatusForbidden},
		{"listar sin principal", http.MethodGet, "/users", "", "", map[string]string{"status": "pending"}, http.StatusUnauthorized},
		{"listar sin status", http.MethodGet, "/users", "admin", "", nil, http.StatusBadRequest},
		{"listar con otro status", http.MethodGet, "/users", "admin", "", map[string]string{"status": "active"}, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &fakeRepo{users: map[string]*domain.User{
				"alice":   {ID: "alice", Role: domain.RoleUser, Status: domain.StatusActive},
				"bob":     {ID: "bob", Role: domain.RoleUser, Status: domain.StatusActive},
				"pending": {ID: "pending", Role: domain.RoleUser, Status: domain.StatusPending},
				"admin":   {ID: "admin", Role: domain.RoleAdmin, Status: domain.StatusActive},
				"exadmin": {ID: "exadmin", Role: domain.RoleAdmin, Status: domain.StatusDeactivated},
			}}

			req := events.APIGatewayProxyRequest{
				HTTPMethod:            tt.method,
				Resource:              tt.resource,
				Body:                  `{"name":"Nuevo"}`,
				PathParameters:        map[string]string{"id": tt.target},
				QueryStringParameters: tt.query,
			}
			if tt.actor != "" {
				req.RequestContext.Authorizer = map[string]interface{}{"principalId": tt.actor}
			}

			resp, err := newTestRouter(repo).Handle(context.Background(), req)

			require.NoError(t, err)
			assert.Equal(t, tt.want, resp.StatusCode)
		})
	}
}

func TestRouter_ListByStatus_OnlyRequestedStatusAndNoPasswordHash(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{users: map[string]*domain.User{
		"admin":   {ID: "admin", Role: domain.RoleAdmin, Status: domain.StatusActive},
		"pending": {ID: "pending", Email: "p@gym.com", PasswordHash: "secret-hash", Role: domain.RoleUser, Status: domain.StatusPending},
		"gone":    {ID: "gone", Email: "g@gym.com", Role: domain.RoleUser, Status: domain.StatusDeactivated},
	}}

	req := events.APIGatewayProxyRequest{
		HTTPMethod:            http.MethodGet,
		Resource:              "/users",
		QueryStringParameters: map[string]string{"status": "pending"},
	}
	req.RequestContext.Authorizer = map[string]interface{}{"principalId": "admin"}

	resp, err := newTestRouter(repo).Handle(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Body, "p@gym.com")
	assert.NotContains(t, resp.Body, "g@gym.com")
	assert.NotContains(t, resp.Body, "secret-hash")
}
