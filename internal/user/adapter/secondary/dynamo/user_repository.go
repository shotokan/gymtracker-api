package dynamo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/shotokan/gymtracker-api/internal/user/domain"
)

// client defines the minimum DynamoDB operations we need.
// Keeping the interface narrow makes mocking in tests straightforward.
type client interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	// Query is required for GSI lookups — GetItem only works with the main table key.
	Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	// Scan is required to list users by status: no index covers that attribute.
	Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
}

type record struct {
	PK           string `dynamodbav:"PK"`
	SK           string `dynamodbav:"SK"`
	GSI1PK       string `dynamodbav:"GSI1PK"`
	GSI1SK       string `dynamodbav:"GSI1SK"`
	ID           string `dynamodbav:"id"`
	Email        string `dynamodbav:"email"`
	PasswordHash string `dynamodbav:"password_hash"`
	Name         string `dynamodbav:"name"`
	Role         string `dynamodbav:"role,omitempty"`
	Status       string `dynamodbav:"status,omitempty"`
	// Active is legacy: it is read from items that have no status yet and never written.
	Active    bool   `dynamodbav:"active,omitempty"`
	CreatedAt string `dynamodbav:"created_at"`
	UpdatedAt string `dynamodbav:"updated_at"`
}

// UserRepository implements output.UserRepository using DynamoDB.
type UserRepository struct {
	db        client
	tableName string
}

// NewUserRepository constructs the repository with the given DynamoDB client and table name.
func NewUserRepository(db client, tableName string) *UserRepository {
	return &UserRepository{db: db, tableName: tableName}
}

// FindByEmail looks up a user using the main table key: PK = "USER#<email>", SK = "PROFILE".
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	out, err := r.db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: emailPK(email)},
			"SK": &types.AttributeValueMemberS{Value: "PROFILE"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("dynamo get item: %w", err)
	}
	if out.Item == nil {
		return nil, domain.ErrNotFound
	}
	return unmarshal(out.Item)
}

// FindByID looks up a user via the GSI1 index: GSI1PK = "USER_ID#<id>", GSI1SK = "PROFILE".
// GetItem cannot be used here because it only works with the main table key (PK + SK).
// Query against the GSI is the correct approach.
func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	out, err := r.db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("GSI1"),
		KeyConditionExpression: aws.String("GSI1PK = :pk AND GSI1SK = :sk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: idPK(id)},
			":sk": &types.AttributeValueMemberS{Value: "PROFILE"},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, fmt.Errorf("dynamo query by id: %w", err)
	}
	if len(out.Items) == 0 {
		return nil, domain.ErrNotFound
	}
	return unmarshal(out.Items[0])
}

// FindByStatus returns every user with the given status. It scans the whole single table, so its
// cost grows with plans and sessions too. Each page holds at most 1 MB, hence the loop.
func (r *UserRepository) FindByStatus(ctx context.Context, status string) ([]domain.User, error) {
	filter := "SK = :sk AND #status = :status"
	names := map[string]string{"#status": "status"}
	values := map[string]types.AttributeValue{
		":sk":     &types.AttributeValueMemberS{Value: "PROFILE"},
		":status": &types.AttributeValueMemberS{Value: status},
	}
	// Legacy items have no status, and none of them can be pending (see statusOf).
	if status != domain.StatusPending {
		filter = "SK = :sk AND (#status = :status OR (attribute_not_exists(#status) AND #active = :legacy))"
		names["#active"] = "active"
		values[":legacy"] = &types.AttributeValueMemberBOOL{Value: status == domain.StatusActive}
	}

	users := []domain.User{}
	var startKey map[string]types.AttributeValue

	for {
		out, err := r.db.Scan(ctx, &dynamodb.ScanInput{
			TableName:                 aws.String(r.tableName),
			FilterExpression:          aws.String(filter),
			ExpressionAttributeNames:  names,
			ExpressionAttributeValues: values,
			ExclusiveStartKey:         startKey,
		})
		if err != nil {
			return nil, fmt.Errorf("dynamo scan users by status: %w", err)
		}

		for _, item := range out.Items {
			u, err := unmarshal(item)
			if err != nil {
				return nil, err
			}
			users = append(users, *u)
		}

		if len(out.LastEvaluatedKey) == 0 {
			return users, nil
		}
		startKey = out.LastEvaluatedKey
	}
}

// Save persists a new user. Uses a condition expression to prevent overwriting existing users.
func (r *UserRepository) Save(ctx context.Context, u *domain.User) error {
	item, err := attributevalue.MarshalMap(toRecord(u))
	if err != nil {
		return fmt.Errorf("marshaling user: %w", err)
	}

	_, err = r.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	})

	var condErr *types.ConditionalCheckFailedException
	if errors.As(err, &condErr) {
		return domain.ErrAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("dynamo put item: %w", err)
	}
	return nil
}

// Update modifies the name, status and updated_at of an existing user, and drops the legacy active attribute.
func (r *UserRepository) Update(ctx context.Context, u *domain.User) error {
	_, err := r.db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: emailPK(u.Email)},
			"SK": &types.AttributeValueMemberS{Value: "PROFILE"},
		},
		UpdateExpression: aws.String("SET #n = :name, #status = :status, updated_at = :updatedAt REMOVE #active"),
		ExpressionAttributeNames: map[string]string{
			"#n":      "name",   // "name" is a reserved word in DynamoDB
			"#status": "status", // "status" is a reserved word in DynamoDB
			"#active": "active",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":name":      &types.AttributeValueMemberS{Value: u.Name},
			":status":    &types.AttributeValueMemberS{Value: u.Status},
			":updatedAt": &types.AttributeValueMemberS{Value: u.UpdatedAt.Format(time.RFC3339)},
		},
		ConditionExpression: aws.String("attribute_exists(PK)"),
	})
	if err != nil {
		var condErr *types.ConditionalCheckFailedException
		if errors.As(err, &condErr) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("dynamo update item: %w", err)
	}
	return nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func emailPK(email string) string { return "USER#" + email }

func idPK(id string) string { return "USER_ID#" + id }

func toRecord(u *domain.User) record {
	return record{
		PK:           emailPK(u.Email),
		SK:           "PROFILE",
		GSI1PK:       idPK(u.ID),
		GSI1SK:       "PROFILE",
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Name:         u.Name,
		Role:         u.Role,
		Status:       u.Status,
		CreatedAt:    u.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    u.UpdatedAt.Format(time.RFC3339),
	}
}

// statusOf resolves the status of a record. Items written before the status attribute existed only
// have active, and back then an inactive account could only mean a deactivated one.
func statusOf(rec record) string {
	switch {
	case rec.Status != "":
		return rec.Status
	case rec.Active:
		return domain.StatusActive
	default:
		return domain.StatusDeactivated
	}
}

func unmarshal(item map[string]types.AttributeValue) (*domain.User, error) {
	var rec record
	if err := attributevalue.UnmarshalMap(item, &rec); err != nil {
		return nil, fmt.Errorf("unmarshaling user: %w", err)
	}
	createdAt, _ := time.Parse(time.RFC3339, rec.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, rec.UpdatedAt)
	return &domain.User{
		ID:           rec.ID,
		Email:        rec.Email,
		PasswordHash: rec.PasswordHash,
		Name:         rec.Name,
		Role:         rec.Role,
		Status:       statusOf(rec),
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}, nil
}
