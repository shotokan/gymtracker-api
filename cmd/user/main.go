package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	awsdynamo "github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/shotokan/gymtracker-api/internal/user/adapter/primary"
	"github.com/shotokan/gymtracker-api/internal/user/adapter/secondary/dynamo"
	"github.com/shotokan/gymtracker-api/internal/user/application"
	"github.com/shotokan/gymtracker-api/pkg/bcrypt"
	"github.com/shotokan/gymtracker-api/pkg/uuidgen"
)

func main() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("loading aws config: %v", err)
	}

	repo := dynamo.NewUserRepository(awsdynamo.NewFromConfig(cfg), mustEnv("DYNAMO_TABLE"))
	hasher := bcrypt.NewPasswordHasher(12)
	idGen := uuidgen.New()

	register := application.NewRegister(repo, hasher, idGen)
	getProfile := application.NewGetProfile(repo)
	updateProfile := application.NewUpdateProfile(repo)
	deactivate := application.NewDeactivate(repo)

	router := primary.NewRouter(
		primary.NewRegisterHandler(register),
		primary.NewGetProfileHandler(getProfile),
		primary.NewUpdateProfileHandler(updateProfile),
		primary.NewDeactivateHandler(deactivate),
	)

	lambda.Start(router.Handle)
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required env var %q is not set", key)
	}

	return v
}
