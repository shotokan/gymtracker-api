package primary

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

// Router dispatches API Gateway events to the correct handler.
type Router struct {
	// Plan handlers
	createPlan    *CreatePlanHandler
	listPlans     *ListPlansHandler
	getPlan       *GetPlanHandler
	getActivePlan *GetActivePlanHandler
	updatePlan    *UpdatePlanHandler
	activatePlan  *ActivatePlanHandler
	deletePlan    *DeletePlanHandler
	// Routine handlers
	addRoutine    *AddRoutineHandler
	updateRoutine *UpdateRoutineHandler
	deleteRoutine *DeleteRoutineHandler
	// Exercise handlers
	addExercise    *AddExerciseHandler
	listExercises  *ListExercisesHandler
	updateExercise *UpdateExerciseHandler
	deleteExercise *DeleteExerciseHandler
}

// NewRouter constructs the router with all handlers.
func NewRouter(
	createPlan *CreatePlanHandler,
	listPlans *ListPlansHandler,
	getPlan *GetPlanHandler,
	getActivePlan *GetActivePlanHandler,
	updatePlan *UpdatePlanHandler,
	activatePlan *ActivatePlanHandler,
	deletePlan *DeletePlanHandler,
	addRoutine *AddRoutineHandler,
	updateRoutine *UpdateRoutineHandler,
	deleteRoutine *DeleteRoutineHandler,
	addExercise *AddExerciseHandler,
	listExercises *ListExercisesHandler,
	updateExercise *UpdateExerciseHandler,
	deleteExercise *DeleteExerciseHandler,
) *Router {
	return &Router{
		createPlan:     createPlan,
		listPlans:      listPlans,
		getPlan:        getPlan,
		getActivePlan:  getActivePlan,
		updatePlan:     updatePlan,
		activatePlan:   activatePlan,
		deletePlan:     deletePlan,
		addRoutine:     addRoutine,
		updateRoutine:  updateRoutine,
		deleteRoutine:  deleteRoutine,
		addExercise:    addExercise,
		listExercises:  listExercises,
		updateExercise: updateExercise,
		deleteExercise: deleteExercise,
	}
}

// Handle dispatches the incoming request to the correct handler.
func (r *Router) Handle(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID := extractUserID(req)
	if userID == "" {
		return jsonResp(http.StatusUnauthorized, errResponse{"unauthorized"})
	}

	method := req.HTTPMethod
	resource := req.Resource

	switch {
	// ── Plans ────────────────────────────────────────────────────────────────
	case method == "POST" && resource == "/plans":
		return r.createPlan.Handle(ctx, req, userID)
	case method == "GET" && resource == "/plans":
		return r.listPlans.Handle(ctx, req, userID)
	case method == "GET" && resource == "/plans/active":
		return r.getActivePlan.Handle(ctx, req, userID)
	case method == "GET" && resource == "/plans/{planId}":
		return r.getPlan.Handle(ctx, req, userID)
	case method == "PATCH" && resource == "/plans/{planId}":
		return r.updatePlan.Handle(ctx, req, userID)
	case method == "DELETE" && resource == "/plans/{planId}":
		return r.deletePlan.Handle(ctx, req, userID)
	case method == "PATCH" && resource == "/plans/{planId}/activate":
		return r.activatePlan.Handle(ctx, req, userID)

	// ── Routines ─────────────────────────────────────────────────────────────
	case method == "POST" && resource == "/plans/{planId}/routines":
		return r.addRoutine.Handle(ctx, req, userID)
	case method == "PATCH" && resource == "/plans/{planId}/routines/{routineId}":
		return r.updateRoutine.Handle(ctx, req, userID)
	case method == "DELETE" && resource == "/plans/{planId}/routines/{routineId}":
		return r.deleteRoutine.Handle(ctx, req, userID)

	// ── Exercises ─────────────────────────────────────────────────────────────
	case method == "POST" && resource == "/plans/{planId}/routines/{routineId}/exercises":
		return r.addExercise.Handle(ctx, req, userID)
	case method == "GET" && resource == "/plans/{planId}/routines/{routineId}/exercises":
		return r.listExercises.Handle(ctx, req, userID)
	case method == "PATCH" && resource == "/plans/{planId}/routines/{routineId}/exercises/{exerciseId}":
		return r.updateExercise.Handle(ctx, req, userID)
	case method == "DELETE" && resource == "/plans/{planId}/routines/{routineId}/exercises/{exerciseId}":
		return r.deleteExercise.Handle(ctx, req, userID)

	default:
		return jsonResp(http.StatusNotFound, errResponse{"route not found"})
	}
}
