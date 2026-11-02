package main

import (
	"net/http"

	awsdynamo "github.com/aws/aws-sdk-go-v2/service/dynamodb"

	authprimary "github.com/shotokan/gymtracker-api/internal/auth/adapter/primary"
	authbcrypt "github.com/shotokan/gymtracker-api/internal/auth/adapter/secondary/bcrypt"
	authjwt "github.com/shotokan/gymtracker-api/internal/auth/adapter/secondary/jwt"
	authapp "github.com/shotokan/gymtracker-api/internal/auth/application"
	authzjwt "github.com/shotokan/gymtracker-api/internal/authorizer/adapter/secondary/jwt"
	authzapp "github.com/shotokan/gymtracker-api/internal/authorizer/application"
	authzinput "github.com/shotokan/gymtracker-api/internal/authorizer/port/input"
	"github.com/shotokan/gymtracker-api/internal/httpapi"
	planprimary "github.com/shotokan/gymtracker-api/internal/plan/adapter/primary"
	plandynamo "github.com/shotokan/gymtracker-api/internal/plan/adapter/secondary/dynamo"
	planapp "github.com/shotokan/gymtracker-api/internal/plan/application"
	sessionprimary "github.com/shotokan/gymtracker-api/internal/session/adapter/primary"
	sessiondynamo "github.com/shotokan/gymtracker-api/internal/session/adapter/secondary/dynamo"
	sessionapp "github.com/shotokan/gymtracker-api/internal/session/application"
	userprimary "github.com/shotokan/gymtracker-api/internal/user/adapter/primary"
	userdynamo "github.com/shotokan/gymtracker-api/internal/user/adapter/secondary/dynamo"
	userapp "github.com/shotokan/gymtracker-api/internal/user/application"
	"github.com/shotokan/gymtracker-api/pkg/bcrypt"
	"github.com/shotokan/gymtracker-api/pkg/uuidgen"
)

// passwordCost matches cmd/user, so both deployments hash passwords alike.
const passwordCost = 12

func newAuthorizer(jwtSecret string) authzinput.AuthorizeUseCase {
	return authzapp.New(authzjwt.NewTokenValidator(jwtSecret, tokenTTL))
}

