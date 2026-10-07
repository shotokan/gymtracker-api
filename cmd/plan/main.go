package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	awsdynamo "github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/shotokan/gymtracker-api/internal/plan/adapter/primary"
	"github.com/shotokan/gymtracker-api/internal/plan/adapter/secondary/dynamo"
	"github.com/shotokan/gymtracker-api/internal/plan/application"
	"github.com/shotokan/gymtracker-api/pkg/uuidgen"
)

func main() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("loading aws config: %v", err)
	}

	repo := dynamo.NewPlanRepository(awsdynamo.NewFromConfig(cfg), mustEnv("DYNAMO_TABLE"))
	idGen := uuidgen.New()

	router := primary.NewRouter(
		primary.NewCreatePlanHandler(application.NewCreatePlan(repo, idGen)),
		primary.NewListPlansHandler(application.NewListPlans(repo)),
		primary.NewGetPlanHandler(application.NewGetPlan(repo)),
		primary.NewGetActivePlanHandler(application.NewGetActivePlan(repo)),
		primary.NewUpdatePlanHandler(application.NewUpdatePlan(repo)),
		primary.NewActivatePlanHandler(application.NewActivatePlan(repo)),
		primary.NewDeletePlanHandler(application.NewDeletePlan(repo)),
		primary.NewAddRoutineHandler(application.NewAddRoutine(repo, idGen)),
		primary.NewUpdateRoutineHandler(application.NewUpdateRoutine(repo)),
		primary.NewDeleteRoutineHandler(application.NewDeleteRoutine(repo)),
		primary.NewAddExerciseHandler(application.NewAddExerciseUseCase(repo, idGen)),
		primary.NewListExercisesHandler(application.NewListExercisesUseCase(repo)),
		primary.NewUpdateExerciseHandler(application.NewUpdateExerciseUseCase(repo)),
		primary.NewDeleteExerciseHandler(application.NewDeleteExerciseUseCase(repo)),
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
