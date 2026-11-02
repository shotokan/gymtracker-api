package httpapi

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shotokan/gymtracker-api/internal/authorizer/port/input"
)

type fakeAuthorizer struct{}

func (fakeAuthorizer) Execute(_ context.Context, cmd input.AuthorizeCommand) (input.AuthorizeResult, error) {
	switch cmd.Token {
	case "valid":
		return input.AuthorizeResult{UserID: "user-1", Allowed: true}, nil
	case "no-subject":
		return input.AuthorizeResult{Allowed: true}, nil
	default:
		return input.AuthorizeResult{}, errors.New("invalid token")
	}
}

func newTestServer(t *testing.T, handler LambdaHandler) (*httptest.Server, *prometheus.Registry) {
	t.Helper()
	reg := prometheus.NewRegistry()
	routes := []Route{
		{Method: http.MethodPost, Resource: "/auth/login", Public: true, Handler: handler},
		{Method: http.MethodGet, Resource: "/plans/active", Handler: handler},
		{Method: http.MethodGet, Resource: "/plans/{planId}", Handler: handler},
		{Method: http.MethodPatch, Resource: "/plans/{planId}/routines/{routineId}", Handler: handler},
	}
	srv := httptest.NewServer(New(routes, fakeAuthorizer{}, reg, slog.New(slog.NewTextHandler(io.Discard, nil))))
	t.Cleanup(srv.Close)
	return srv, reg
}

func do(t *testing.T, method, url, token, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), method, url, strings.NewReader(body))
	require.NoError(t, err)
	if token != "" {
		req.Header.Set("Authorization", token)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

func okHandler(_ context.Context, _ events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{StatusCode: http.StatusOK}, nil
}

func TestAuthentication(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		path   string
		token  string
		want   int
	}{
		{"ruta pública sin token", http.MethodPost, "/auth/login", "", http.StatusOK},
		{"ruta protegida sin token", http.MethodGet, "/plans/active", "", http.StatusUnauthorized},
		{"esquema distinto de Bearer", http.MethodGet, "/plans/active", "Basic abc", http.StatusForbidden},
		{"token inválido", http.MethodGet, "/plans/active", "Bearer nope", http.StatusForbidden},
		{"token sin usuario", http.MethodGet, "/plans/active", "Bearer no-subject", http.StatusForbidden},
		{"token válido", http.MethodGet, "/plans/active", "Bearer valid", http.StatusOK},
		{"prefijo de stage", http.MethodGet, "/v1/plans/active", "Bearer valid", http.StatusOK},
		{"ruta inexistente", http.MethodGet, "/nope", "Bearer valid", http.StatusNotFound},
		{"método no permitido", http.MethodDelete, "/auth/login", "", http.StatusMethodNotAllowed},
	}

	srv, _ := newTestServer(t, okHandler)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			resp := do(t, tt.method, srv.URL+tt.path, tt.token, "")
			assert.Equal(t, tt.want, resp.StatusCode)
		})
	}
}

func TestRequestTranslation(t *testing.T) {
	t.Parallel()

	var got events.APIGatewayProxyRequest
	srv, _ := newTestServer(t, func(_ context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		got = req
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusTeapot,
			Headers:    map[string]string{"X-From-Handler": "yes"},
			Body:       `{"ok":true}`,
		}, nil
	})

	resp := do(t, http.MethodPatch, srv.URL+"/v1/plans/p-1/routines/r-2?name=fuerza", "Bearer valid", `{"name":"Empuje"}`)

	assert.Equal(t, "/plans/{planId}/routines/{routineId}", got.Resource)
	assert.Equal(t, "/plans/p-1/routines/r-2", got.Path)
	assert.Equal(t, http.MethodPatch, got.HTTPMethod)
	assert.JSONEq(t, `{"name":"Empuje"}`, got.Body)
	assert.Equal(t, map[string]string{"planId": "p-1", "routineId": "r-2"}, got.PathParameters)
	assert.Equal(t, "fuerza", got.QueryStringParameters["name"])
	assert.Equal(t, "user-1", got.RequestContext.Authorizer["principalId"])

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
	assert.Equal(t, "yes", resp.Header.Get("X-From-Handler"))
	assert.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"))
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.JSONEq(t, `{"ok":true}`, string(body))
}

func TestSpecificRouteWinsOverPathParam(t *testing.T) {
	t.Parallel()

	var got string
	srv, _ := newTestServer(t, func(_ context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		got = req.Resource
		return events.APIGatewayProxyResponse{StatusCode: http.StatusOK}, nil
	})

	do(t, http.MethodGet, srv.URL+"/plans/active", "Bearer valid", "")

	assert.Equal(t, "/plans/active", got)
}

func TestPublicRouteHasNoPrincipal(t *testing.T) {
	t.Parallel()

	var got events.APIGatewayProxyRequest
	srv, _ := newTestServer(t, func(_ context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		got = req
		return events.APIGatewayProxyResponse{StatusCode: http.StatusOK}, nil
	})

	// Un token enviado a una ruta pública no debe identificar a nadie.
	do(t, http.MethodPost, srv.URL+"/auth/login", "Bearer valid", "{}")

	assert.Nil(t, got.RequestContext.Authorizer)
}

func TestHandlerErrorIsNotExposed(t *testing.T) {
	t.Parallel()

	srv, _ := newTestServer(t, func(context.Context, events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		return events.APIGatewayProxyResponse{}, errors.New("dynamo: secret table detail")
	})

	resp := do(t, http.MethodPost, srv.URL+"/auth/login", "", "{}")

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.NotContains(t, string(body), "secret table detail")
}

func TestBodyTooLarge(t *testing.T) {
	t.Parallel()

	srv, _ := newTestServer(t, okHandler)

	resp := do(t, http.MethodPost, srv.URL+"/auth/login", "", strings.Repeat("x", maxBodyBytes+1))

	assert.Equal(t, http.StatusRequestEntityTooLarge, resp.StatusCode)
}

func TestMetricsUseResourceNotURL(t *testing.T) {
	t.Parallel()

	srv, reg := newTestServer(t, okHandler)

	do(t, http.MethodGet, srv.URL+"/plans/secret-id-123", "Bearer valid", "")

	expected := `
# HELP gymtracker_http_requests_total HTTP requests served, by method, route and status code.
# TYPE gymtracker_http_requests_total counter
gymtracker_http_requests_total{method="GET",route="/plans/{planId}",status="200"} 1
`
	require.NoError(t, testutil.GatherAndCompare(reg, strings.NewReader(expected), "gymtracker_http_requests_total"))
}

func TestAdmin(t *testing.T) {
	t.Parallel()

	reg := prometheus.NewRegistry()
	srv := httptest.NewServer(NewAdmin(reg, func(context.Context) error { return errors.New("table missing") }))
	t.Cleanup(srv.Close)

	assert.Equal(t, http.StatusOK, do(t, http.MethodGet, srv.URL+"/healthz", "", "").StatusCode)
	assert.Equal(t, http.StatusServiceUnavailable, do(t, http.MethodGet, srv.URL+"/readyz", "", "").StatusCode)
	assert.Equal(t, http.StatusOK, do(t, http.MethodGet, srv.URL+"/metrics", "", "").StatusCode)
}
