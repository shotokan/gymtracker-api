package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	awsdynamo "github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/shotokan/gymtracker-api/internal/session/adapter/primary"
	"github.com/shotokan/gymtracker-api/internal/session/adapter/secondary/dynamo"
	"github.com/shotokan/gymtracker-api/internal/session/application"
	"github.com/shotokan/gymtracker-api/pkg/uuidgen"
)

func main() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("loading aws config: %v", err)
	}

	repo := dynamo.NewSessionRepository(awsdynamo.NewFromConfig(cfg), mustEnv("DYNAMO_TABLE"))
	idGen := uuidgen.New()

	router := primary.NewRouter(
		primary.NewCreateSessionHandler(application.NewCreateSession(repo, idGen)),
		primary.NewListSessionsHandler(application.NewListSessions(repo)),
		primary.NewGetSessionHandler(application.NewGetSession(repo)),
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
