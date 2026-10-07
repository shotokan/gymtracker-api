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
}

type record struct {
	PK           string `dynamodbav:"PK"`
	SK           string `dynamodbav:"SK"`
	GSI1PK       string `dynamodbav:"GSI1PK"`
	GSI1SK       string `dynamodbav:"GSI1SK"`
	GSI2PK       string `dynamodbav:"GSI2PK,omitempty"` // For Google ID lookup
	GSI2SK       string `dynamodbav:"GSI2SK,omitempty"`
	ID           string `dynamodbav:"id"`
	Email        string `dynamodbav:"email"`
	PasswordHash string `dynamodbav:"password_hash"`
	GoogleID     string `dynamodbav:"google_id,omitempty"`
	Name         string `dynamodbav:"name"`
	Active       bool   `dynamodbav:"active"`
	CreatedAt    string `dynamodbav:"created_at"`
	UpdatedAt    string `dynamodbav:"updated_at"`
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

// FindByGoogleID looks up a user via the GSI2 index: GSI2PK = "GOOGLE#<googleID>", GSI2SK = "PROFILE".
func (r *UserRepository) FindByGoogleID(ctx context.Context, googleID string) (*domain.User, error) {
	out, err := r.db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("GSI2"),
		KeyConditionExpression: aws.String("GSI2PK = :pk AND GSI2SK = :sk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: googleIDPK(googleID)},
			":sk": &types.AttributeValueMemberS{Value: "PROFILE"},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, fmt.Errorf("dynamo query by google id: %w", err)
	}
	if len(out.Items) == 0 {
		return nil, domain.ErrNotFound
	}
	return unmarshal(out.Items[0])
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

// Update modifies the name, active status and updated_at of an existing user.
func (r *UserRepository) Update(ctx context.Context, u *domain.User) error {
	_, err := r.db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: emailPK(u.Email)},
			"SK": &types.AttributeValueMemberS{Value: "PROFILE"},
		},
		UpdateExpression: aws.String("SET #n = :name, active = :active, updated_at = :updatedAt"),
		ExpressionAttributeNames: map[string]string{
			"#n": "name", // "name" is a reserved word in DynamoDB
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":name":      &types.AttributeValueMemberS{Value: u.Name},
			":active":    &types.AttributeValueMemberBOOL{Value: u.Active},
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

func googleIDPK(googleID string) string { return "GOOGLE#" + googleID }

func toRecord(u *domain.User) record {
	rec := record{
		PK:           emailPK(u.Email),
		SK:           "PROFILE",
		GSI1PK:       idPK(u.ID),
		GSI1SK:       "PROFILE",
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		GoogleID:     u.GoogleID,
		Name:         u.Name,
		Active:       u.Active,
		CreatedAt:    u.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    u.UpdatedAt.Format(time.RFC3339),
	}
	// Only set GSI2 if user has Google ID
	if u.GoogleID != "" {
		rec.GSI2PK = googleIDPK(u.GoogleID)
		rec.GSI2SK = "PROFILE"
	}
	return rec
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
		GoogleID:     rec.GoogleID,
		Name:         rec.Name,
		Active:       rec.Active,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}, nil
}
