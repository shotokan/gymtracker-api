package main

import (
	"context"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"

	"github.com/shotokan/gymtracker-api/internal/auth/adapter/primary"
	"github.com/shotokan/gymtracker-api/internal/auth/adapter/secondary/bcrypt"
	googleValidator "github.com/shotokan/gymtracker-api/internal/auth/adapter/secondary/google"
	"github.com/shotokan/gymtracker-api/internal/auth/adapter/secondary/jwt"
	"github.com/shotokan/gymtracker-api/internal/auth/application"
	userRepo "github.com/shotokan/gymtracker-api/internal/user/adapter/secondary/dynamo"
)

type uuidGen struct{}

func (g *uuidGen) New() string { return uuid.NewString() }

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

	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	if googleClientID == "" {
		panic("GOOGLE_CLIENT_ID environment variable is required")
	}

	// Initialize shared dependencies
	userRepository := userRepo.NewUserRepository(db, tableName)
	jwtTokenGenerator := jwt.NewTokenGenerator(jwtSecret, 24*time.Hour)
	idGen := &uuidGen{}

	// Initialize Login use case and handler
	passwordHasher := bcrypt.NewPasswordHasher()
	loginUC := application.NewLogin(userRepository, passwordHasher, jwtTokenGenerator)
	loginHandler := primary.NewLoginHandler(loginUC)

	// Initialize Google Auth use case and handler
	googleTokenValidator := googleValidator.NewTokenValidator(googleClientID)
	googleAuthUC := application.NewGoogleAuth(
		googleTokenValidator,
		userRepository,
		idGen,
		jwtTokenGenerator,
	)
	googleAuthHandler := primary.NewGoogleAuthHandler(googleAuthUC)

	// Initialize router
	router := primary.NewRouter(loginHandler, googleAuthHandler)

	// Start Lambda
	lambda.Start(router.Handle)
}
