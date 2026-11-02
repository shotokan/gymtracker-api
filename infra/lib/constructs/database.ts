import { RemovalPolicy } from "aws-cdk-lib";
import * as dynamodb from "aws-cdk-lib/aws-dynamodb";
import { Construct } from "constructs";

/**
 * Single Table Design:
 *
 * PK                      SK                              Description
 * USER#<email>            PROFILE                         User profile
 * USER#<userID>           SESSION#<date>#<sessionID>      User sessions
 * USER#<userID>           PLAN#<planID>                   User plans
 *
 * GSI1PK                  GSI1SK                          Description
 * USER_ID#<id>            PROFILE                         Lookup user by ID
 * SESSION#<sessionID>     USER#<userID>                   Lookup session by ID
 * PLAN#<planID>           USER#<userID>                   Lookup plan by ID
 */
export class DatabaseConstruct extends Construct {
  public readonly table: dynamodb.Table;

  constructor(scope: Construct, id: string) {
    super(scope, id);

    this.table = new dynamodb.Table(this, "Table", {
      tableName: "gym-tracker",
      partitionKey: { name: "PK", type: dynamodb.AttributeType.STRING },
      sortKey: { name: "SK", type: dynamodb.AttributeType.STRING },
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
      removalPolicy: RemovalPolicy.RETAIN,
      encryption: dynamodb.TableEncryption.AWS_MANAGED,
      // Replaces deprecated pointInTimeRecovery: true
      pointInTimeRecoverySpecification: {
        pointInTimeRecoveryEnabled: true,
      },
    });

    // GSI1 for user lookup by ID (used in GetProfile, Update, Deactivate)
    this.table.addGlobalSecondaryIndex({
      indexName: "GSI1",
      partitionKey: { name: "GSI1PK", type: dynamodb.AttributeType.STRING },
      sortKey: { name: "GSI1SK", type: dynamodb.AttributeType.STRING },
      projectionType: dynamodb.ProjectionType.ALL,
    });
  }
}
