import * as path from "path";
import { Duration, RemovalPolicy } from "aws-cdk-lib";
import * as dynamodb from "aws-cdk-lib/aws-dynamodb";
import * as lambda from "aws-cdk-lib/aws-lambda";
import * as logs from "aws-cdk-lib/aws-logs";
import * as ssm from "aws-cdk-lib/aws-ssm";
import { Construct } from "constructs";

interface LambdasProps {
  table: dynamodb.Table;
}

export class LambdasConstruct extends Construct {
  public readonly authFn: lambda.Function;
  public readonly authorizerFn: lambda.Function;
  public readonly userFn: lambda.Function;
  public readonly planFn: lambda.Function;
  public readonly sessionFn: lambda.Function;

  // Repo root: infra/lib/constructs → ../../..
  private static readonly REPO_ROOT = path.resolve(__dirname, "../../..");

  constructor(scope: Construct, id: string, props: LambdasProps) {
    super(scope, id);

    // JWT_SECRET stored in SSM Parameter Store.
    // Must be created manually BEFORE the first deploy via: make secret
    const jwtSecret = ssm.StringParameter.valueForStringParameter(
      this,
      "/gym-tracker/jwt-secret",
    );

    // Base config shared across all Lambda functions.
    // Defined without Partial<> so TypeScript knows all properties are concrete
    // and not undefined — required by the lambda.Function constructor.
    const base = {
      runtime: lambda.Runtime.PROVIDED_AL2023,
      architecture: lambda.Architecture.ARM_64,
      handler: "bootstrap",
      timeout: Duration.seconds(15),
      memorySize: 128,
    };

    // Helper: create a CloudWatch log group for a Lambda.
    // Replaces deprecated logRetention property.
    const logGroup = (fnName: string) =>
      new logs.LogGroup(this, `${fnName}Logs`, {
        logGroupName: `/aws/lambda/gym-tracker-${fnName}`,
        retention: logs.RetentionDays.ONE_WEEK,
        removalPolicy: RemovalPolicy.DESTROY,
      });

    // ── Auth (Login + Google OAuth) ────────────────────────────────────────
    // Unified Lambda that handles both /auth/login and /auth/google endpoints
    // Requires GOOGLE_CLIENT_ID to be set in SSM Parameter Store
    const googleClientId = ssm.StringParameter.valueForStringParameter(
      this,
      "/gym-tracker/google-client-id",
    );
    this.authFn = new lambda.Function(this, "AuthFn", {
      ...base,
      functionName: "gym-tracker-auth",
      description: "Autentica usuarios (email/password y Google OAuth)",
      code: this.goCode("./cmd/auth"),
      logGroup: logGroup("auth"),
      environment: {
        TABLE_NAME: props.table.tableName,
        GOOGLE_CLIENT_ID: googleClientId,
        JWT_SECRET: jwtSecret,
      },
    });
    // Read/Write: may create new users (Google OAuth) or just read (login)
    props.table.grantReadWriteData(this.authFn);

    // ── Authorizer ─────────────────────────────────────────────────────────
    this.authorizerFn = new lambda.Function(this, "AuthorizerFn", {
      ...base,
      functionName: "gym-tracker-authorizer",
      description: "Valida JWT para API Gateway",
      code: this.goCode("./cmd/authorizer"),
      logGroup: logGroup("authorizer"),
      environment: {
        JWT_SECRET: jwtSecret,
      },
    });

    // ── User ───────────────────────────────────────────────────────────────
    this.userFn = new lambda.Function(this, "UserFn", {
      ...base,
      functionName: "gym-tracker-user",
      description: "CRUD de usuarios (register, profile, update, deactivate)",
      code: this.goCode("./cmd/user"),
      logGroup: logGroup("user"),
      environment: {
        DYNAMO_TABLE: props.table.tableName,
      },
    });
    props.table.grantReadWriteData(this.userFn);
    // ── Plan ───────────────────────────────────────────────────────────────
    this.planFn = new lambda.Function(this, "PlanFn", {
      ...base,
      functionName: "gym-tracker-plan",
      description: "CRUD de planes y rutinas",
      code: this.goCode("./cmd/plan"),
      logGroup: logGroup("plan"),
      environment: {
        DYNAMO_TABLE: props.table.tableName,
      },
    });
    props.table.grantReadWriteData(this.planFn);

    // ── Session ────────────────────────────────────────────────────────────
    this.sessionFn = new lambda.Function(this, "SessionFn", {
      ...base,
      functionName: "gym-tracker-session",
      description: "Workout session tracking",
      code: this.goCode("./cmd/session"),
      logGroup: logGroup("session"),
      environment: {
        DYNAMO_TABLE: props.table.tableName,
      },
    });
    props.table.grantReadWriteData(this.sessionFn);
  }

  /**
   * Returns the Lambda code asset from a pre-compiled binary.
   * Run `make build` before any CDK command to compile the binaries.
   *
   * Alternative with Docker (for CI/CD):
   *   return lambda.Code.fromAsset(LambdasConstruct.REPO_ROOT, {
   *     bundling: {
   *       image: lambda.Runtime.PROVIDED_AL2023.bundlingImage,
   *       command: [
   *         "bash", "-c",
   *         `GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -ldflags="-s -w" -o /asset-output/bootstrap ${cmdPath}`,
   *       ],
   *       environment: { GOCACHE: "/tmp/go-cache", GOMODCACHE: "/tmp/go-mod-cache" },
   *     },
   *   });
   */
  private goCode(cmdPath: string): lambda.AssetCode {
    const binName = cmdPath.split("/").pop()!; // 'login' | 'authorizer' | 'user'
    return lambda.Code.fromAsset(
      path.join(LambdasConstruct.REPO_ROOT, "bin", binName),
    );
  }
}
