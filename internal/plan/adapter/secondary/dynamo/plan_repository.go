package dynamo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/shotokan/gymtracker-api/internal/plan/domain"
)

type client interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
	Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
}

// DynamoDB record structs — separate from the domain to act as anti-corruption layer.
type exerciseRec struct {
	ID     string  `dynamodbav:"id"`
	Name   string  `dynamodbav:"name"`
	Sets   int     `dynamodbav:"sets"`
	Reps   int     `dynamodbav:"reps"`
	Weight float64 `dynamodbav:"weight"`
	Notes  string  `dynamodbav:"notes"`
}

type routineRec struct {
	ID        string        `dynamodbav:"id"`
	Name      string        `dynamodbav:"name"`
	Exercises []exerciseRec `dynamodbav:"exercises"`
}

type planRec struct {
	PK        string       `dynamodbav:"PK"`
	SK        string       `dynamodbav:"SK"`
	ID        string       `dynamodbav:"id"`
	UserID    string       `dynamodbav:"user_id"`
	Name      string       `dynamodbav:"name"`
	Active    bool         `dynamodbav:"active"`
	Routines  []routineRec `dynamodbav:"routines"`
	CreatedAt string       `dynamodbav:"created_at"`
	UpdatedAt string       `dynamodbav:"updated_at"`
}

// PlanRepository implements output.PlanRepository using DynamoDB.
// Plans are stored as:
//
//	PK = USER#<userID>   SK = PLAN#<planID>
type PlanRepository struct {
	db        client
	tableName string
}

// NewPlanRepository construye el repositorio con el cliente y tabla dados.
func NewPlanRepository(db client, tableName string) *PlanRepository {
	return &PlanRepository{db: db, tableName: tableName}
}

// Save persists a new plan (PutItem, no overwrite guard needed for plans).
func (r *PlanRepository) Save(ctx context.Context, plan *domain.Plan) error {
	item, err := attributevalue.MarshalMap(toRecord(plan))
	if err != nil {
		return fmt.Errorf("marshaling plan: %w", err)
	}

	_, err = r.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("dynamo put plan: %w", err)
	}

	return nil
}

// FindByID retrieves a specific plan by userID and planID.
func (r *PlanRepository) FindByID(ctx context.Context, userID, planID string) (*domain.Plan, error) {
	out, err := r.db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: pk(userID)},
			"SK": &types.AttributeValueMemberS{Value: sk(planID)},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("dynamo get plan: %w", err)
	}
	if out.Item == nil {
		return nil, domain.ErrPlanNotFound
	}

	return unmarshal(out.Item)
}

// FindAll returns all plans for a user, optionally filtered by name substring.
func (r *PlanRepository) FindAll(ctx context.Context, userID, nameFilter string) ([]domain.Plan, error) {
	out, err := r.db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :prefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":     &types.AttributeValueMemberS{Value: pk(userID)},
			":prefix": &types.AttributeValueMemberS{Value: "PLAN#"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("dynamo query plans: %w", err)
	}

	plans := make([]domain.Plan, 0, len(out.Items))
	for _, item := range out.Items {
		p, err := unmarshal(item)
		if err != nil {
			continue
		}
		// Apply optional name filter in-memory (case-insensitive)
		if nameFilter != "" && !strings.Contains(strings.ToLower(p.Name), strings.ToLower(nameFilter)) {
			continue
		}
		plans = append(plans, *p)
	}

	return plans, nil
}

// FindActive returns the currently active plan for a user.
func (r *PlanRepository) FindActive(ctx context.Context, userID string) (*domain.Plan, error) {
	out, err := r.db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :prefix)"),
		FilterExpression:       aws.String("active = :true"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":     &types.AttributeValueMemberS{Value: pk(userID)},
			":prefix": &types.AttributeValueMemberS{Value: "PLAN#"},
			":true":   &types.AttributeValueMemberBOOL{Value: true},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, fmt.Errorf("dynamo query active plan: %w", err)
	}
	if len(out.Items) == 0 {
		return nil, domain.ErrNoActivePlan
	}

	return unmarshal(out.Items[0])
}

// Update replaces the entire plan item (PutItem).
// Since routines are embedded, this covers plan and routine mutations.
func (r *PlanRepository) Update(ctx context.Context, plan *domain.Plan) error {
	item, err := attributevalue.MarshalMap(toRecord(plan))
	if err != nil {
		return fmt.Errorf("marshaling plan: %w", err)
	}

	_, err = r.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("dynamo update plan: %w", err)
	}

	return nil
}

// Delete removes a plan item from the table.
func (r *PlanRepository) Delete(ctx context.Context, userID, planID string) error {
	_, err := r.db.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: pk(userID)},
			"SK": &types.AttributeValueMemberS{Value: sk(planID)},
		},
	})
	if err != nil {
		return fmt.Errorf("dynamo delete plan: %w", err)
	}

	return nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func pk(userID string) string { return "USER#" + userID }
func sk(planID string) string { return "PLAN#" + planID }

func toRecord(p *domain.Plan) planRec {
	routines := make([]routineRec, len(p.Routines))
	for i, r := range p.Routines {
		exercises := make([]exerciseRec, len(r.Exercises))
		for j, e := range r.Exercises {
			exercises[j] = exerciseRec{ID: e.ID, Name: e.Name, Sets: e.Sets, Reps: e.Reps}
		}
		routines[i] = routineRec{ID: r.ID, Name: r.Name, Exercises: exercises}
	}

	return planRec{
		PK:        pk(p.UserID),
		SK:        sk(p.ID),
		ID:        p.ID,
		UserID:    p.UserID,
		Name:      p.Name,
		Active:    p.Active,
		Routines:  routines,
		CreatedAt: p.CreatedAt.Format(time.RFC3339),
		UpdatedAt: p.UpdatedAt.Format(time.RFC3339),
	}
}

func unmarshal(item map[string]types.AttributeValue) (*domain.Plan, error) {
	var rec planRec
	if err := attributevalue.UnmarshalMap(item, &rec); err != nil {
		return nil, fmt.Errorf("unmarshaling plan: %w", err)
	}

	routines := make([]domain.Routine, len(rec.Routines))
	for i, r := range rec.Routines {
		exercises := make([]domain.Exercise, len(r.Exercises))
		for j, e := range r.Exercises {
			exercises[j] = domain.Exercise{ID: e.ID, Name: e.Name, Sets: e.Sets, Reps: e.Reps}
		}
		routines[i] = domain.Routine{ID: r.ID, Name: r.Name, Exercises: exercises}
	}

	createdAt, _ := time.Parse(time.RFC3339, rec.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, rec.UpdatedAt)

	return &domain.Plan{
		ID:        rec.ID,
		UserID:    rec.UserID,
		Name:      rec.Name,
		Active:    rec.Active,
		Routines:  routines,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}
