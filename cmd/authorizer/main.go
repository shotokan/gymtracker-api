package main

import (
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/shotokan/gymtracker-api/internal/authorizer/adapter/primary"
	"github.com/shotokan/gymtracker-api/internal/authorizer/adapter/secondary/jwt"
	"github.com/shotokan/gymtracker-api/internal/authorizer/application"
)

func main() {
	validator := jwt.NewTokenValidator(mustEnv("JWT_SECRET"), 24*time.Hour)
	useCase := application.New(validator)
	handler := primary.NewHandler(useCase)

	lambda.Start(handler.Handle)
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required env var %q is not set", key)
	}
	return v
}
