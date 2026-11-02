// Command server runs the whole API as one HTTP process, for container deployments.
// The Lambda binaries in the other cmd directories remain the serverless entry points.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	awsdynamo "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"golang.org/x/sync/errgroup"

	"github.com/shotokan/gymtracker-api/internal/httpapi"
)

const (
	tokenTTL        = 24 * time.Hour
	shutdownTimeout = 20 * time.Second
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	table, err := requireEnv("DYNAMO_TABLE")
	if err != nil {
		return err
	}
	jwtSecret, err := requireEnv("JWT_SECRET")
	if err != nil {
		return err
	}

	db, err := newDynamoClient(ctx)
	if err != nil {
		return err
	}

	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewGoCollector(), collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	api := httpapi.New(buildRoutes(db, table, jwtSecret), newAuthorizer(jwtSecret), reg, logger)
	admin := httpapi.NewAdmin(reg, func(ctx context.Context) error {
		_, err := db.DescribeTable(ctx, &awsdynamo.DescribeTableInput{TableName: aws.String(table)})
		return err //nolint:wrapcheck // Only its presence decides readiness
	})

	apiAddr := ":" + envOr("PORT", "8080")
	adminAddr := ":" + envOr("ADMIN_PORT", "9090")
	logger.Info("listening", "api", apiAddr, "admin", adminAddr)

	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() error { return serve(ctx, apiAddr, api) })
	group.Go(func() error { return serve(ctx, adminAddr, admin) })
	return group.Wait() //nolint:wrapcheck // serve already adds the address
}

// serve drains in-flight requests when ctx ends, so a scale-down or a rollout does
// not cut requests short.
func serve(ctx context.Context, addr string, handler http.Handler) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		BaseContext:       func(net.Listener) context.Context { return context.WithoutCancel(ctx) },
	}

	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()

	select {
	case err := <-errCh:
		return fmt.Errorf("serving %s: %w", addr, err)
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutting down %s: %w", addr, err)
	}
	if err := <-errCh; !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serving %s: %w", addr, err)
	}
	return nil
}

// newDynamoClient honors DYNAMO_ENDPOINT so local environments can use DynamoDB Local.
func newDynamoClient(ctx context.Context) (*awsdynamo.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading aws config: %w", err)
	}
	return awsdynamo.NewFromConfig(cfg, func(o *awsdynamo.Options) {
		if endpoint := os.Getenv("DYNAMO_ENDPOINT"); endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
	}), nil
}

func requireEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("required env var %q is not set", key)
	}
	return value, nil
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
