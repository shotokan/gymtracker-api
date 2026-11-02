// Package httpapi exposes the Lambda routers over plain HTTP so the same handlers
// can run in a container. It plays the role API Gateway and the JWT authorizer play
// in the serverless deployment.
package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/shotokan/gymtracker-api/internal/authorizer/port/input"
)

const (
	bearerPrefix = "Bearer "
	stagePrefix  = "/v1"
	// maxBodyBytes matches the largest request API Gateway forwards to Lambda.
	maxBodyBytes = 6 << 20
)

var pathParam = regexp.MustCompile(`\{(\w+)\}`)

// LambdaHandler is the signature shared by every Lambda router.
type LambdaHandler func(context.Context, events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

// Route binds a method and an API Gateway style resource to a handler.
type Route struct {
	Method   string
	Resource string
	Public   bool
	Handler  LambdaHandler
}

// New builds the public API handler. Registering metrics more than once on the
// same registerer panics, so each server owns its registry.
func New(routes []Route, authorizer input.AuthorizeUseCase, reg prometheus.Registerer, logger *slog.Logger) http.Handler {
	metrics := newMetrics(reg)
	mux := http.NewServeMux()

	for _, route := range routes {
		handler := serveLambda(route)
		if !route.Public {
			handler = requireToken(authorizer, handler)
		}
		mux.Handle(route.Method+" "+route.Resource, metrics.instrument(route, handler))
	}

	return withStagePrefix(withSecurityHeaders(withAccessLog(logger, mux)))
}

// withStagePrefix accepts the /v1 prefix so clients use the same paths as in API Gateway.
func withStagePrefix(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if path, ok := strings.CutPrefix(r.URL.Path, stagePrefix); ok && strings.HasPrefix(path, "/") {
			r.URL.Path = path
		}
		next.ServeHTTP(w, r)
	})
}

func withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

// withAccessLog logs each request without its body or headers, which carry
// passwords and tokens.
func withAccessLog(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

// requireToken mirrors API Gateway: 401 without a token, 403 when it is not valid.
func requireToken(authorizer input.AuthorizeUseCase, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			writeMessage(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		token, ok := strings.CutPrefix(header, bearerPrefix)
		if !ok {
			writeMessage(w, http.StatusForbidden, "Forbidden")
			return
		}

		result, err := authorizer.Execute(r.Context(), input.AuthorizeCommand{Token: token})
		if err != nil || result.UserID == "" {
			writeMessage(w, http.StatusForbidden, "Forbidden")
			return
		}

		next.ServeHTTP(w, r.WithContext(withPrincipal(r.Context(), result.UserID)))
	})
}

type principalKey struct{}

func withPrincipal(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, principalKey{}, userID)
}

func serveLambda(route Route) http.Handler {
	names := pathParamNames(route.Resource)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodyBytes))
		if err != nil {
			var tooLarge *http.MaxBytesError
			if errors.As(err, &tooLarge) {
				writeMessage(w, http.StatusRequestEntityTooLarge, "request body too large")
				return
			}
			writeMessage(w, http.StatusBadRequest, "invalid request body")
			return
		}

		req := events.APIGatewayProxyRequest{
			Resource:              route.Resource,
			Path:                  r.URL.Path,
			HTTPMethod:            r.Method,
			Body:                  string(body),
			PathParameters:        make(map[string]string, len(names)),
			QueryStringParameters: firstValues(r),
		}
		for _, name := range names {
			req.PathParameters[name] = r.PathValue(name)
		}
		if userID, ok := r.Context().Value(principalKey{}).(string); ok {
			req.RequestContext.Authorizer = map[string]interface{}{"principalId": userID}
		}

		resp, err := route.Handler(r.Context(), req)
		if err != nil {
			writeMessage(w, http.StatusInternalServerError, "internal error")
			return
		}

		for key, value := range resp.Headers {
			w.Header().Set(key, value)
		}
		w.WriteHeader(resp.StatusCode)
		_, _ = io.WriteString(w, resp.Body)
	})
}

func pathParamNames(resource string) []string {
	matches := pathParam.FindAllStringSubmatch(resource, -1)
	names := make([]string, 0, len(matches))
	for _, match := range matches {
		names = append(names, match[1])
	}
	return names
}

func firstValues(r *http.Request) map[string]string {
	query := r.URL.Query()
	values := make(map[string]string, len(query))
	for key := range query {
		values[key] = query.Get(key)
	}
	return values
}

func writeMessage(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": message})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(status int) {
	s.status = status
	s.ResponseWriter.WriteHeader(status)
}

type metrics struct {
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
	inFlight prometheus.Gauge
}

func newMetrics(reg prometheus.Registerer) *metrics {
	m := &metrics{
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "gymtracker_http_requests_total",
			Help: "HTTP requests served, by method, route and status code.",
		}, []string{"method", "route", "status"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "gymtracker_http_request_duration_seconds",
			Help:    "HTTP request latency, by method and route.",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		}, []string{"method", "route"}),
		inFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "gymtracker_http_requests_in_flight",
			Help: "HTTP requests being served right now.",
		}),
	}
	reg.MustRegister(m.requests, m.duration, m.inFlight)
	return m
}

// instrument labels by resource and not by URL, so ids never become label values.
func (m *metrics) instrument(route Route, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.inFlight.Inc()
		defer m.inFlight.Dec()

		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		m.requests.WithLabelValues(route.Method, route.Resource, strconv.Itoa(rec.status)).Inc()
		m.duration.WithLabelValues(route.Method, route.Resource).Observe(time.Since(start).Seconds())
	})
}
