package main

import (
	"context"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/shotokan/gymtracker-api/internal/auth/adapter/primary"
	"github.com/shotokan/gymtracker-api/internal/auth/adapter/secondary/bcrypt"
	"github.com/shotokan/gymtracker-api/internal/auth/adapter/secondary/jwt"
	"github.com/shotokan/gymtracker-api/internal/auth/application"
	userRepo "github.com/shotokan/gymtracker-api/internal/user/adapter/secondary/dynamo"
)

func main() {
	ctx := context.Background()

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic("unable to load SDK config: " + err.Error())
	}

	db := dynamodb.NewFromConfig(cfg)
	tableName := os.Getenv("TABLE_NAME")
	if tableName == "" {
		panic("TABLE_NAME environment variable is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		panic("JWT_SECRET environment variable is required")
	}

	// Initialize shared dependencies
	userRepository := userRepo.NewUserRepository(db, tableName)
	jwtTokenGenerator := jwt.NewTokenGenerator(jwtSecret, 24*time.Hour)

	// Initialize Login use case and handler
	passwordHasher := bcrypt.NewPasswordHasher()
	loginUC := application.NewLogin(userRepository, passwordHasher, jwtTokenGenerator)
	loginHandler := primary.NewLoginHandler(loginUC)

	// Initialize router
	router := primary.NewRouter(loginHandler)

	// Start Lambda
	lambda.Start(router.Handle)
}
