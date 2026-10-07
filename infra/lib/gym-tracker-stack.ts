import { Stack, StackProps, CfnOutput } from "aws-cdk-lib";
import { Construct } from "constructs";

import { DatabaseConstruct } from "./constructs/database";
import { LambdasConstruct } from "./constructs/lambdas";
import { ApiGatewayConstruct } from "./constructs/api-gateway";

export class GymTrackerStack extends Stack {
  constructor(scope: Construct, id: string, props?: StackProps) {
    super(scope, id, props);

    const db = new DatabaseConstruct(this, "Database");

    const lambdas = new LambdasConstruct(this, "Lambdas", {
      table: db.table,
    });

    new ApiGatewayConstruct(this, "ApiGateway", {
      authFn: lambdas.authFn,
      authorizerFn: lambdas.authorizerFn,
      userFn: lambdas.userFn,
      planFn: lambdas.planFn,
      sessionFn: lambdas.sessionFn,
    });

    new CfnOutput(this, "DynamoTable", {
      value: db.table.tableName,
      description: "Nombre de la tabla DynamoDB",
    });
  }
}
