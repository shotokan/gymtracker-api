// Command dbinit creates the single table in DynamoDB Local. In AWS the table comes
// from Terraform or CDK, so it refuses to run without DYNAMO_ENDPOINT.
package main

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	retryEvery = 2 * time.Second
	timeout    = 2 * time.Minute
)

func main() {
	table := os.Getenv("DYNAMO_TABLE")
	endpoint := os.Getenv("DYNAMO_ENDPOINT")
	if table == "" || endpoint == "" {
		log.Fatal("DYNAMO_TABLE and DYNAMO_ENDPOINT are required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("loading aws config: %v", err)
	}
	db := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) { o.BaseEndpoint = aws.String(endpoint) })

	// DynamoDB Local may still be starting: retry until it answers.
	for {
		_, err = db.CreateTable(ctx, tableDefinition(table))
		var exists *types.ResourceInUseException
		switch {
		case err == nil:
			log.Print("table created")
			return
		case errors.As(err, &exists):
			log.Print("table already exists")
			return
		case ctx.Err() != nil:
			log.Fatalf("creating table: %v", err)
		}
		log.Printf("waiting for DynamoDB: %v", err)
		time.Sleep(retryEvery)
	}
}

// tableDefinition mirrors infra/lib/constructs/database.ts and terraform/modules/dynamodb.
func tableDefinition(table string) *dynamodb.CreateTableInput {
	stringAttr := func(name string) types.AttributeDefinition {
		return types.AttributeDefinition{AttributeName: aws.String(name), AttributeType: types.ScalarAttributeTypeS}
	}
	key := func(hash, sort string) []types.KeySchemaElement {
		return []types.KeySchemaElement{
			{AttributeName: aws.String(hash), KeyType: types.KeyTypeHash},
			{AttributeName: aws.String(sort), KeyType: types.KeyTypeRange},
		}
	}

	return &dynamodb.CreateTableInput{
		TableName:   aws.String(table),
		BillingMode: types.BillingModePayPerRequest,
		AttributeDefinitions: []types.AttributeDefinition{
			stringAttr("PK"), stringAttr("SK"), stringAttr("GSI1PK"), stringAttr("GSI1SK"),
		},
		KeySchema: key("PK", "SK"),
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{{
			IndexName:  aws.String("GSI1"),
			KeySchema:  key("GSI1PK", "GSI1SK"),
			Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
		}},
	}
}
