import { Duration, CfnOutput } from "aws-cdk-lib";
import * as apigateway from "aws-cdk-lib/aws-apigateway";
import * as lambda from "aws-cdk-lib/aws-lambda";
import * as logs from "aws-cdk-lib/aws-logs";
import { Construct } from "constructs";

interface ApiGatewayProps {
  authFn: lambda.Function;
  authorizerFn: lambda.Function;
  userFn: lambda.Function;
  planFn: lambda.Function;
  sessionFn: lambda.Function;
}

export class ApiGatewayConstruct extends Construct {
  public readonly api: apigateway.RestApi;

  constructor(scope: Construct, id: string, props: ApiGatewayProps) {
    super(scope, id);

    const accessLogs = new logs.LogGroup(this, "AccessLogs", {
      retention: logs.RetentionDays.ONE_WEEK,
    });

    this.api = new apigateway.RestApi(this, "Api", {
      restApiName: "gym-tracker-api",
      description: "GymTracker REST API",
      deployOptions: {
        stageName: "v1",
        accessLogDestination: new apigateway.LogGroupLogDestination(accessLogs),
        accessLogFormat: apigateway.AccessLogFormat.jsonWithStandardFields(),
        loggingLevel: apigateway.MethodLoggingLevel.ERROR,
        metricsEnabled: true,
        tracingEnabled: true, // AWS X-Ray
      },
      defaultCorsPreflightOptions: {
        allowOrigins: apigateway.Cors.ALL_ORIGINS,
        allowMethods: apigateway.Cors.ALL_METHODS,
        allowHeaders: ["Content-Type", "Authorization"],
        maxAge: Duration.days(1),
      },
    });

    // Token Authorizer: cachea la política 5 min → menos invocaciones = menos costo
    const authorizer = new apigateway.TokenAuthorizer(this, "JwtAuthorizer", {
      handler: props.authorizerFn,
      identitySource: "method.request.header.Authorization",
      resultsCacheTtl: Duration.minutes(5),
      authorizerName: "JwtAuthorizer",
    });

    const authMethod: apigateway.MethodOptions = {
      authorizer,
      authorizationType: apigateway.AuthorizationType.CUSTOM,
    };

    // ── /auth ──────────────────────────────────────────────────────────────
    const authInt = new apigateway.LambdaIntegration(props.authFn);
    const auth = this.api.root.addResource("auth");
    auth.addResource("login").addMethod("POST", authInt, {
      operationName: "Login",
    });

    // ── /users ─────────────────────────────────────────────────────────────
    const users = this.api.root.addResource("users");

    // POST /users → registro público (sin authorizer)
    users.addMethod("POST", new apigateway.LambdaIntegration(props.userFn), {
      operationName: "RegisterUser",
    });

    // GET /users?status=pending|deactivated → solo administradores (validado en el caso de uso)
    users.addMethod("GET", new apigateway.LambdaIntegration(props.userFn), {
      ...authMethod,
      operationName: "ListUsersByStatus",
    });

    // /users/{id} → protegido con JWT
    const user = users.addResource("{id}");
    user.addMethod("GET", new apigateway.LambdaIntegration(props.userFn), {
      ...authMethod,
      operationName: "GetProfile",
    });
    user.addMethod("PATCH", new apigateway.LambdaIntegration(props.userFn), {
      ...authMethod,
      operationName: "UpdateProfile",
    });
    user.addMethod("DELETE", new apigateway.LambdaIntegration(props.userFn), {
      ...authMethod,
      operationName: "DeactivateUser",
    });

    // /users/{id}/activate → solo administradores (validado en el caso de uso)
    user
      .addResource("activate")
      .addMethod("PATCH", new apigateway.LambdaIntegration(props.userFn), {
        ...authMethod,
        operationName: "ActivateUser",
      });

    // ── /plans ─────────────────────────────────────────────────────────────
    const planInt = new apigateway.LambdaIntegration(props.planFn);

    const plans = this.api.root.addResource("plans");
    const activePlan = plans.addResource("active");
    const planId = plans.addResource("{planId}");
    const activate = planId.addResource("activate");
    const routines = planId.addResource("routines");
    const routineId = routines.addResource("{routineId}");

    plans.addMethod("POST", planInt, {
      ...authMethod,
      operationName: "CreatePlan",
    });
    plans.addMethod("GET", planInt, {
      ...authMethod,
      operationName: "ListPlans",
    });

    activePlan.addMethod("GET", planInt, {
      ...authMethod,
      operationName: "GetActivePlan",
    });

    planId.addMethod("GET", planInt, {
      ...authMethod,
      operationName: "GetPlan",
    });
    planId.addMethod("PATCH", planInt, {
      ...authMethod,
      operationName: "UpdatePlan",
    });
    planId.addMethod("DELETE", planInt, {
      ...authMethod,
      operationName: "DeletePlan",
    });

    activate.addMethod("PATCH", planInt, {
      ...authMethod,
      operationName: "ActivatePlan",
    });

    routines.addMethod("POST", planInt, {
      ...authMethod,
      operationName: "AddRoutine",
    });

    routineId.addMethod("PATCH", planInt, {
      ...authMethod,
      operationName: "UpdateRoutine",
    });
    routineId.addMethod("DELETE", planInt, {
      ...authMethod,
      operationName: "DeleteRoutine",
    });

    // ── /plans/{planId}/routines/{routineId}/exercises ─────────────────────

    const exercises = routineId.addResource("exercises");
    const exerciseId = exercises.addResource("{exerciseId}");

    exercises.addMethod("POST", planInt, {
      ...authMethod,
      operationName: "AddExercise",
    });
    exercises.addMethod("GET", planInt, {
      ...authMethod,
      operationName: "ListExercises",
    });
    exerciseId.addMethod("PATCH", planInt, {
      ...authMethod,
      operationName: "UpdateExercise",
    });
    exerciseId.addMethod("DELETE", planInt, {
      ...authMethod,
      operationName: "DeleteExercise",
    });

    // ── /sessions ──────────────────────────────────────────────────────────
    const sessionInt = new apigateway.LambdaIntegration(props.sessionFn);
    const sessions = this.api.root.addResource("sessions");
    const sessionId = sessions.addResource("{sessionId}");

    sessions.addMethod("POST", sessionInt, {
      ...authMethod,
      operationName: "CreateSession",
    });
    sessions.addMethod("GET", sessionInt, {
      ...authMethod,
      operationName: "ListSessions",
    });
    sessionId.addMethod("GET", sessionInt, {
      ...authMethod,
      operationName: "GetSession",
    });

    // Output de la URL base del API
    new CfnOutput(scope, "ApiUrl", {
      value: this.api.url,
      description: "API Gateway URL — úsala como VITE_API_URL en el frontend",
      exportName: "GymTrackerApiUrl",
    });
  }
}