// buildRoutes wires the same routers as the Lambda binaries. Its resources match
// infra/lib/constructs/api-gateway.ts.
func buildRoutes(db *awsdynamo.Client, table, jwtSecret string) []httpapi.Route {
	idGen := uuidgen.New()

	userRepo := userdynamo.NewUserRepository(db, table)
	auth := authprimary.NewRouter(authprimary.NewLoginHandler(authapp.NewLogin(
		userRepo, authbcrypt.NewPasswordHasher(), authjwt.NewTokenGenerator(jwtSecret, tokenTTL),
	)))
	user := userprimary.NewRouter(
		userprimary.NewRegisterHandler(userapp.NewRegister(userRepo, bcrypt.NewPasswordHasher(passwordCost), idGen)),
		userprimary.NewGetProfileHandler(userapp.NewGetProfile(userRepo)),
		userprimary.NewUpdateProfileHandler(userapp.NewUpdateProfile(userRepo)),
		userprimary.NewDeactivateHandler(userapp.NewDeactivate(userRepo)),
		userprimary.NewActivateHandler(userapp.NewActivate(userRepo)),
		userprimary.NewListByStatusHandler(userapp.NewListByStatus(userRepo)),
	)

	planRepo := plandynamo.NewPlanRepository(db, table)
	plan := planprimary.NewRouter(
		planprimary.NewCreatePlanHandler(planapp.NewCreatePlan(planRepo, idGen)),
		planprimary.NewListPlansHandler(planapp.NewListPlans(planRepo)),
		planprimary.NewGetPlanHandler(planapp.NewGetPlan(planRepo)),
		planprimary.NewGetActivePlanHandler(planapp.NewGetActivePlan(planRepo)),
		planprimary.NewUpdatePlanHandler(planapp.NewUpdatePlan(planRepo)),
		planprimary.NewActivatePlanHandler(planapp.NewActivatePlan(planRepo)),
		planprimary.NewDeletePlanHandler(planapp.NewDeletePlan(planRepo)),
		planprimary.NewAddRoutineHandler(planapp.NewAddRoutine(planRepo, idGen)),
		planprimary.NewUpdateRoutineHandler(planapp.NewUpdateRoutine(planRepo)),
		planprimary.NewDeleteRoutineHandler(planapp.NewDeleteRoutine(planRepo)),
		planprimary.NewAddExerciseHandler(planapp.NewAddExerciseUseCase(planRepo, idGen)),
		planprimary.NewListExercisesHandler(planapp.NewListExercisesUseCase(planRepo)),
		planprimary.NewUpdateExerciseHandler(planapp.NewUpdateExerciseUseCase(planRepo)),
		planprimary.NewDeleteExerciseHandler(planapp.NewDeleteExerciseUseCase(planRepo)),
	)

	session := sessionprimary.NewRouter(
		sessionprimary.NewCreateSessionHandler(sessionapp.NewCreateSession(sessiondynamo.NewSessionRepository(db, table), idGen)),
		sessionprimary.NewListSessionsHandler(sessionapp.NewListSessions(sessiondynamo.NewSessionRepository(db, table))),
		sessionprimary.NewGetSessionHandler(sessionapp.NewGetSession(sessiondynamo.NewSessionRepository(db, table))),
	)

	const (
		routine  = "/plans/{planId}/routines/{routineId}"
		exercise = routine + "/exercises/{exerciseId}"
	)

	return []httpapi.Route{
		{Method: http.MethodPost, Resource: "/auth/login", Public: true, Handler: auth.Handle},

		{Method: http.MethodPost, Resource: "/users", Public: true, Handler: user.Handle},
		{Method: http.MethodGet, Resource: "/users", Handler: user.Handle},
		{Method: http.MethodGet, Resource: "/users/{id}", Handler: user.Handle},
		{Method: http.MethodPatch, Resource: "/users/{id}", Handler: user.Handle},
		{Method: http.MethodDelete, Resource: "/users/{id}", Handler: user.Handle},
		{Method: http.MethodPatch, Resource: "/users/{id}/activate", Handler: user.Handle},

		{Method: http.MethodPost, Resource: "/plans", Handler: plan.Handle},
		{Method: http.MethodGet, Resource: "/plans", Handler: plan.Handle},
		{Method: http.MethodGet, Resource: "/plans/active", Handler: plan.Handle},
		{Method: http.MethodGet, Resource: "/plans/{planId}", Handler: plan.Handle},
		{Method: http.MethodPatch, Resource: "/plans/{planId}", Handler: plan.Handle},
		{Method: http.MethodDelete, Resource: "/plans/{planId}", Handler: plan.Handle},
		{Method: http.MethodPatch, Resource: "/plans/{planId}/activate", Handler: plan.Handle},
		{Method: http.MethodPost, Resource: "/plans/{planId}/routines", Handler: plan.Handle},
		{Method: http.MethodPatch, Resource: routine, Handler: plan.Handle},
		{Method: http.MethodDelete, Resource: routine, Handler: plan.Handle},
		{Method: http.MethodPost, Resource: routine + "/exercises", Handler: plan.Handle},
		{Method: http.MethodGet, Resource: routine + "/exercises", Handler: plan.Handle},
		{Method: http.MethodPatch, Resource: exercise, Handler: plan.Handle},
		{Method: http.MethodDelete, Resource: exercise, Handler: plan.Handle},

		{Method: http.MethodPost, Resource: "/sessions", Handler: session.Handle},
		{Method: http.MethodGet, Resource: "/sessions", Handler: session.Handle},
		{Method: http.MethodGet, Resource: "/sessions/{sessionId}", Handler: session.Handle},
	}
}
